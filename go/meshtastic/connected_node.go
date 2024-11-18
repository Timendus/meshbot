package meshtastic

import (
	"io"
	"log"
	"time"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"google.golang.org/protobuf/proto"
)

type ConnectedNode struct {
	stream            io.ReadWriteCloser
	connectedCallback func(ConnectedNode)
	messageCallback   func(Message)
	FirmwareVersion   string
	Channels          []channel
	Node              *Node
}

func NewConnectedNode(stream io.ReadWriteCloser, connected func(ConnectedNode), message func(Message)) (*ConnectedNode, error) {
	// Create the new connected node
	newNode := ConnectedNode{
		stream:            stream,
		connectedCallback: connected,
		messageCallback:   message,
		Node: &Node{
			ShortName: "UNKN",
			LongName:  "Unknown node",
			id:        0,
			connected: true,
			NodeList:  NewNodeList(),
		},
	}

	// Spin up a goroutine to read messages from the device
	go newNode.ReadMessages(stream)

	// Wake the device
	if err := wakeDevice(stream); err != nil {
		return nil, err
	}

	// Tell the device that we can speak ProtoBuf
	if err := writeMessage(stream, &meshtastic.ToRadio{
		PayloadVariant: &meshtastic.ToRadio_WantConfigId{
			WantConfigId: 1,
		},
	}); err != nil {
		return nil, err
	}

	return &newNode, nil
}

func (n *ConnectedNode) Close() error {
	return n.stream.Close()
}

func (n *ConnectedNode) String() string {
	return n.Node.String()
}

func (n *ConnectedNode) SendMessage(message meshtastic.ToRadio_Packet) error {
	if err := writeMessage(n.stream, &meshtastic.ToRadio{
		PayloadVariant: &message,
	}); err != nil {
		return err
	}
	return nil
}

func (n *ConnectedNode) ReadMessages(stream io.ReadCloser) error {
	for {
		packet, err := readMessage(stream)
		if err != nil {
			log.Println("Error: " + err.Error())
			if err == io.EOF {
				log.Println("EOF probably means the device has disconnected. Stopping execution.")
				return n.Close()
			}
			continue
		}

		switch packet.PayloadVariant.(type) {
		case *meshtastic.FromRadio_ConfigCompleteId:
			n.connectedCallback(*n)
		case *meshtastic.FromRadio_MyInfo:
			n.Node.id = packet.GetMyInfo().MyNodeNum
			n.Node.NodeList.nodes[n.Node.id] = n.Node
		case *meshtastic.FromRadio_Metadata:
			n.FirmwareVersion = packet.GetMetadata().FirmwareVersion
		case *meshtastic.FromRadio_NodeInfo:
			n.parseNodeInfo(packet.GetNodeInfo())
		case *meshtastic.FromRadio_Channel:
			n.Channels = append(n.Channels, NewChannel(packet.GetChannel()))
		case *meshtastic.FromRadio_Config:
			// log.Println("Ignoring configuration packet (for now)")
		case *meshtastic.FromRadio_ModuleConfig:
			// log.Println("Ignoring module config packet (for now)")
		case *meshtastic.FromRadio_FileInfo:
			// Silently ignore file info packets
		case *meshtastic.FromRadio_Packet:
			n.parseMeshPacket(packet.GetPacket())
		default:
			log.Println("Unhandled message:" + packet.String())
		}
	}
}

func (n *ConnectedNode) parseNodeInfo(nodeInfo *meshtastic.NodeInfo) {
	// Create or update the node that this info relates to
	relevantNode, exists := n.Node.NodeList.nodes[nodeInfo.Num]
	if !exists {
		n.Node.NodeList.nodes[nodeInfo.Num] = NewNode(nodeInfo)
	} else {
		relevantNode.Update(nodeInfo)
	}
}

