package main

// https://meshtastic.org/docs/development/device/client-api/
// https://buf.build/meshtastic/protobufs/docs/main:meshtastic#meshtastic.ToRadio

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/timendus/meshbot/meshbot"
	m "github.com/timendus/meshbot/meshwrapper"
	"go.bug.st/serial"
)

var bot *meshbot.Chatbot

func main() {
	log.Println("Starting Meshed Potatoes!")

	m.MessageEvents.Subscribe(m.AnyMessageEvent, message)
	m.ConnectionEvents.Subscribe(m.ConnectedEvent, connected)
	m.ConnectionEvents.Subscribe(m.DisconnectedEvent, disconnected)

	// Attempt to auto-detect Meshtestic device on a serial port. Otherwise,
	// connect over TCP.

	var node *m.ConnectedNode

	ports, err := serial.GetPortsList()
	if err != nil {
		log.Println(err)
	}

	if err == nil && len(ports) > 0 {
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

	// Launch the chat bot

	bot = meshbot.NewChatbot()
	err = bot.ReloadPlugins()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(bot.String())

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
	if bot != nil {
		bot.HandleMessage(message)
	}
}
