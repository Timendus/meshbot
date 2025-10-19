package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/timendus/meshbot/meshbot"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	bot := meshbot.NewChatbot()
	err := bot.ReloadPlugins()
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.Replace(input, "\r", "", -1)
		input = strings.Replace(input, "\n", "", -1)

		bot.HandleMessage(chatMessage{
			MessageType: meshbot.TEXT_MESSAGE,
			Text:        input,
			FromNode:    chatUser{NodeID: 34875},
			ToNode:      chatUser{NodeID: 23857},
		})
	}
}

type chatMessage struct {
	MessageType string
	Text        string
	FromNode    chatUser
	ToNode      chatUser
}

func (m chatMessage) GetText() string {
	return m.Text
}

func (m chatMessage) IsPrivateMessage() bool {
	return true
}

func (m chatMessage) GetType() string {
	return meshbot.TEXT_MESSAGE
}

func (m chatMessage) GetChannelName() string {
	return ""
}

func (m chatMessage) GetSenderNode() meshbot.ChatUser {
	return m.FromNode
}

func (m chatMessage) GetReceiverNode() meshbot.ChatUser {
	return m.ToNode
}

func (m chatMessage) FindNode(needle string) meshbot.ChatUser {
	if needle == m.FromNode.GetShortName() {
		return m.FromNode
	}
	if needle == m.ToNode.GetShortName() {
		return m.ToNode
	}
	return nil
}

func (m chatMessage) String() string {
	return m.Text
}

func (m chatMessage) Reply(message string, timeout ...time.Duration) chan bool {
	fmt.Println(message)
	ch := make(chan bool, 1)
	go func() {
		time.Sleep(2 * time.Second)
		ch <- true
	}()
	return ch
}

type chatUser struct {
	NodeID int
}

func (m chatUser) GetId() int {
	return m.NodeID
}

func (m chatUser) GetIDExpression() string {
	return fmt.Sprintf("!%08x", m.NodeID)
}

func (m chatUser) GetShortName() string {
	return m.GetIDExpression()[5:]
}

func (m chatUser) GetLongName() string {
	return fmt.Sprintf("Node %d", m.NodeID)
}

func (m chatUser) String() string {
	return fmt.Sprintf("[%s] %s", m.GetShortName(), m.GetLongName())
}

func (m chatUser) VerboseString() string {
	return fmt.Sprintf("Node %s", m.String())
}

func (m chatUser) GetPosition() [3]float32 {
	return [3]float32{0, 0, 0}
}

func (m chatUser) GetHopsAway() int {
	return 0
}

func (m chatUser) GetRSSI() float32 {
	return -50.0
}

func (m chatUser) GetSNR() float32 {
	return 5.2
}

func (m chatUser) IsSelf() bool {
	return m.NodeID == 23857
}
