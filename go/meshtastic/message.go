package meshtastic

import (
	"fmt"
	"math/rand/v2"
	"time"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"github.com/timendus/meshbot/meshtastic/helpers"
)

const (
	MESSAGE_TYPE_TEXT_MESSAGE          = "text message"
	MESSAGE_TYPE_NODE_INFO             = "node info"
	MESSAGE_TYPE_POSITION              = "position"
	MESSAGE_TYPE_NEIGHBOR_INFO         = "neighbor info"
	MESSAGE_TYPE_ROUTING               = "routing"
	MESSAGE_TYPE_TRACEROUTE            = "traceroute"
	MESSAGE_TYPE_TELEMETRY_DEVICE      = "device telemetry"
	MESSAGE_TYPE_TELEMETRY_ENVIRONMENT = "environment telemetry"
	MESSAGE_TYPE_TELEMETRY_HEALTH      = "health telemetry"
	MESSAGE_TYPE_TELEMETRY_AIR_QUALITY = "air quality telemetry"
	MESSAGE_TYPE_TELEMETRY_POWER       = "power telemetry"
	MESSAGE_TYPE_TELEMETRY_LOCAL_STATS = "local stats telemetry"
	MESSAGE_TYPE_OTHER                 = "other"
)

type Message struct {
	FromNode      *Node
	ToNode        *Node
	ReceivingNode *ConnectedNode

	Timestamp time.Time
	Snr       float32
	HopsAway  uint32

	MessageType        string
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

func (m *Message) Reply(message string) uint32 {
	id := rand.Uint32()
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
	return id
}

func (m *Message) ReplyBlocking(message string) chan bool {
	ch := make(chan bool)
	id := m.Reply(message)
	m.ReceivingNode.Acks[id] = ch
	return ch
}

func (m *Message) String() string {
	direction := m.FromNode.String() + " -> " + m.ToNode.String()

	if m.MessageType == MESSAGE_TYPE_NEIGHBOR_INFO {
		return fmt.Sprintf("%s: \033[1mNeighbor list:\033[0m %s %s", direction, m.radioMetricsString(), m.FromNode.Neighbors.String())
	}

	var content string
	if m.MessageType == MESSAGE_TYPE_TEXT_MESSAGE {
		content = m.Text
	} else {
		content = "\033[1m" + m.MessageType + " packet\033[0m"
	}

	return fmt.Sprintf("%s: %s %s", direction, content, m.radioMetricsString())
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
