package meshtastic

import "buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"

type deviceMetrics struct {
	batteryLevel       *uint32
	voltage            *float32
	channelUtilization *float32
	airUtilizationTx   *float32
	uptime             *uint32
}

func NewDeviceMetrics(metrics *meshtastic.DeviceMetrics) *deviceMetrics {
	if metrics == nil {
		return nil
	}
	return &deviceMetrics{
		batteryLevel:       metrics.BatteryLevel,
		voltage:            metrics.Voltage,
		channelUtilization: metrics.ChannelUtilization,
		airUtilizationTx:   metrics.AirUtilTx,
		uptime:             metrics.UptimeSeconds,
	}
}
