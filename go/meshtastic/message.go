package meshtastic

import (
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"github.com/timendus/meshbot/meshtastic/helpers"
)

const (
	MESSAGE_TYPE_TEXT_MESSAGE = iota
	MESSAGE_TYPE_NODE_INFO
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
	FromNode      *Node
	ToNode        *Node
	ReceivingNode *ConnectedNode

	Timestamp time.Time
	Snr       float32
	HopsAway  uint32

	MessageType        int
	Text               string
	DeviceMetrics      *meshtastic.DeviceMetrics
	EnvironmentMetrics *meshtastic.EnvironmentMetrics
	HealthMetrics      *meshtastic.HealthMetrics
	AirQualityMetrics  *meshtastic.AirQualityMetrics
	PowerMetrics       *meshtastic.PowerMetrics
	LocalStats         *meshtastic.LocalStats
	NeighborInfo       *meshtastic.NeighborInfo
	Position           *position
}

func (m *Message) Reply(message string) {
	id := rand.Uint32()
	log.Println("Sending message with ID", id)
	m.ReceivingNode.SendMessage(meshtastic.ToRadio_Packet{
		Packet: &meshtastic.MeshPacket{
			Id:       id,
			To:       m.FromNode.Id,
			From:     m.ToNode.Id,
			HopLimit: 3,
			WantAck:  true,
			Priority: meshtastic.MeshPacket_Priority(meshtastic.MeshPacket_Priority_value["RELIABLE"]),

			// PkiEncrypted: true,
			// PublicKey:    []byte{1, 2, 3},
			// Channel: 0,

			PayloadVariant: &meshtastic.MeshPacket_Decoded{
				Decoded: &meshtastic.Data{
					Portnum: meshtastic.PortNum_TEXT_MESSAGE_APP,
					Payload: []byte(message),
				},
			},
		},
	})
}

func (m *Message) String() string {
	direction := m.FromNode.String() + " -> " + m.ToNode.String()

	var content string
	if m.MessageType == MESSAGE_TYPE_TEXT_MESSAGE {
		content = m.Text
	} else {
		content = "\033[1m" + m.typeString() + " packet\033[0m"
	}

	return fmt.Sprintf("%s: %s %s", direction, content, m.radioMetricsString())
}

func (m *Message) typeString() string {
	switch m.MessageType {
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
	if m.FromNode.Connected {
		return ""
	}

	snr := ""
	if m.Snr != 0 {
		snr = fmt.Sprintf("SNR %.2f, ", m.Snr)
	}
	return fmt.Sprintf(
		"\033[90m(%s%d %s away)\033[0m",
		snr,
		m.HopsAway,
		helpers.Pluralize("hop", int(m.HopsAway)),
	)
}
