package meshwrapper

import (
	"fmt"
	"time"
	"unicode/utf8"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"github.com/timendus/meshbot/meshwrapper/helpers"
)

type Node struct {
	ShortName        string
	LongName         string
	Id               uint32
	HwModel          meshtastic.HardwareModel
	Role             meshtastic.Config_DeviceConfig_Role
	Snr              float32
	LastHeard        time.Time
	HopsAway         uint32
	IsLicensed       bool
	ReceivedMessages []*IncomingMessage
	Connected        bool
	PublicKey        []byte
	Neighbors        NeighborList
	Position         *Position
}

func NewNode(connectedNode *ConnectedNode, info *meshtastic.NodeInfo) *Node {
	node := Node{
		Id:               info.Num,
		HopsAway:         0,
		ShortName:        "UNKN",
		LongName:         "Unknown node",
		HwModel:          meshtastic.HardwareModel_UNSET,
		IsLicensed:       false,
		ReceivedMessages: make([]*IncomingMessage, 0),
		Neighbors:        make(NeighborList, 0),
	}

	node.ingestNodeInfo(connectedNode, info)
	return &node
}

// NodeInfo is basically the node list we get from the device when we first
// connect over serial (I think) as well as the list of nodes in Neighbour Info
func (n *Node) ingestNodeInfo(connectedNode *ConnectedNode, info *meshtastic.NodeInfo) {
	if info == nil || info.Num != n.Id {
		return
	}

	n.Snr = info.Snr
	n.LastHeard = time.Unix(int64(info.LastHeard), 0)

	if info.Position != nil {
		n.ReceivedMessages = append(n.ReceivedMessages, &IncomingMessage{
			FromNode:      n,
			ToNode:        &Broadcast,
			ReceivingNode: connectedNode,
			Timestamp:     time.Unix(int64(info.LastHeard), 0),
			MessageType:   MESSAGE_TYPE_POSITION,
			Position:      NewPosition(info.Position),
		})
		n.Position = NewPosition(info.Position)
	}

	if info.DeviceMetrics != nil {
		n.ReceivedMessages = append(n.ReceivedMessages, &IncomingMessage{
			FromNode:      n,
			ToNode:        &Broadcast,
			ReceivingNode: connectedNode,
			Timestamp:     time.Unix(int64(info.LastHeard), 0),
			MessageType:   MESSAGE_TYPE_TELEMETRY_DEVICE,
			DeviceMetrics: info.DeviceMetrics,
		})
	}

	if info.HopsAway != nil {
		n.HopsAway = *info.HopsAway
	}

	if info.User != nil {
		n.ShortName = info.User.ShortName
		n.LongName = info.User.LongName
		n.HwModel = info.User.HwModel
		n.Role = info.User.Role
		n.IsLicensed = info.User.IsLicensed
		n.PublicKey = info.User.PublicKey
	}
}

// If we receive a messages that came from this node, make sure we update our
// node accordingly and store the message in our list
func (n *Node) receiveMessage(connectedNode *ConnectedNode, message IncomingMessage) {
	n.ReceivedMessages = append(n.ReceivedMessages, &message)
	n.LastHeard = message.Timestamp
	n.HopsAway = message.HopsAway
	if message.HopsAway == 0 {
		// Assumption: the packet RxSnr is the signal quality of the received
		// packet, which may have hopped through other nodes. So only update
		// this node's SNR if we haven't hopped yet.
		n.Snr = message.Snr
	}

	switch message.MessageType {
	case MESSAGE_TYPE_NODE_INFO:
		n.ShortName = message.UserInfo.ShortName
		n.LongName = message.UserInfo.LongName
		n.HwModel = message.UserInfo.HwModel
		n.Role = message.UserInfo.Role
		n.IsLicensed = message.UserInfo.IsLicensed
		n.PublicKey = message.UserInfo.PublicKey
	case MESSAGE_TYPE_POSITION:
		n.Position = message.Position
	case MESSAGE_TYPE_NEIGHBOR_INFO:
		n.Neighbors = NewNeighbourList(connectedNode, message)
	}
}

func (n *Node) GetId() int {
	return int(n.Id)
}

func (n *Node) GetIDExpression() string {
	return fmt.Sprintf("!%08x", n.Id)
}

func (n *Node) GetShortName() string {
	return n.ShortName
}

func (n *Node) GetLongName() string {
	return n.LongName
}

func (n *Node) ColorString() string {
	var col string
	if n.Connected {
		col = "92"
	} else if n.Id == Broadcast.Id || n.Id == Unknown.Id {
		col = "95"
	} else if n.HopsAway == 0 {
		col = "96"
	} else {
		col = "94"
	}

	var shortName string
	if len(n.ShortName) == 4 && utf8.RuneCountInString(n.ShortName) == 1 {
		// Short name is an emoji
		shortName = fmt.Sprintf(" %s ", n.ShortName)
	} else {
		shortName = fmt.Sprintf("%-4s", n.ShortName)
	}

	return fmt.Sprintf(
		"\033[%sm[%s] %s (%s)]\033[0m",
		col,
		shortName,
		n.LongName,
		n.GetIDExpression(),
	)
}

func (n *Node) String() string {
	return fmt.Sprintf(
		"[%s] %s (%s)",
		n.ShortName,
		n.LongName,
		n.GetIDExpression(),
	)
}

func (n *Node) VerboseString() string {
	hardware := n.HwModel.String()
	role := n.Role.String()

	snr := ""
	if n.Snr != 0 {
		snr = fmt.Sprintf(", SNR %.2f", n.Snr)
	}

	hopsAway := ""
	if n.HopsAway > 0 {
		hopsAway = fmt.Sprintf(", %d %s away", n.HopsAway, helpers.Pluralize("hop", int(n.HopsAway)))
	}

	return fmt.Sprintf(
		"%s \033[90m(%s, %s, last heard %s ago%s%s)\033[0m",
		n.String(),
		hardware,
		role,
		helpers.TimeAgo(n.LastHeard),
		snr,
		hopsAway,
	)
}

func (n *Node) GetPosition() [3]float32 {
	if n.Position == nil {
		return [3]float32{0, 0, 0}
	}
	return [3]float32{
		n.Position.latitude,
		n.Position.longitude,
		float32(n.Position.altitude),
	}
}

func (n *Node) GetHopsAway() int {
	return int(n.HopsAway)
}

func (n *Node) GetRSSI() float32 {
	panic("TODO: implement")
}

func (n *Node) GetSNR() float32 {
	return n.Snr
}

func (n *Node) IsSelf() bool {
	return n.Connected
}
