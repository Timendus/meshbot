package meshtastic

import (
	"io"
	"log"
	"time"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
)

type ConnectedNode struct {
	stream          io.ReadWriteCloser
	firmwareVersion string
	channels        []channel
	node            node
}

func NewConnectedNode(stream io.ReadWriteCloser) (*ConnectedNode, error) {
	// Create the new connected node
	newNode := ConnectedNode{
		stream: stream,
		node: node{
			shortName: "UNKN",
			longName:  "Unknown node",
			id:        0,
			nodeList:  NewNodeList(),
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
	return n.node.String(&color)
}

func (n *ConnectedNode) SendMessage(message meshtastic.ToRadio_Packet) error {
	if err := writeMessage(n.stream, &meshtastic.ToRadio{
		PayloadVariant: &message,
	}); err != nil {
		return err
	}
	return nil
}

func (n *ConnectedNode) ReadMessages(stream io.ReadWriteCloser) error {
	for {
		packet, err := readMessage(stream)
		if err != nil {
			log.Println("Error: " + err.Error())
			continue
		}

		switch packet.PayloadVariant.(type) {
		case *meshtastic.FromRadio_ConfigCompleteId:
			log.Println("Loaded all device information")
			log.Println("This is me: " + n.String())
			log.Println("Node list: \n" + n.node.nodeList.String())
			log.Println("Channel list:")
			for _, channel := range n.channels {
				log.Println("   " + channel.String())
			}
		case *meshtastic.FromRadio_MyInfo:
			n.node.id = packet.GetMyInfo().MyNodeNum
		case *meshtastic.FromRadio_Metadata:
			n.firmwareVersion = packet.GetMetadata().FirmwareVersion
		case *meshtastic.FromRadio_NodeInfo:
			nodeInfo := packet.GetNodeInfo()

			if nodeInfo.User.Id == n.node.IDExpression() {
				n.node.shortName = nodeInfo.User.LongName
				n.node.longName = nodeInfo.User.ShortName
				break
			}

			var hopsAway uint32 = 0
			if nodeInfo.HopsAway != nil {
				hopsAway = *nodeInfo.HopsAway
			}

			relevantNode, exists := n.node.nodeList.nodes[nodeInfo.Num]
			if !exists {
				relevantNode = node{
					nodeList: nodeList{nodes: make(map[uint32]node)},
				}
			}

			relevantNode.id = nodeInfo.Num
			relevantNode.shortName = nodeInfo.User.ShortName
			relevantNode.longName = nodeInfo.User.LongName
			relevantNode.macAddr = nodeInfo.User.Macaddr
			relevantNode.hwModel = nodeInfo.User.HwModel
			relevantNode.role = nodeInfo.User.Role
			relevantNode.snr = nodeInfo.Snr
			relevantNode.lastHeard = time.Unix(int64(nodeInfo.LastHeard), 0)
			relevantNode.hopsAway = hopsAway
			relevantNode.isLicensed = nodeInfo.User.IsLicensed
			relevantNode.position = NewPosition(nodeInfo.Position)
			relevantNode.deviceMetrics = NewDeviceMetrics(nodeInfo.DeviceMetrics)
			n.node.nodeList.nodes[nodeInfo.Num] = relevantNode
		case *meshtastic.FromRadio_Channel:
			n.channels = append(n.channels, NewChannel(packet.GetChannel()))
		case *meshtastic.FromRadio_Config:
			// log.Println("Ignoring configuration packet (for now)")
		case *meshtastic.FromRadio_ModuleConfig:
			// log.Println("Ignoring module config packet (for now)")
		case *meshtastic.FromRadio_FileInfo:
			// Silently ignore file info packets
		case *meshtastic.FromRadio_Packet:
			log.Println("Got packet: " + packet.String())
		default:
			log.Println("Unhandled message: " + packet.String())
		}
	}
}
