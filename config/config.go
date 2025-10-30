package config

import (
	"encoding/json"
	"io"
	"os"
)

type ConnectionType int

const (
	UNKNOWN = iota
	SERIAL_CONNECTION
	TCP_CONNECTION
)

type Config struct {
	Connections   []Connection   `json:"connections"`
	Settings      Settings       `json:"settings"`
	Rooms         []Room         `json:"rooms"`
	Announcements []Announcement `json:"announcements"`
}

type Connection struct {
	ConnectionType ConnectionType
	Name           string `json:"name"`
	Hostname       string `json:"hostname"`
	Port           int    `json:"port"`
	SerialDevice   string `json:"device"`
}

type Settings struct {
	AllowTCP                bool   `json:"allow_tcp"`
	AllowSerial             bool   `json:"allow_serial"`
	AllowTransmit           bool   `json:"allow_transmit"`
	TransmitExceptionNodeId uint32 `json:"transmit_exception_node_id"`
	AllowTransmitToChannels bool   `json:"allow_transmit_to_channels"`
}

type Room struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Announcement struct {
	Message      string `json:"message"`
	Channel      string `json:"channel"`
	DelayMinutes int    `json:"delayMinutes"`
	MaxHops      int    `json:"maxHops"`
}

var config Config

func InitConfig() error {
	configFile, err := os.Open("config.json")
	if err != nil {
		return err
	}
	configBytes, _ := io.ReadAll(configFile)
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return err
	}
	for i, connection := range config.Connections {
		if connection.Port == 0 {
			config.Connections[i].Port = 4403
		}
		if connection.Hostname != "" {
			config.Connections[i].ConnectionType = TCP_CONNECTION
		}
		if connection.SerialDevice != "" {
			config.Connections[i].ConnectionType = SERIAL_CONNECTION
		}
	}
	for i, announcement := range config.Announcements {
		if announcement.DelayMinutes == 0 {
			config.Announcements[i].DelayMinutes = 1440 // Default to once per day
		}
	}
	return nil
}

func GetConfig() Config {
	return config
}
