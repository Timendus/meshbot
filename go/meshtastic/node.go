package meshtastic

import (
	"fmt"
	"time"
	"unicode/utf8"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
)

type Node struct {
	ShortName     string
	LongName      string
	id            uint32
	macAddr       []byte
	HwModel       meshtastic.HardwareModel
	Role          meshtastic.Config_DeviceConfig_Role
	snr           float32
	LastHeard     time.Time
	HopsAway      uint32
	NodeList      nodeList
	Position      *position
	IsLicensed    bool
	deviceMetrics *deviceMetrics
}

func NewNode(info *meshtastic.NodeInfo) *Node {
	node := Node{
		id:         info.Num,
		HopsAway:   0,
		ShortName:  "UNKN",
		LongName:   "Unknown node",
		HwModel:    meshtastic.HardwareModel_UNSET,
		IsLicensed: false,
	}

	node.Update(info)
	return &node
}

func (n *Node) Update(info *meshtastic.NodeInfo) {
	if info == nil || info.Num != n.id {
		return
	}

	n.snr = info.Snr
	n.LastHeard = time.Unix(int64(info.LastHeard), 0)
	n.Position = NewPosition(info.Position)
	n.deviceMetrics = NewDeviceMetrics(info.DeviceMetrics)

	if info.HopsAway != nil {
		n.HopsAway = *info.HopsAway
	}

	if info.User != nil {
		n.ShortName = info.User.ShortName
		n.LongName = info.User.LongName
		n.macAddr = info.User.Macaddr
		n.HwModel = info.User.HwModel
		n.Role = info.User.Role
		n.IsLicensed = info.User.IsLicensed
	}
}

func (n *Node) String(color ...*string) string {
	var col string
	if len(color) > 0 && color[0] != nil {
		col = *color[0]
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
	if n.snr != 0 {
		snr = fmt.Sprintf(", SNR %.2f", n.snr)
	}

	hopsAway := ""
	if n.HopsAway > 0 {
		hopsAway = fmt.Sprintf(", %d %s away", n.HopsAway, pluralize("hop", int(n.HopsAway)))
	}

	return fmt.Sprintf(
		"%s \033[90m(%s, %s, last heard %s ago%s%s)\033[0m",
		n.String(),
		hardware,
		role,
		timeAgo(n.LastHeard),
		snr,
		hopsAway,
	)
}

func (n *Node) IDExpression() string {
	return fmt.Sprintf("!%x", n.id)
}