func (n *ConnectedNode) parseMeshPacket(meshPacket *meshtastic.MeshPacket) {
	// Ignore broken, encrypted or empty packets
	if meshPacket == nil || meshPacket.GetDecoded() == nil || meshPacket.GetDecoded().GetPayload() == nil {
		return
	}

	var hops uint32
	if meshPacket.HopStart == 0 {
		hops = 0
	} else {
		hops = meshPacket.HopStart - meshPacket.HopLimit
	}

	payload := meshPacket.GetDecoded().GetPayload()

	fromNode := n.Node.NodeList.nodes[meshPacket.From]
	fromNode.snr = meshPacket.RxSnr
	fromNode.HopsAway = hops
	toNode := n.Node.NodeList.nodes[meshPacket.To]

	message := Message{
		fromNode:    fromNode,
		toNode:      toNode,
		timeStamp:   time.Unix(int64(meshPacket.RxTime), 0),
		messageType: MESSAGE_TYPE_OTHER,
		snr:         meshPacket.RxSnr,
		hopsAway:    hops,
	}

	switch meshPacket.GetDecoded().Portnum {
	case meshtastic.PortNum_NODEINFO_APP:
		// Update our node list with this new information

		// result := meshtastic.NodeInfo{}
		result := meshtastic.NodeInfo{}
		err := proto.Unmarshal(payload, &result)
		if err != nil {
			log.Println("Error: Could not unmarshall NodeInfo mesh packet: " + err.Error())
			return
		}

		log.Println("Got Node Info:", result.String())

		// relevantNode, exists := n.Node.NodeList.nodes[nodeInfo.Num]
		// if !exists {
		// 	relevantNode = Node{
		// 		NodeList: nodeList{nodes: make(map[uint32]Node)},
		// 	}
		// }
		// n.Node.NodeList.nodes[nodeInfo.Num] = relevantNode
	case meshtastic.PortNum_TELEMETRY_APP:
		result := meshtastic.Telemetry{}
		err := proto.Unmarshal(payload, &result)
		if err != nil {
			log.Println("Error: Could not unmarshall Telemetry mesh packet: " + err.Error())
			return
		}
		switch result.Variant.(type) {
		case *meshtastic.Telemetry_DeviceMetrics:
			message.messageType = MESSAGE_TYPE_TELEMETRY_DEVICE
			message.deviceMetrics = result.GetDeviceMetrics()
		case *meshtastic.Telemetry_EnvironmentMetrics:
			message.messageType = MESSAGE_TYPE_TELEMETRY_ENVIRONMENT
		case *meshtastic.Telemetry_HealthMetrics:
			message.messageType = MESSAGE_TYPE_TELEMETRY_HEALTH
		case *meshtastic.Telemetry_AirQualityMetrics:
			message.messageType = MESSAGE_TYPE_TELEMETRY_AIR_QUALITY
		case *meshtastic.Telemetry_PowerMetrics:
			message.messageType = MESSAGE_TYPE_TELEMETRY_POWER
		case *meshtastic.Telemetry_LocalStats:
			message.messageType = MESSAGE_TYPE_TELEMETRY_LOCAL_STATS
		default:
			log.Println("Warning: Unknown telemetry variant:", result.String())
		}
	case meshtastic.PortNum_POSITION_APP:
		result := meshtastic.Position{}
		err := proto.Unmarshal(payload, &result)
		if err != nil {
			log.Println("Error: Could not unmarshall Position mesh packet: " + err.Error())
			return
		}
		message.messageType = MESSAGE_TYPE_POSITION
		message.position = NewPosition(&result)
	case meshtastic.PortNum_NEIGHBORINFO_APP:
		result := meshtastic.NeighborInfo{}
		err := proto.Unmarshal(payload, &result)
		if err != nil {
			log.Println("Error: Could not unmarshall NeighborInfo mesh packet: " + err.Error())
			return
		}
		message.messageType = MESSAGE_TYPE_NEIGHBOR_INFO
		message.neighborInfo = &result
	case meshtastic.PortNum_TEXT_MESSAGE_APP:
		message.messageType = MESSAGE_TYPE_TEXT_MESSAGE
		message.text = string(payload)
	default:
		log.Println("Warning: Unknown mesh packet:", meshPacket.String())
	}

	n.messageCallback(message)
}
