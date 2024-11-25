package main

import (
	"context"
	"fmt"
	"time"

	"github.com/timendus/meshbot/meshwrapper"
	lua "github.com/yuin/gopher-lua"
)

type plugin struct {
	Name        string
	Description string
	version     string
	Commands    []command
}

type command struct {
	State       State
	Command     []string
	Prefix      []string
	Description string
	Private     bool
	Channel     bool
	Function    *lua.LFunction
}

type loadedPlugin struct {
	Definition *plugin
	State      *lua.LState
}

type State string

const luaMessageTypeName = "message"

func main() {
	// Load plugin
	plugin := LoadPlugin("plugins/about.lua")
	defer plugin.State.Close()

	// Execute the first command from the plugin and print the result
	state, err := Execute(plugin.Definition.Commands[0], &meshwrapper.Message{
		MessageType: meshwrapper.MESSAGE_TYPE_TEXT_MESSAGE,
		Text:        "Hello World!",
	}, plugin.State)
	if err != nil {
		// Don't panic here, just provide feedback
		panic(err)
	}
	fmt.Println(state)
}

func LoadPlugin(filename string) loadedPlugin {
	L := createLuaVM()
	if err := L.DoFile(filename); err != nil {
		panic(err)
	}
	definition := L.GetGlobal("plugin").(*lua.LTable)
	return loadedPlugin{newPlugin(definition), L}
}

func Execute(command command, message *meshwrapper.Message, L *lua.LState) (State, error) {
	mud := L.NewUserData()
	mud.Value = message
	L.SetMetatable(mud, L.GetTypeMetatable(luaMessageTypeName))
	err := L.CallByParam(lua.P{
		Fn:      command.Function,
		NRet:    1,
		Protect: true,
	}, mud)
	if err != nil {
		return "ERROR", err
	}
	ret := L.Get(-1)
	L.Pop(1)
	return State(ret.String()), nil
}

func newPlugin(definition *lua.LTable) *plugin {
	plugin := plugin{
		Name:        definition.RawGetString("name").String(),
		Description: definition.RawGetString("description").String(),
		version:     definition.RawGetString("version").String(),
		Commands:    make([]command, 0),
	}

	commands := definition.RawGetString("commands")
	if commands, ok := commands.(*lua.LTable); ok {
		commands.ForEach(func(k, v lua.LValue) {
			command := newCommand(v.(*lua.LTable))
			plugin.Commands = append(plugin.Commands, command)
		})
	}

	return &plugin
}

func newCommand(definition *lua.LTable) command {
	commands := definition.RawGetString("command")
	commandList := make([]string, 0)
	if commands, ok := commands.(*lua.LTable); ok {
		commands.ForEach(func(k, v lua.LValue) {
			commandList = append(commandList, v.String())
		})
	}
	if command, ok := commands.(lua.LString); ok {
		commandList = append(commandList, command.String())
	}
	// TODO: fix this, commands can be catch all, how?
	if command, ok := commands.(lua.LNumber); ok {
		commandList = append(commandList, command.String())
	}

	prefixes := definition.RawGetString("prefix")
	prefixList := make([]string, 0)
	if prefixes, ok := prefixes.(*lua.LTable); ok {
		prefixes.ForEach(func(k, v lua.LValue) {
			prefixList = append(prefixList, v.String())
		})
	}
	if prefix, ok := prefixes.(lua.LString); ok {
		prefixList = append(prefixList, prefix.String())
	}

	state := definition.RawGetString("state").String()
	if state == "nil" {
		state = "MAIN"
	}

	command := command{
		State:       State(state),
		Command:     commandList,
		Prefix:      prefixList,
		Description: definition.RawGetString("description").String(),
		Private:     lua.LVAsBool(definition.RawGetString("private")),
		Channel:     lua.LVAsBool(definition.RawGetString("channel")),
		Function:    definition.RawGetString("func").(*lua.LFunction),
	}

	return command
}

func createLuaVM() *lua.LState {
	// Initialize a bare-bones Lua VM
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	lua.OpenBase(L)

	// Make some properties of the bot available to Lua
	bot := L.NewTable()
	L.SetGlobal("bot", bot)
	bot.RawSetString("CATCH_ALL_TEXT", lua.LNumber(meshwrapper.TextMessageEvent))
	bot.RawSetString("CATCH_ALL_EVENTS", lua.LNumber(meshwrapper.AnyMessageEvent))
	botMT := L.NewTable()
	botMT.RawSetString("__tostring", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("Hello, world!"))
		return 1
	}))
	L.SetMetatable(bot, botMT)

	// This is pretty crude, but it provides a way to save some data from the
	// Lua scripts, that we can actually persist and make thread safe in the
	// future.
	L.SetContext(context.WithValue(context.Background(), "storage", make(map[string]string)))
	memory := L.NewTable()
	memory.RawSetString("write", L.NewFunction(func(L *lua.LState) int {
		ctx := L.Context()
		key := L.CheckString(1)
		value := L.CheckString(2)
		ctx.Value("storage").(map[string]string)[key] = value
		return 0
	}))
	memory.RawSetString("read", L.NewFunction(func(L *lua.LState) int {
		ctx := L.Context()
		key := L.CheckString(1)
		value := ctx.Value("storage").(map[string]string)[key]
		L.Push(lua.LString(value))
		return 1
	}))
	bot.RawSetString("memory", memory)

	// Register the Message usertype
	mt := L.NewTypeMetatable(luaMessageTypeName)
	L.SetGlobal(luaMessageTypeName, mt)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), messageMethods))

	return L
}

var messageMethods = map[string]lua.LGFunction{
	"reply":         messageReply,
	"replyBlocking": messageReplyBlocking,
}

// Checks whether the first lua argument is a *LUserData with *Message and returns this *Message
func checkMessage(L *lua.LState) *meshwrapper.Message {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*meshwrapper.Message); ok {
		return v
	}
	L.ArgError(1, "message expected")
	return nil
}

func messageReply(L *lua.LState) int {
	message := checkMessage(L)
	message.Reply(L.CheckString(2))
	return 0
}

func messageReplyBlocking(L *lua.LState) int {
	message := checkMessage(L)
	timeout := time.Second * time.Duration(L.OptInt(3, int(meshwrapper.DEFAULT_BLOCKING_MESSAGE_TIMEOUT)))
	delivered := <-message.ReplyBlocking(L.CheckString(2), timeout)
	L.Push(lua.LBool(delivered))
	return 1
}
