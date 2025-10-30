package roomserver

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/timendus/meshbot/config"
	m "github.com/timendus/meshbot/meshwrapper"
)

type Room struct {
	Config   config.Room
	Messages []Message
	Users    []*User
}

type Message struct {
	Sender   *User
	Contents string
}

type User struct {
	Node            *m.Node
	Send            func(string) chan bool
	Sending         atomic.Bool
	UpdatingBacklog sync.Mutex
	Backlog         []string
	Selected        *Room
}

var Rooms []Room
var Users map[*m.Node]*User

func Init(cfg config.Config) {
	for _, room := range cfg.Rooms {
		Rooms = append(Rooms, Room{
			Config: room,
		})
	}
	Users = make(map[*m.Node]*User)
}

func UserExists(msg m.IncomingMessage) bool {
	_, ok := Users[msg.FromNode]
	return ok
}

func GetUser(msg m.IncomingMessage) *User {
	if user, ok := Users[msg.FromNode]; ok {
		return user
	}
	user := &User{
		Node: msg.FromNode,
		Send: func(m string) chan bool { return msg.ReplyReliably(m) },
	}
	Users[msg.FromNode] = user
	return user
}

func RoomList(user *User) string {
	text := ""
	for i, room := range Rooms {
		public := " ✅ "
		if room.Config.Password != "" {
			public = " 🔐 "
		}
		joined := ""
		if room.present(user) {
			joined = " (joined)"
		}
		selected := ""
		if user.Selected == &Rooms[i] {
			selected = " (selected)"
		}
		text += public + room.Config.Name + joined + selected + "\n"
	}
	return text
}

func Join(user *User, roomName string, password string) error {
	roomName = strings.ToLower(roomName)
	for i, room := range Rooms {
		if roomName == strings.ToLower(room.Config.Name) {
			if room.Config.Password != "" && room.Config.Password != password {
				return fmt.Errorf("Invalid password for %s", room.Config.Name)
			}
			if room.present(user) {
				return fmt.Errorf("You are already in room %s", room.Config.Name)
			}
			Rooms[i].Users = append(Rooms[i].Users, user)
			user.Selected = &Rooms[i]
			return nil
		}
	}
	return fmt.Errorf("Can't find that room!")
}

func Leave(user *User, roomName string) error {
	roomName = strings.ToLower(roomName)
	for i, room := range Rooms {
		if roomName == strings.ToLower(room.Config.Name) {
			for j, u := range room.Users {
				if u.Node.Id == user.Node.Id {
					Rooms[i].Users = append(Rooms[i].Users[:j], Rooms[i].Users[j+1:]...)
					user.autoSelectRoom()
					return nil
				}
			}
			return fmt.Errorf("Looks like you were not in room %s", roomName)
		}
	}
	return fmt.Errorf("Can't find that room!")
}

func Select(user *User, roomName string) error {
	roomName = strings.ToLower(roomName)
	for i, room := range Rooms {
		if roomName == strings.ToLower(room.Config.Name) {
			return user.selectRoom(&Rooms[i])
		}
	}
	return fmt.Errorf("Can't find that room!")
}

func Send(user *User, message string) error {
	helpText := `<BREAK-MESSAGE>
You receive messages from rooms you have joined.
You send messages to the room you have selected.
Send /rooms to see available rooms.
Send /help to see all commands`

	rooms := user.rooms()
	if len(rooms) == 0 {
		firstPublicRoom := ""
		for _, room := range Rooms {
			if room.Config.Password == "" {
				firstPublicRoom = room.Config.Name
			}
		}
		if firstPublicRoom == "" {
			return fmt.Errorf("You're not in any rooms. /join a room.%s", helpText)
		}
		err := Join(user, firstPublicRoom, "")
		if err != nil {
			return fmt.Errorf("You're not in any rooms. /join a room.%s", helpText)
		}
		return fmt.Errorf("You were not in any rooms. I took the liberty of putting you in room %s.\n\n🔴 Note: All messages you send to me from now on will be broadcast to room %s! 🔴%s", firstPublicRoom, firstPublicRoom, helpText)
	}

	if user.Selected == nil {
		return fmt.Errorf("You have not selected a room to send to. Please /select a room.%s", helpText)
	}
	user.Selected.send(Message{Sender: user, Contents: message})
	return nil
}

func (u *User) SendBacklog() {
	// Check if another attempt to send the backlog is already running
	if !u.Sending.CompareAndSwap(false, true) {
		// Another Goroutine is already running this
		return
	}
	defer u.Sending.Store(false)

	// Also, we're mutating the backlog, keep anyone else from messing with it
	// for a while
	u.UpdatingBacklog.Lock()
	defer u.UpdatingBacklog.Unlock()

	// Do we have a backlog to send to this user?
	successful := 0
	for _, msg := range u.Backlog {
		ok := <-u.Send(msg) // With retries, this can take a couple of minutes
		if !ok {
			// It seems we're not getting through, try again later
			break
		}
		// We can remove message from backlog
		successful++
	}
	u.Backlog = u.Backlog[successful:]
}

func (room *Room) send(msg Message) {
	room.Messages = append(room.Messages, msg)
	for _, user := range room.Users {
		go func() {
			var text string
			if len(user.rooms()) > 1 {
				text = "[" + msg.Sender.Node.ShortName + " in " + room.Config.Name + "] " + msg.Contents
			} else {
				text = "[" + msg.Sender.Node.ShortName + "] " + msg.Contents
			}

			// Safely mutate backlog and send new message
			user.UpdatingBacklog.Lock()
			user.Backlog = append(user.Backlog, text)
			user.UpdatingBacklog.Unlock()
			user.SendBacklog()
		}()
	}
}

func (room *Room) present(user *User) bool {
	for _, u := range room.Users {
		if u == user {
			return true
		}
	}
	return false
}

func (user *User) rooms() []Room {
	rooms := make([]Room, 0)
	for _, room := range Rooms {
		if room.present(user) {
			rooms = append(rooms, room)
		}
	}
	return rooms
}

func (user *User) selectRoom(room *Room) error {
	if !room.present(user) {
		return fmt.Errorf("You can't select a room you haven't joined")
	}

	if user.Selected == room {
		return fmt.Errorf("Room %s was already selected", room.Config.Name)
	}

	user.Selected = room
	return nil
}

func (user *User) autoSelectRoom() {
	rooms := user.rooms()

	// If no rooms; no selection
	if len(rooms) == 0 {
		user.Selected = nil
		return
	}

	// If selected is a valid room, keep it
	if user.Selected != nil {
		for i := range rooms {
			if user.Selected == &rooms[i] {
				return
			}
		}
	}

	// Otherwise, select the first room
	user.Selected = &rooms[0]
}
