package meshbot

import "time"

const (
	TEXT_MESSAGE = "text message"
)

type ChatMessage interface {
	GetText() string
	IsPrivateMessage() bool
	GetType() string
	GetChannelName() string
	GetSenderNode() ChatUser
	GetReceiverNode() ChatUser
	FindNode(string) ChatUser
	String() string
	Reply(string, ...time.Duration) chan bool
}

type ChatUser interface {
	GetId() int
	GetIDExpression() string
	GetShortName() string
	GetLongName() string
	String() string
	VerboseString() string
	GetPosition() [3]float32
	GetHopsAway() int
	GetRSSI() float32
	GetSNR() float32
	IsSelf() bool
}
