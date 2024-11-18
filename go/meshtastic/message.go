package meshtastic

import (
	"fmt"
	"time"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
)

const (
	MESSAGE_TYPE_TEXT_MESSAGE = iota
	MESSAGE_TYPE_POSITION
	MESSAGE_TYPE_NEIGHBOR_INFO
	MESSAGE_TYPE_TELEMETRY_DEVICE
	MESSAGE_TYPE_TELEMETRY_ENVIRONMENT
	MESSAGE_TYPE_TELEMETRY_HEALTH
	MESSAGE_TYPE_TELEMETRY_AIR_QUALITY
	MESSAGE_TYPE_TELEMETRY_POWER
	MESSAGE_TYPE_TELEMETRY_LOCAL_STATS
	MESSAGE_TYPE_OTHER
)

type Message struct {
	fromNode      *Node
	toNode        *Node
	timeStamp     time.Time
	messageType   int
	text          string
	deviceMetrics *meshtastic.DeviceMetrics
	neighborInfo  *meshtastic.NeighborInfo
	position      *position
	snr           float32
	hopsAway      uint32
}

func (m *Message) String() string {
	direction := m.fromNode.String() + " -> " + m.toNode.String()

	var content string
	if m.messageType == MESSAGE_TYPE_TEXT_MESSAGE {
		content = m.text
	} else {
		content = "\033[1m" + m.TypeString() + " packet\033[0m"
	}

	return fmt.Sprintf("%s: %s %s", direction, content, m.radioMetricsString())
}

func (m *Message) TypeString() string {
	switch m.messageType {
	case MESSAGE_TYPE_TEXT_MESSAGE:
		return "text message"
	case MESSAGE_TYPE_POSITION:
		return "position"
	case MESSAGE_TYPE_NEIGHBOR_INFO:
		return "neighbor info"
	case MESSAGE_TYPE_TELEMETRY_DEVICE:
		return "device telemetry"
	case MESSAGE_TYPE_TELEMETRY_ENVIRONMENT:
		return "environment telemetry"
	case MESSAGE_TYPE_TELEMETRY_HEALTH:
		return "health telemetry"
	case MESSAGE_TYPE_TELEMETRY_AIR_QUALITY:
		return "air quality telemetry"
	case MESSAGE_TYPE_TELEMETRY_POWER:
		return "power telemetry"
	case MESSAGE_TYPE_TELEMETRY_LOCAL_STATS:
		return "local stats telemetry"
	default:
		return "other"
	}
}

func (m *Message) radioMetricsString() string {
	if m.fromNode.connected {
		return ""
	}

	snr := ""
	if m.snr != 0 {
		snr = fmt.Sprintf("SNR %.2f, ", m.snr)
	}
	return fmt.Sprintf(
		"\033[90m(%s%d %s away)\033[0m",
		snr,
		m.hopsAway,
		pluralize("hop", int(m.hopsAway)),
	)
}
