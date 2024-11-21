package meshwrapper

import (
	"io"
	"log"
	"strconv"
	"time"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"github.com/timendus/meshbot/meshwrapper/helpers"
	"google.golang.org/protobuf/proto"
)

type ConnectedNode struct {
	stream          io.ReadWriteCloser
	Connected       bool
	FirmwareVersion string
	Channels        []channel
	Node            *Node
	NodeList        nodeList
	Acks            map[uint32]chan bool
}

func NewConnectedNode(stream io.ReadWriteCloser) (*ConnectedNode, error) {
	// Create the new connected node
	newNode := ConnectedNode{
		stream:    stream,
		Connected: false,
		NodeList:  NewNodeList(),
		Acks:      make(map[uint32]chan bool),
		Node: &Node{
			ShortName: "UNKN",
			LongName:  "Unknown node",
			Id:        0,
			Connected: true,
		},
	}

	// Spin up a goroutine to read messages from the device
	go newNode.readMessages(stream)

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
	n.Connected = false
	ConnectionEvents.publish(DisconnectedEvent, *n)
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

func (n *ConnectedNode) readMessages(stream io.ReadCloser) error {
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
			n.Connected = true
			ConnectionEvents.publish(ConnectedEvent, *n)
		case *meshtastic.FromRadio_MyInfo:
			n.Node.Id = packet.GetMyInfo().MyNodeNum
			n.NodeList.nodes[n.Node.Id] = n.Node
		case *meshtastic.FromRadio_Metadata:
			n.FirmwareVersion = packet.GetMetadata().FirmwareVersion
		case *meshtastic.FromRadio_NodeInfo:
			n.parseNodeInfo(packet.GetNodeInfo())
		case *meshtastic.FromRadio_Channel:
			n.Channels = append(n.Channels, NewChannel(packet.GetChannel()))
		case *meshtastic.FromRadio_Packet:
			n.parseMeshPacket(packet.GetPacket())
		case *meshtastic.FromRadio_Config:
		case *meshtastic.FromRadio_ModuleConfig:
		case *meshtastic.FromRadio_FileInfo:
		case *meshtastic.FromRadio_QueueStatus:
			// Silently ignore these packets
		default:
			log.Println("Unhandled message:" + packet.String())
		}
	}
}

