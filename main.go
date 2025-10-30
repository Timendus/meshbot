package main

// https://meshtastic.org/docs/development/device/client-api/
// https://buf.build/meshtastic/protobufs/docs/main:meshtastic#meshtastic.ToRadio

import (
	"fmt"
	"io"
	"log"
	"maps"
	"net"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/timendus/meshbot/config"
	m "github.com/timendus/meshbot/meshwrapper"
	"github.com/timendus/meshbot/meshwrapper/helpers"
	"github.com/timendus/meshbot/roomserver"
	"github.com/timendus/meshbot/weather"
	"go.bug.st/serial"
)

func main() {
	log.Println("Starting Meshed Potatoes!")
	err := config.InitConfig()
	if err != nil {
		log.Println("Encountered issue reading config.json:")
		log.Fatal(err)
	}
	cfg := config.GetConfig()
	roomserver.Init(cfg)

	m.IncomingMessageEvents.Subscribe(m.IncomingMessageEvent, incoming)
	m.OutgoingMessageEvents.Subscribe(m.OutgoingMessageEvent, outgoing)
	m.ConnectionEvents.Subscribe(m.ConnectedEvent, connected)
	m.ConnectionEvents.Subscribe(m.DisconnectedEvent, disconnected)

	// Connect to the meshtastic devices mentioned in the configuration file
	for _, connection := range cfg.Connections {
		var node *m.ConnectedNode
		var err error

		switch connection.ConnectionType {
		case config.SERIAL_CONNECTION:
			if !cfg.Settings.AllowSerial {
				log.Fatal("Serial connection configured, but not allowed by settings")
			}
			node = m.NewConnectedNode(func() (io.ReadWriteCloser, error) {
				log.Println("Attempting to open serial connection to " + connection.SerialDevice)
				stream, err := serial.Open(connection.SerialDevice, &serial.Mode{
					BaudRate: 115200,
				})
				if err != nil {
					return nil, fmt.Errorf("Could not open serial connection to '"+connection.SerialDevice+"': ", err)
				}
				return stream, nil
			})

		case config.TCP_CONNECTION:
			if !cfg.Settings.AllowTCP {
				log.Fatal("TCP connection configured, but not allowed by settings")
			}
			node = m.NewConnectedNode(func() (io.ReadWriteCloser, error) {
				conn := connection.Hostname + ":" + strconv.Itoa(connection.Port)
				log.Println("Attempting to open TCP connection to " + conn)
				stream, err := net.Dial("tcp", conn)
				if err != nil {
					return nil, fmt.Errorf("Could not open TCP connection to '"+conn+"': ", err)
				}
				return stream, nil
			})

		default:
			log.Fatal("Invalid connection type!")
		}

		err = node.Connect()
		if err != nil {
			log.Fatal(err)
		}
		defer node.Close()
	}

	// Endless loop to keep the program from ending
	for {
		time.Sleep(100 * time.Millisecond)
	}
}

// For later use
func getSerialDevices() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return ports, err
	}

	if len(ports) > 0 {
		log.Printf("Found %d serial ports:\n", len(ports))
		for i, port := range ports {
			log.Printf("  [%d] %s\n", i, port)
		}
	}

	return ports, err
}

var announcersRunning bool

func connected(node m.ConnectedNode) {
	log.Println("Connected to " + node.String())
	// log.Println("Node list: \n" + node.NodeList.String())

	// Inform user of available channels
	log.Println("Channel list:")
	keys := slices.Collect(maps.Keys(node.Channels))
	slices.Sort(keys)
	for _, key := range keys {
		channel := node.Channels[key]
		log.Println("   " + channel.String())
	}

	// Start announcer service(s)
	if !announcersRunning {
		for _, announcement := range config.GetConfig().Announcements {
			channel, ok := node.FindChannel(announcement.Channel)
			if !ok {
				log.Printf("Announcer: Can't find channel %s\n", announcement.Channel)
				continue
			}
			go func() {
				for {
					log.Println("Announcement time!")
					m.NewOutgoingChannelMessage(announcement.Message, &node, channel, announcement.MaxHops).Send()
					time.Sleep(time.Duration(announcement.DelayMinutes) * time.Minute)
				}
			}()
		}
		announcersRunning = true
	}
}

