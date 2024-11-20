package main

// https://meshtastic.org/docs/development/device/client-api/
// https://buf.build/meshtastic/protobufs/docs/main:meshtastic#meshtastic.ToRadio

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/timendus/meshbot/meshtastic"
	"go.bug.st/serial"
)

func main() {
	log.Println("Starting Meshed Potatoes!")

	meshtastic.MessageEvents.Subscribe("any", message)
	meshtastic.MessageEvents.Subscribe("text message", textMessage)
	meshtastic.NodeEvents.Subscribe("connected", connected)
	meshtastic.NodeEvents.Subscribe("disconnected", disconnected)

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

		node, err = meshtastic.NewConnectedNode(serialPort)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		tcpPort, err := net.Dial("tcp", "meshtastic.thuis:4403")
		if err != nil {
			log.Fatal(err)
		}

		node, err = meshtastic.NewConnectedNode(tcpPort)
		if err != nil {
			log.Fatal(err)
		}
	}

	defer node.Close()

	for {
		time.Sleep(100 * time.Millisecond)
	}
}

func connected(node meshtastic.ConnectedNode) {
	log.Println("Connected to a node!")
	log.Println("This is me: " + node.String())
	log.Println("Node list: \n" + node.NodeList.String())
	log.Println("Channel list:")
	for _, channel := range node.Channels {
		log.Println("   " + channel.String())
	}
}

func disconnected(node meshtastic.ConnectedNode) {
	log.Println("Disconnected from the node. Maybe some retry-logic here?")
}

func message(message meshtastic.Message) {
	fmt.Println(message.String())
}

func textMessage(message meshtastic.Message) {
	if message.ToNode.Id != meshtastic.Broadcast.Id && message.FromNode.Id == 0x56598860 {
		log.Println("Sending message and waiting...")
		<-message.ReplyBlocking("Hello, world!")
		log.Println("Process unblocked!")
	}
}
