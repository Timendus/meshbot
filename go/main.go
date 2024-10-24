package main

// https://meshtastic.org/docs/development/device/client-api/
// https://buf.build/meshtastic/protobufs/docs/main:meshtastic#meshtastic.ToRadio

import (
	"log"
	"net"
	"time"

	"github.com/timendus/meshbot/meshtastic"
	"go.bug.st/serial"
)

func main() {
	log.Println("Starting Meshed Potatoes!")

	// Attempt to auto-detect Meshtestic device on a serial port. Otherwise,
	// connect over TCP.

	var node *meshtastic.ConnectedNode

	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}

	if len(ports) > 0 {
		log.Printf("Found %d serial ports:\n", len(ports))
		for i, port := range ports {
			log.Printf("  [%d] %s\n", i, port)
		}
		log.Println("Defaulting to port: " + ports[0])

		serialPort, err := serial.Open(ports[0], &serial.Mode{
			BaudRate: 115200,
		})
		if err != nil {
			log.Fatal(err)
		}

		node, err = meshtastic.NewConnectedNode(serialPort, "serial-log.txt")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		tcpPort, err := net.Dial("tcp", "meshtastic.local:4403")
		if err != nil {
			log.Fatal(err)
		}

		node, err = meshtastic.NewConnectedNode(tcpPort, "tcp-log.txt")
		if err != nil {
			log.Fatal(err)
		}
	}

	defer node.Close()

	for {
		time.Sleep(100 * time.Millisecond)
	}
}
