package meshtastic

import (
	"fmt"
	"time"
	"unicode/utf8"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
)

type node struct {
	shortName     string
	longName      string
	id            uint32
	macAddr       []byte
	hwModel       meshtastic.HardwareModel
	role          meshtastic.Config_DeviceConfig_Role
	snr           float32
	lastHeard     time.Time
	hopsAway      uint32
	nodeList      nodeList
	position      *position
	isLicensed    bool
	deviceMetrics *deviceMetrics
}

func (n *node) String(color ...*string) string {
	var col string
	if len(color) > 0 && color[0] != nil {
		col = *color[0]
	} else if n.hopsAway == 0 {
		col = "96"
	} else {
		col = "94"
	}

	var shortName string
	if len(n.shortName) == 4 && utf8.RuneCountInString(n.shortName) == 1 {
		// Short name is an emoji
		shortName = fmt.Sprintf(" %s ", n.shortName)
	} else {
		shortName = fmt.Sprintf("%-4s", n.shortName)
	}

	return fmt.Sprintf(
		"\033[%sm[%s] %s (%s)]\033[0m",
		col,
		shortName,
		n.longName,
		n.IDExpression(),
	)
}

func (n *node) VerboseString() string {
	hardware := n.hwModel.String()
	role := n.role.String()

	snr := ""
	if n.snr != 0 {
		snr = fmt.Sprintf(", SNR %.2f", n.snr)
	}

	hopsAway := ""
	if n.hopsAway > 0 {
		hopsAway = fmt.Sprintf(", %d %s away", n.hopsAway, pluralize("hop", int(n.hopsAway)))
	}

	return fmt.Sprintf(
		"%s \033[90m(%s, %s, last heard %s ago%s%s)\033[0m",
		n.String(),
		hardware,
		role,
		timeAgo(n.lastHeard),
		snr,
		hopsAway,
	)
}

func (n *node) IDExpression() string {
	return fmt.Sprintf("!%x", n.id)
}
