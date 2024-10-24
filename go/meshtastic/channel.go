package meshtastic

import (
	"fmt"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
)

type channel struct {
	id      int32
	name    string
	passkey []byte
}

func NewChannel(unit *meshtastic.Channel) channel {
	if unit == nil {
		return channel{}
	}
	return channel{
		id:      unit.Index,
		name:    unit.GetSettings().Name,
		passkey: unit.GetSettings().Psk,
	}
}

func (c channel) String() string {
	return fmt.Sprintf("[%d] %s", c.id, c.name)
}
