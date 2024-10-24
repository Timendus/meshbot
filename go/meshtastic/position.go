package meshtastic

import (
	"math"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
)

type position struct {
	latitude  float32
	longitude float32
	altitude  int32
}

func NewPosition(pos *meshtastic.Position) *position {
	if pos == nil {
		return nil
	}
	var latI float64 = 0
	var lonI float64 = 0
	var alt int32 = 0
	if pos.LatitudeI != nil {
		latI = float64(*pos.LatitudeI)
	}
	if pos.LongitudeI != nil {
		lonI = float64(*pos.LongitudeI)
	}
	if pos.Altitude != nil {
		alt = *pos.Altitude
	}
	if latI == 0 && lonI == 0 && alt == 0 {
		return nil
	}
	return &position{
		latitude:  float32(latI / math.Pow(10, 7)),
		longitude: float32(lonI / math.Pow(10, 7)),
		altitude:  alt,
	}
}