func disconnected(node m.ConnectedNode) {
	log.Println("Disconnected from the node")
	backoff := 1 * time.Second
	for !node.Connected {
		log.Println("Attemping to reconnect to the node...")
		err := node.Connect()
		if err == nil {
			log.Println("Succesfully reconnected to the node!")
			break
		}
		log.Println("Could not connect, backing off exponentially...", err)
		time.Sleep(backoff)
		backoff *= 2
	}
}

func incoming(message m.IncomingMessage) {
	fmt.Println(message.String())

	if roomserver.UserExists(message) {
		user := roomserver.GetUser(message)
		go user.SendBacklog()
	}

	if message.MessageType != m.MESSAGE_TYPE_TEXT_MESSAGE {
		return
	}

	command := strings.ToUpper(message.Text)

	if strings.HasPrefix(command, "/PING") {
		message.Reply("🤖🏓 Pong!")
		return
	}

	if strings.HasPrefix(command, "/HELP") || strings.HasPrefix(command, "/ABOUT") {
		message.Reply(
			`🤖👋 Hello! I'm your friendly neighbourhood roomserver bot. I understand these commands:

 - /rooms
 - /join <room name> <optional password>
 - /select <room name>
 - /leave <room name>

<BREAK-MESSAGE>

Bonus features:

 - /neighbours
 - /signal <optional node>
 - /weather
 - /forecast

For details, see: github.com/Timendus/meshbot/blob/main/manual.md`)
		return
	}

	if strings.HasPrefix(command, "/SIGNAL") {
		input := strings.TrimSpace(message.Text)
		subject := message.FromNode
		ok := true
		if len(input) > len("/SIGNAL") {
			needle := input[len("/SIGNAL"):]
			subject = message.FindNode(needle)
		}

		if !ok || subject == nil {
			message.Reply("🤖🧨 I don't know who that is. Sorry!\n\nI need the short name (example: TDRP), node ID (example: !87e35ac8) or part of the long name of a node that I know.")
			return
		}

		if subject.HopsAway == 0 {
			message.Reply("🤖📶 I last heard " + subject.String() + " " + helpers.TimeAgo(subject.LastHeard) + " ago with an SNR of " +
				strconv.FormatFloat(float64(subject.GetSNR()), 'f', 2, 32))
		} else {
			message.Reply("🤖📶 " + subject.String() + " is " + strconv.Itoa(int(subject.HopsAway)) + " " + helpers.Pluralize("hop", int(subject.HopsAway)) + " away")
		}
		return
	}

	if strings.HasPrefix(command, "/NEIGHBOURS") {
		message.Reply("🤖👂 These are the nodes I've heard in the last hour:\n\n" + message.ReceivingNode.NodeList.Neighbours())
		return
	}

	if strings.HasPrefix(command, "/WEATHER") {
		var text string
		var pos [3]float32
		if message.FromNode != nil {
			pos = message.FromNode.GetPosition()
			text = "Here's the current weather at your location:"
		}
		if message.FromNode == nil || pos[0] == 0 || pos[1] == 0 {
			pos = message.ToNode.GetPosition()
			text = "I can't see your location, so I'll give you the current weather at my location:"
		}
		if pos[0] == 0 || pos[1] == 0 {
			message.Reply("🤖🧨 I'm sorry! I can't give you a weather report, because I don't know the location of either of us.")
			return
		}
		weather, err := weather.FetchWeather(weather.Position{
			Latitude:  float64(pos[0]),
			Longitude: float64(pos[1]),
		})
		if err != nil {
			message.Reply("🤖🌂 I can't get a weather report at this time.")
		} else {
			ok := <-message.Reply("🤖🌂 " + text + "\n\n" + weather)
			if !ok {
				log.Println("Could not send the full weather message :/")
			}
		}
		return
	}

	if strings.HasPrefix(command, "/FORECAST") {
		var text string
		var pos [3]float32
		if message.FromNode != nil {
			pos = message.FromNode.GetPosition()
			text = "Here's the weather forecast at your location:"
		}
		if message.FromNode == nil || pos[0] == 0 || pos[1] == 0 {
			pos = message.ToNode.GetPosition()
			text = "I can't see your location, so I'll give you the weather forecast at my location:"
		}
		if pos[0] == 0 || pos[1] == 0 {
			message.Reply("🤖🧨 I'm sorry! I can't give you a weather forecast, because I don't know the location of either of us.")
			return
		}
		forecast, err := weather.FetchForecast(weather.Position{
			Latitude:  float64(pos[0]),
			Longitude: float64(pos[1]),
		})
		if err != nil {
			message.Reply("🤖🌂 I can't get a weather forecast at this time.")
		} else {
			ok := <-message.Reply("🤖🌂 " + text + "\n\n" + forecast)
			if !ok {
				log.Println("Could not send the full weather message :/")
			}
		}
		return
	}

	// We've fallen through the generic queries, roomserver code starts here

	// Make sure we don't spam channels
	if !message.IsPrivateMessage() {
		return
	}

	// Find our user
	user := roomserver.GetUser(message)

	if strings.HasPrefix(command, "/ROOMS") {
		message.ReplyReliably("🤖💬 These are the available rooms:\n\n" + roomserver.RoomList(user))
		return
	}

	if strings.HasPrefix(command, "/JOIN") {
		params := strings.Split(strings.TrimSpace(message.Text[len("/JOIN"):]), " ")
		if len(params) == 0 {
			message.ReplyReliably("🤖🧨 You need to specify the name of a room to join")
			return
		}
		roomName := params[0]
		password := ""
		if len(params) > 1 {
			password = params[1]
		}
		err := roomserver.Join(user, roomName, password)
		if err != nil {
			message.ReplyReliably("🤖💬 " + err.Error())
			return
		}
		message.ReplyReliably("🤖💬 You joined " + roomName + ". It was also automatically selected for you to send to.")
		return
	}

	if strings.HasPrefix(command, "/LEAVE") {
		params := strings.Split(strings.TrimSpace(message.Text[len("/LEAVE"):]), " ")
		if len(params) == 0 {
			message.ReplyReliably("🤖🧨 You need to specify the name of a room to leave")
			return
		}
		roomName := params[0]
		err := roomserver.Leave(user, roomName)
		if err != nil {
			message.ReplyReliably("🤖🧨 " + err.Error())
			return
		}
		message.ReplyReliably("🤖💬 You left " + roomName)
		return
	}

	if strings.HasPrefix(command, "/SELECT") {
		params := strings.Split(strings.TrimSpace(message.Text[len("/SELECT"):]), " ")
		if len(params) == 0 {
			message.ReplyReliably("🤖🧨 You need to specify the name of a room to select")
			return
		}
		roomName := params[0]
		err := roomserver.Select(user, roomName)
		if err != nil {
			message.ReplyReliably("🤖💬 " + err.Error())
			return
		}
		message.ReplyReliably("🤖💬 You selected " + roomName)
		return
	}

	if strings.HasPrefix(command, "/") {
		message.ReplyReliably("🤖❓ I don't know that command. See /help for the things I understand!")
		return
	}

	// Handle freeform messages to a room
	msg := strings.TrimSpace(message.Text)
	if len(msg) == 0 {
		return
	}
	err := roomserver.Send(user, msg)
	if err != nil {
		message.ReplyReliably("🤖💬 " + err.Error())
		return
	}
}

func outgoing(message m.OutgoingMessage) {
	fmt.Println(message.String())
}