func (n *ConnectedNode) parseNodeInfo(nodeInfo *meshtastic.NodeInfo) {
	// Create or update the node that this info relates to
	relevantNode, exists := n.NodeList.nodes[nodeInfo.Num]
	if !exists {
		n.NodeList.nodes[nodeInfo.Num] = NewNode(nodeInfo)
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

	toNode := n.NodeList.nodes[meshPacket.To]
	fromNode := n.NodeList.nodes[meshPacket.From]

	if fromNode == nil {
		// If the sending node is not in our node list yet, just add it.
		fromNode = NewNode(&meshtastic.NodeInfo{
			Num:       meshPacket.From,
			LastHeard: meshPacket.RxTime,
		})
		n.NodeList.nodes[meshPacket.From] = fromNode
	}

	fromNode.HopsAway = hops
	if hops == 0 {
		// Assumption: the packet RxSnr is the signal quality of the received
		// packet, which may have hopped through other nodes. So only update
		// this node's SNR if we haven't hopped yet.
		fromNode.Snr = meshPacket.RxSnr
	}

	message := Message{
		FromNode:      fromNode,
		ToNode:        toNode,
		ReceivingNode: n,
		Timestamp:     time.Unix(int64(meshPacket.RxTime), 0),
		MessageType:   MESSAGE_TYPE_OTHER,
		Snr:           meshPacket.RxSnr,
		HopsAway:      hops,
	}

	fromNode.ReceivedMessages = append(fromNode.ReceivedMessages, &message)

	switch meshPacket.GetDecoded().Portnum {
	case meshtastic.PortNum_NODEINFO_APP:
		result := meshtastic.User{}
		err := proto.Unmarshal(payload, &result)
		if err != nil {
			log.Println("Error: Could not unmarshall NodeInfo User mesh packet: " + err.Error())
			return
		}
		fromNode.ShortName = result.ShortName
		fromNode.LongName = result.LongName
		fromNode.HwModel = result.HwModel
		fromNode.Role = result.Role
		fromNode.IsLicensed = result.IsLicensed
		fromNode.PublicKey = result.PublicKey
		message.MessageType = MESSAGE_TYPE_NODE_INFO
		MessageEvents.publish(NodeInfoEvent, message)

	case meshtastic.PortNum_TELEMETRY_APP:
		result := meshtastic.Telemetry{}
		err := proto.Unmarshal(payload, &result)
		if err != nil {
			log.Println("Error: Could not unmarshall Telemetry mesh packet: " + err.Error())
			return
		}
		switch result.Variant.(type) {
		case *meshtastic.Telemetry_DeviceMetrics:
			message.MessageType = MESSAGE_TYPE_TELEMETRY_DEVICE
			message.DeviceMetrics = result.GetDeviceMetrics()
			MessageEvents.publish(DeviceTelemetryEvent, message)
		case *meshtastic.Telemetry_EnvironmentMetrics:
			message.MessageType = MESSAGE_TYPE_TELEMETRY_ENVIRONMENT
			message.EnvironmentMetrics = result.GetEnvironmentMetrics()
			MessageEvents.publish(EnvironmentTelemetryEvent, message)
		case *meshtastic.Telemetry_HealthMetrics:
			message.MessageType = MESSAGE_TYPE_TELEMETRY_HEALTH
			message.HealthMetrics = result.GetHealthMetrics()
			MessageEvents.publish(HealthTelemetryEvent, message)
		case *meshtastic.Telemetry_AirQualityMetrics:
			message.MessageType = MESSAGE_TYPE_TELEMETRY_AIR_QUALITY
			message.AirQualityMetrics = result.GetAirQualityMetrics()
			MessageEvents.publish(AirQualityTelemetryEvent, message)
		case *meshtastic.Telemetry_PowerMetrics:
			message.MessageType = MESSAGE_TYPE_TELEMETRY_POWER
			message.PowerMetrics = result.GetPowerMetrics()
			MessageEvents.publish(PowerTelemetryEvent, message)
		case *meshtastic.Telemetry_LocalStats:
			message.MessageType = MESSAGE_TYPE_TELEMETRY_LOCAL_STATS
			message.LocalStats = result.GetLocalStats()
			MessageEvents.publish(LocalStatsTelemetryEvent, message)
		default:
			log.Println("Warning: Unknown telemetry variant:", result.String())
		}
		MessageEvents.publish(TelemetryEvent, message)

	case meshtastic.PortNum_POSITION_APP:
		result := meshtastic.Position{}
		err := proto.Unmarshal(payload, &result)
		if err != nil {
			log.Println("Error: Could not unmarshall Position mesh packet: " + err.Error())
			return
		}
		message.MessageType = MESSAGE_TYPE_POSITION
		message.Position = NewPosition(&result)
		MessageEvents.publish(PositionEvent, message)

	case meshtastic.PortNum_NEIGHBORINFO_APP:
		result := meshtastic.NeighborInfo{}
		err := proto.Unmarshal(payload, &result)
		if err != nil {
			log.Println("Error: Could not unmarshall NeighborInfo mesh packet: " + err.Error())
			return
		}
		message.MessageType = MESSAGE_TYPE_NEIGHBOR_INFO
		message.NeighborInfo = &result
		helpers.Assert(result.NodeId == meshPacket.From, "I don't understand this format well enough: received "+message.String()+" but it has NodeId "+strconv.Itoa(int(result.NodeId)))
		fromNode.Neighbors = NewNeighbourList(&n.NodeList, meshPacket.RxTime, result.Neighbors)
		MessageEvents.publish(NeighborInfoEvent, message)

	case meshtastic.PortNum_TEXT_MESSAGE_APP:
		message.MessageType = MESSAGE_TYPE_TEXT_MESSAGE
		message.Text = string(payload)
		MessageEvents.publish(TextMessageEvent, message)

	case meshtastic.PortNum_ROUTING_APP:
		if meshPacket.GetDecoded() != nil {
			messageId := meshPacket.GetDecoded().RequestId
			if n.Acks[messageId] != nil {
				n.Acks[messageId] <- true
				delete(n.Acks, messageId)
			}
		}
		message.MessageType = MESSAGE_TYPE_ROUTING
		MessageEvents.publish(RoutingEvent, message)

	case meshtastic.PortNum_TRACEROUTE_APP:
		message.MessageType = MESSAGE_TYPE_TRACEROUTE
		MessageEvents.publish(TraceRouteEvent, message)

	default:
		log.Println("Warning: Unknown mesh packet:", meshPacket.String())
	}

	MessageEvents.publish(AnyMessageEvent, message)
}
