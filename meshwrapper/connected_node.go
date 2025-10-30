package meshwrapper

import (
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"time"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"github.com/timendus/meshbot/config"
)

type ConnectedNode struct {
	aquireStream    func() (io.ReadWriteCloser, error)
	stream          io.ReadWriteCloser
	Connected       bool
	FirmwareVersion string
	Channels        map[uint32]Channel
	Node            *Node
	NodeList        nodeList
	Acks            map[uint32]*acknowledgement
}

func NewConnectedNode(aquire func() (io.ReadWriteCloser, error)) *ConnectedNode {
	return &ConnectedNode{
		aquireStream: aquire,
		Connected:    false,
		NodeList:     NewNodeList(),
		Acks:         make(map[uint32]*acknowledgement),
		Channels:     make(map[uint32]Channel),
		Node: &Node{
			ShortName: "UNKN",
			LongName:  "Unknown node",
			Id:        0,
			Connected: true,
		},
	}
}

func (n *ConnectedNode) FindChannel(name string) (*Channel, bool) {
	for _, channel := range n.Channels {
		if channel.name == name {
			return &channel, true
		}
	}
	return nil, false
}

func (n *ConnectedNode) Connect() error {
	// Connect to the actual device
	stream, err := n.aquireStream()
	if err != nil {
		return err
	}
	n.stream = stream

	// Spin up a goroutine to read messages from the device
	go n.readMessages(n.stream)

	// Wake the device
	if err := wakeDevice(n.stream); err != nil {
		return err
	}

	// Tell the device that we can speak ProtoBuf
	if err := writeMessage(n.stream, &meshtastic.ToRadio{
		PayloadVariant: &meshtastic.ToRadio_WantConfigId{
			WantConfigId: 1,
		},
	}); err != nil {
		return err
	}
	return nil
}

func (n *ConnectedNode) Close() error {
	n.Connected = false
	ConnectionEvents.publish(DisconnectedEvent, *n)
	return n.stream.Close()
}

func (n *ConnectedNode) String() string {
	return n.Node.ColorString()
}

func (n *ConnectedNode) SendMessage(channel uint32, recipient *Node, message string, hopLimit uint32) (uint32, error) {
	id := rand.Uint32()
	err := n.SendPacket(meshtastic.ToRadio_Packet{
		Packet: &meshtastic.MeshPacket{
			Id:       id,
			Channel:  channel,
			To:       recipient.Id,
			From:     n.Node.Id,
			HopLimit: hopLimit,
			WantAck:  true,
			Priority: meshtastic.MeshPacket_Priority(meshtastic.MeshPacket_Priority_value["RELIABLE"]),
			PayloadVariant: &meshtastic.MeshPacket_Decoded{
				Decoded: &meshtastic.Data{
					Portnum: meshtastic.PortNum_TEXT_MESSAGE_APP,
					Payload: []byte(message),
				},
			},
		},
	})
	return id, err
}

func (n *ConnectedNode) SendPacket(message meshtastic.ToRadio_Packet) error {
	// Only transmit anything if the configuration allows it or the
	// configuration has this particular node id as the exception. Otherwise,
	// just silently drop the transmission.
	cfg := config.GetConfig()
	nodeAllowed := cfg.Settings.TransmitExceptionNodeId != 0 && message.Packet.To == cfg.Settings.TransmitExceptionNodeId
	if !(cfg.Settings.AllowTransmit || nodeAllowed) {
		return fmt.Errorf("not allowed to transmit by config.json")
	}

	// If message is a message in a channel, but the configuration does not
	// allow this, again just drop the transmission
	if !cfg.Settings.AllowTransmitToChannels && message.Packet.To == Broadcast.Id {
		return fmt.Errorf("not allowed to transmit in a channel by config.json")
	}

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
			channel := packet.GetChannel()
			if channel != nil && channel.Index >= 0 && channel.GetRole() != meshtastic.Channel_DISABLED {
				n.Channels[uint32(channel.Index)] = NewChannel(channel)
			}
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
		n.NodeList.nodes[nodeInfo.Num] = NewNode(n, nodeInfo)
	} else {
		relevantNode.ingestNodeInfo(n, nodeInfo)
	}
}

func (n *ConnectedNode) parseMeshPacket(meshPacket *meshtastic.MeshPacket) {
	// Ignore broken, encrypted or empty packets
	if meshPacket == nil || meshPacket.GetDecoded() == nil || meshPacket.GetDecoded().GetPayload() == nil {
		return
	}

	fromNode, ok := n.NodeList.nodes[meshPacket.From]
	if !ok {
		fromNode = NewNode(n, &meshtastic.NodeInfo{
			Num: meshPacket.From,
		})
		n.NodeList.nodes[meshPacket.From] = fromNode
	}

	toNode, ok := n.NodeList.nodes[meshPacket.To]
	if !ok {
		toNode = NewNode(n, &meshtastic.NodeInfo{
			Num: meshPacket.To,
		})
		n.NodeList.nodes[meshPacket.To] = toNode
	}

	channel, ok := n.Channels[meshPacket.Channel]
	if !ok {
		channel = Channel{
			id:   meshPacket.Channel,
			name: "Unknown",
		}
		n.Channels[meshPacket.Channel] = channel
	}

	message := IncomingMessage{
		FromNode:      fromNode,
		ToNode:        toNode,
		ReceivingNode: n,
		Channel:       &channel,
		Timestamp:     time.Unix(int64(meshPacket.RxTime), 0),
		MessageType:   MESSAGE_TYPE_OTHER,
		Snr:           meshPacket.RxSnr,
	}
	message.ingestMeshPacket(n, meshPacket)
	fromNode.receiveMessage(n, message)

	IncomingMessageEvents.publish(IncomingMessageEvent, message)
}
