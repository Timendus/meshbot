package meshtastic

import (
	"fmt"
	"time"
	"unicode/utf8"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"github.com/timendus/meshbot/meshtastic/helpers"
)

type Node struct {
	ShortName     string
	LongName      string
	Id            uint32
	HwModel       meshtastic.HardwareModel
	Role          meshtastic.Config_DeviceConfig_Role
	Snr           float32
	LastHeard     time.Time
	HopsAway      uint32
	NodeList      nodeList
	Position      []*position
	IsLicensed    bool
	DeviceMetrics []*meshtastic.DeviceMetrics
	Connected     bool
}

func NewNode(info *meshtastic.NodeInfo) *Node {
	node := Node{
		Id:            info.Num,
		HopsAway:      0,
		ShortName:     "UNKN",
		LongName:      "Unknown node",
		HwModel:       meshtastic.HardwareModel_UNSET,
		IsLicensed:    false,
		Position:      make([]*position, 0),
		DeviceMetrics: make([]*meshtastic.DeviceMetrics, 0),
	}

	node.Update(info)
	return &node
}

func (n *Node) Update(info *meshtastic.NodeInfo) {
	if info == nil || info.Num != n.Id {
		return
	}

	n.Snr = info.Snr
	n.LastHeard = time.Unix(int64(info.LastHeard), 0)

	if info.Position != nil {
		n.Position = append(n.Position, NewPosition(info.Position))
	}

	if info.DeviceMetrics != nil {
		n.DeviceMetrics = append(n.DeviceMetrics, info.DeviceMetrics)
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
	}
}

func (n *Node) String() string {
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
		n.IDExpression(),
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

func (n *Node) IDExpression() string {
	return fmt.Sprintf("!%x", n.Id)
}
