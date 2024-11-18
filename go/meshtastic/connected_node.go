package meshtastic

import (
	"fmt"
	"io"
	"log"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"google.golang.org/protobuf/proto"
)

type ConnectedNode struct {
	stream            io.ReadWriteCloser
	connectedCallback func(ConnectedNode)
	messageCallback   func(Message)
	FirmwareVersion   string
	Channels          []channel
	Node              Node
}

func NewConnectedNode(stream io.ReadWriteCloser, connected func(ConnectedNode), message func(Message)) (*ConnectedNode, error) {
	// Create the new connected node
	newNode := ConnectedNode{
		stream:            stream,
		connectedCallback: connected,
		messageCallback:   message,
		Node: Node{
			ShortName: "UNKN",
			LongName:  "Unknown node",
			id:        0,
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
	color := "92"
	return n.Node.String(&color)
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
	// Does this pertain to the connected node?
	if nodeInfo.User != nil && nodeInfo.User.Id == n.Node.IDExpression() {
		n.Node.ShortName = nodeInfo.User.ShortName
		n.Node.LongName = nodeInfo.User.LongName
		return
	}

	// Otherwise, create or update a neighbouring node
	relevantNode, exists := n.Node.NodeList.nodes[nodeInfo.Num]
	if !exists {
		n.Node.NodeList.nodes[nodeInfo.Num] = *NewNode(nodeInfo)
	} else {
		relevantNode.Update(nodeInfo)
	}
}

func (n *ConnectedNode) parseMeshPacket(meshPacket *meshtastic.MeshPacket) {
	// Ignore broken, encrypted or empty packets
	if meshPacket == nil || meshPacket.GetDecoded() == nil || meshPacket.GetDecoded().GetPayload() == nil {
		return
	}

	payload := meshPacket.GetDecoded().GetPayload()

	directionString := ""
	fromNode, ok := n.Node.NodeList.nodes[meshPacket.From]
	if ok {
		directionString += fromNode.String()
	} else {
		directionString += "Unknown node (" + fmt.Sprintf("!%x", meshPacket.From) + ")"
	}
	toNode, ok := n.Node.NodeList.nodes[meshPacket.To]
	if ok {
		directionString += " -> " + toNode.String()
	} else {
		directionString += " -> Unknown node (" + fmt.Sprintf("!%x", meshPacket.To) + ")"
	}

	switch meshPacket.GetDecoded().Portnum {
	case meshtastic.PortNum_NODEINFO_APP:
		// Update our node list with this new information
		// NOTE: is this needed? Or does the node also send a NodeInfo packet..?
		// Looks like we will have to do this ourselves

		// We need to decode this crap somehow..?
		// log.Println("Node info: " + string(payload))

		// result := meshtastic.NodeInfo{}
		result := meshtastic.NodeInfo{}
		err := proto.Unmarshal(payload, &result)
		if err != nil {
			log.Println("Error unmarshalling NodeInfo mesh packet: " + err.Error())
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
		// result := meshtastic.DeviceMetrics{}
		err := proto.Unmarshal(payload, &result)
		if err != nil {
			log.Println("Error unmarshalling Telemetry mesh packet: " + err.Error())
		}
		switch result.Variant.(type) {
		case *meshtastic.Telemetry_DeviceMetrics:
			log.Println(directionString, "Device metrics:", result.String())
		case *meshtastic.Telemetry_EnvironmentMetrics:
			log.Println(directionString, "Enviroment metrics:", result.String())
		case *meshtastic.Telemetry_HealthMetrics:
			log.Println(directionString, "Health metrics:", result.String())
		case *meshtastic.Telemetry_AirQualityMetrics:
			log.Println(directionString, "Air quality: metrics", result.String())
		case *meshtastic.Telemetry_PowerMetrics:
			log.Println(directionString, "Power metrics:", result.String())
		case *meshtastic.Telemetry_LocalStats:
			log.Println(directionString, "Local stats:", result.String())
		default:
			log.Println(directionString, "Unknown telemetry variant:", result.String())
		}
	case meshtastic.PortNum_POSITION_APP:
		result := meshtastic.Position{}
		err := proto.Unmarshal(payload, &result)
		if err != nil {
			log.Println("Error unmarshalling Position mesh packet: " + err.Error())
		}
		log.Println(directionString, "Position: ", result.String())
	case meshtastic.PortNum_NEIGHBORINFO_APP:
		result := meshtastic.NeighborInfo{}
		err := proto.Unmarshal(payload, &result)
		if err != nil {
			log.Println("Error unmarshalling NeighborInfo mesh packet: " + err.Error())
		}
		log.Println(directionString, "NeighborInfo:", result.String())
	case meshtastic.PortNum_TEXT_MESSAGE_APP:
		log.Println(directionString, string(payload))
	default:
		log.Println(directionString, "Unhandled mesh packet:"+meshPacket.String())
	}
}
