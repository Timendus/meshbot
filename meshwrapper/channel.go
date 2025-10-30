package meshwrapper

import (
	"fmt"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
)

type Channel struct {
	id      uint32
	name    string
	passkey []byte
}

func NewChannel(unit *meshtastic.Channel) Channel {
	if unit == nil || unit.Settings == nil {
		return Channel{}
	}

	name := unit.Settings.Name
	if name == "" {
		name = "Default"
	}

	passkey := unit.Settings.Psk
	if len(passkey) == 0 {
		// This comes from the Protobuf documentation, untested
		passkey = []byte{0xd4, 0xf1, 0xbb, 0x3a, 0x20, 0x29, 0x07, 0x59, 0xf0, 0xbc, 0xff, 0xab, 0xcf, 0x4e, 0x69, 0x01}
	}

	return Channel{
		id:      uint32(unit.Index),
		name:    name,
		passkey: passkey,
	}
}

func (c Channel) String() string {
	return fmt.Sprintf("[%d] %s", c.id, c.name)
}
