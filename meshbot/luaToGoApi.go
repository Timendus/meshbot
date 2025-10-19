package meshbot

// Just pass on the Go interfaces to something that Lua understands

import (
	"time"

	lua "github.com/yuin/gopher-lua"
)

var messageMethods = map[string]lua.LGFunction{
	"getText":     messageText,
	"isPrivate":   messageIsPrivate,
	"getType":     messageGetType,
	"getChannel":  messageGetChannel,
	"getSender":   messageGetSender,
	"getReceiver": messageGetReceiver,
	"findNode":    messageFindNode,
	"reply":       messageReply,
}

var userMethods = map[string]lua.LGFunction{
	"getId":           userGetId,
	"getIdExpression": userGetIDExpression,
	"getShortName":    userGetShortName,
	"getLongName":     userGetLongName,
	"verboseString":   userVerboseString,
	"getPosition":     userGetPosition,
	"getHopsAway":     userGetHopsAway,
	"getRSSI":         userGetRSSI,
	"getSNR":          userGetSNR,
	"isSelf":          userIsSelf,
}

// Checks whether the first lua argument is a *LUserData with *ChatMessage and returns this *ChatMessage
func checkMessage(L *lua.LState) *ChatMessage {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*ChatMessage); ok {
		return v
	}
	L.ArgError(1, "message expected")
	return nil
}

func messageText(L *lua.LState) int {
	message := *checkMessage(L)
	if message == nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LString(message.GetText()))
	return 1
}

func messageIsPrivate(L *lua.LState) int {
	message := *checkMessage(L)
	if message == nil {
		L.Push(lua.LFalse)
		return 1
	}
	L.Push(lua.LBool(message.IsPrivateMessage()))
	return 1
}

func messageGetType(L *lua.LState) int {
	message := *checkMessage(L)
	if message == nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LString(message.GetType()))
	return 1
}

func messageGetChannel(L *lua.LState) int {
	message := *checkMessage(L)
	if message == nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LString(message.GetChannelName()))
	return 1
}

func messageGetSender(L *lua.LState) int {
	message := *checkMessage(L)
	if message == nil {
		L.Push(lua.LNil)
		return 1
	}
	node := message.GetSenderNode()
	if node == nil {
		L.Push(lua.LNil)
		return 1
	}
	userUserData := L.NewUserData()
	userUserData.Value = &node
	L.SetMetatable(userUserData, L.GetTypeMetatable(luaUserTypeName))
	L.Push(userUserData)
	return 1
}

func messageGetReceiver(L *lua.LState) int {
	message := *checkMessage(L)
	if message == nil {
		L.Push(lua.LNil)
		return 1
	}
	node := message.GetReceiverNode()
	if node == nil {
		L.Push(lua.LNil)
		return 1
	}
	userUserData := L.NewUserData()
	userUserData.Value = &node
	L.SetMetatable(userUserData, L.GetTypeMetatable(luaUserTypeName))
	L.Push(userUserData)
	return 1
}

func messageFindNode(L *lua.LState) int {
	message := *checkMessage(L)
	if message == nil {
		L.Push(lua.LNil)
		return 1
	}
	node := message.FindNode(L.CheckString(2))
	if node == nil {
		L.Push(lua.LNil)
		return 1
	}
	userUserData := L.NewUserData()
	userUserData.Value = &node
	L.SetMetatable(userUserData, L.GetTypeMetatable(luaUserTypeName))
	L.Push(userUserData)
	return 1
}

func messageToString(L *lua.LState) int {
	message := *checkMessage(L)
	if message == nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LString(message.String()))
	return 1
}

func messageReply(L *lua.LState) int {
	message := *checkMessage(L)
	text := L.CheckString(2)
	timeout := L.OptInt(3, -1)

	channel := make(chan lua.LValue, 1)
	L.Push(lua.LChannel(channel))

	if message == nil || text == "" {
		channel <- lua.LFalse
		close(channel)
		return 1
	}

	go func(L *lua.LState, text string, timeout int) {
		if timeout == -1 {
			channel <- lua.LBool(<-message.Reply(text))
		} else {
			timeout := time.Second * time.Duration(timeout)
			channel <- lua.LBool(<-message.Reply(text, timeout))
		}
		close(channel)
	}(L, text, timeout)

	return 1
}

func checkUser(L *lua.LState) *ChatUser {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*ChatUser); ok {
		return v
	}
	L.ArgError(1, "user expected")
	return nil
}

func userGetId(L *lua.LState) int {
	user := *checkUser(L)
	if user == nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LNumber(user.GetId()))
	return 1
}

func userGetIDExpression(L *lua.LState) int {
	user := *checkUser(L)
	if user == nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LString(user.GetIDExpression()))
	return 1
}

func userGetShortName(L *lua.LState) int {
	user := *checkUser(L)
	if user == nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LString(user.GetShortName()))
	return 1
}

func userGetLongName(L *lua.LState) int {
	user := *checkUser(L)
	if user == nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LString(user.GetLongName()))
	return 1
}

func userToString(L *lua.LState) int {
	user := *checkUser(L)
	if user == nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LString(user.String()))
	return 1
}

func userVerboseString(L *lua.LState) int {
	user := *checkUser(L)
	if user == nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LString(user.VerboseString()))
	return 1
}

func userGetPosition(L *lua.LState) int {
	user := *checkUser(L)
	if user == nil {
		L.Push(lua.LNil)
		return 1
	}
	position := user.GetPosition()
	L.Push(lua.LNumber(position[0]))
	L.Push(lua.LNumber(position[1]))
	L.Push(lua.LNumber(position[2]))
	return 1
}

func userGetHopsAway(L *lua.LState) int {
	user := *checkUser(L)
	if user == nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LNumber(user.GetHopsAway()))
	return 1
}

func userGetRSSI(L *lua.LState) int {
	user := *checkUser(L)
	if user == nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LNumber(user.GetRSSI()))
	return 1
}

func userGetSNR(L *lua.LState) int {
	user := *checkUser(L)
	if user == nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LNumber(user.GetSNR()))
	return 1
}

func userIsSelf(L *lua.LState) int {
	user := *checkUser(L)
	if user == nil {
		L.Push(lua.LFalse)
		return 1
	}
	L.Push(lua.LBool(user.IsSelf()))
	return 1
}
