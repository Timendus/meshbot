package main

// https://meshtastic.org/docs/development/device/client-api/
// https://buf.build/meshtastic/protobufs/docs/main:meshtastic#meshtastic.ToRadio

import (
	"fmt"
	"log"
	"net"
	"time"

	m "github.com/timendus/meshbot/meshwrapper"
	"go.bug.st/serial"
)

func main() {
	log.Println("Starting Meshed Potatoes!")

	m.MessageEvents.Subscribe(m.AnyMessageEvent, message)
	m.MessageEvents.Subscribe(m.TextMessageEvent, textMessage)
	m.ConnectionEvents.Subscribe(m.ConnectedEvent, connected)
	m.ConnectionEvents.Subscribe(m.DisconnectedEvent, disconnected)

	// Attempt to auto-detect Meshtestic device on a serial port. Otherwise,
	// connect over TCP.

	var node *m.ConnectedNode

	ports, err := serial.GetPortsList()
	if err != nil {
		log.Println(err)
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

		node, err = m.NewConnectedNode(serialPort)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		tcpPort, err := net.Dial("tcp", "meshtastic.thuis:4403")
		if err != nil {
			log.Fatal(err)
		}

		node, err = m.NewConnectedNode(tcpPort)
		if err != nil {
			log.Fatal(err)
		}
	}

	defer node.Close()

	for {
		time.Sleep(100 * time.Millisecond)
	}
}

func connected(node m.ConnectedNode) {
	log.Println("Connected to a node!")
	log.Println("This is me: " + node.String())
	log.Println("Node list: \n" + node.NodeList.String())
	log.Println("Channel list:")
	for _, channel := range node.Channels {
		log.Println("   " + channel.String())
	}
}

func disconnected(node m.ConnectedNode) {
	log.Println("Disconnected from the node. Maybe some retry-logic here?")
}

func message(message m.Message) {
	fmt.Println(message.String())
}

func textMessage(message m.Message) {
	if message.ToNode.Id != m.Broadcast.Id && message.FromNode.Id == 0x56598860 {
		log.Println("Sending message and waiting...")
		delivered := <-message.ReplyBlocking("Hello, world!", 10*time.Second)
		if delivered {
			log.Println("Message delivered!")
		} else {
			log.Println("No delivery confirmation received within 10 seconds :(")
		}
	}
}
