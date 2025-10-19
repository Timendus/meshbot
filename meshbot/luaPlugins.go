package meshbot

import (
	"context"
	"errors"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type plugin struct {
	Name        string
	Description string
	Version     string
	Hidden      bool
	Commands    []command
	States      []State
	LuaState    *lua.LState
}

type command struct {
	State          State
	Command        []string
	Prefix         []string
	Description    string
	Private        bool
	Channel        bool
	IsCatchAll     bool
	IsCatchAllText bool
	Hidden         bool
	Function       func(*ChatMessage) (State, error)
}

type contextKey string

const (
	luaMessageTypeName = "message"
	luaUserTypeName    = "user"
	CATCH_ALL_EVENTS   = iota
	CATCH_ALL_TEXT
)

func LoadPlugin(filename string, bot *Chatbot) (*plugin, error) {
	L := createLuaVM(bot)
	if err := L.DoFile(filename); err != nil {
		return nil, err
	}
	definition, ok := L.GetGlobal("Plugin").(*lua.LTable)
	if !ok {
		return nil, errors.New("no plugin definition found in file " + filename)
	}
	return newPlugin(definition, L)
}

func newPlugin(definition *lua.LTable, L *lua.LState) (*plugin, error) {
	plugin := plugin{
		Name:        definition.RawGetString("name").String(),
		Description: definition.RawGetString("description").String(),
		Version:     definition.RawGetString("version").String(),
		Hidden:      lua.LVAsBool(definition.RawGetString("hidden")),
		Commands:    make([]command, 0),
		States:      make([]State, 0),
		LuaState:    L,
	}

	commands := definition.RawGetString("commands")
	errs := make([]error, 0)
	if commands, ok := commands.(*lua.LTable); ok {
		commands.ForEach(func(k, v lua.LValue) {
			command, err := newCommand(v.(*lua.LTable), L)
			if err != nil {
				errs = append(errs, err)
			} else {
				plugin.Commands = append(plugin.Commands, *command)
			}
		})
	} else {
		return nil, errors.New("can't have a plugin without commands")
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	states := definition.RawGetString("states")
	if states, ok := states.(*lua.LTable); ok {
		states.ForEach(func(k, v lua.LValue) {
			plugin.States = append(plugin.States, State(v.String()))
		})
	}

	return &plugin, nil
}

func newCommand(definition *lua.LTable, L *lua.LState) (*command, error) {
	state := definition.RawGetString("state").String()
	if state == "nil" {
		state = "MAIN"
	}

	private := definition.RawGetString("private")
	if private == lua.LNil {
		private = lua.LTrue
	}

	luaFunction, ok := definition.RawGetString("func").(*lua.LFunction)
	if !ok {
		return nil, errors.New("can't have a command without a function")
	}

	command := command{
		State:          State(state),
		Command:        make([]string, 0),
		Prefix:         make([]string, 0),
		Description:    definition.RawGetString("description").String(),
		Private:        lua.LVAsBool(private),
		Channel:        lua.LVAsBool(definition.RawGetString("channel")),
		IsCatchAll:     false,
		IsCatchAllText: false,
		Hidden:         lua.LVAsBool(definition.RawGetString("hidden")),
		Function: func(message *ChatMessage) (State, error) {
			messageUserData := L.NewUserData()
			messageUserData.Value = message
			L.SetMetatable(messageUserData, L.GetTypeMetatable(luaMessageTypeName))
			err := L.CallByParam(lua.P{
				Fn:      luaFunction,
				NRet:    1,
				Protect: true,
			}, messageUserData)
			if err != nil {
				return "MAIN", err
			}
			ret := L.Get(-1)
			L.Pop(1)
			if ret.Type() == lua.LTNil {
				return "MAIN", nil
			} else {
				return State(ret.String()), nil
			}
		},
	}

	commands := definition.RawGetString("command")
	if commands, ok := commands.(*lua.LTable); ok {
		commands.ForEach(func(k, v lua.LValue) {
			command.Command = append(command.Command, v.String())
		})
	}
	if cmd, ok := commands.(lua.LString); ok {
		command.Command = append(command.Command, cmd.String())
	}

	prefixes := definition.RawGetString("prefix")
	if prefixes, ok := prefixes.(*lua.LTable); ok {
		prefixes.ForEach(func(k, v lua.LValue) {
			command.Prefix = append(command.Prefix, v.String())
		})
	}
	if prefix, ok := prefixes.(lua.LString); ok {
		command.Prefix = append(command.Prefix, prefix.String())
	}

	if cmd, ok := commands.(lua.LNumber); ok {
		if cmd == CATCH_ALL_EVENTS {
			command.IsCatchAll = true
		}
		if cmd == CATCH_ALL_TEXT {
			command.IsCatchAllText = true
		}
	}

	return &command, nil
}

func createLuaVM(cb *Chatbot) *lua.LState {
	// Initialize a bare-bones Lua VM
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	lua.OpenBase(L)
	lua.OpenString(L)
	lua.OpenTable(L)
	lua.OpenChannel(L)

	// Make some properties of the bot available to Lua
	bot := L.NewTable()
	L.SetGlobal("Bot", bot)
	bot.RawSetString("CATCH_ALL_TEXT", lua.LNumber(CATCH_ALL_TEXT))
	bot.RawSetString("CATCH_ALL_EVENTS", lua.LNumber(CATCH_ALL_EVENTS))
	botMT := L.NewTable()
	botMT.RawSetString("__tostring", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(cb.String()))
		return 1
	}))
	L.SetMetatable(bot, botMT)

	// Allow Lua scripts to get the time and date on bot, without access to the
	// whole `os` library
	L.SetField(bot, "date", L.NewFunction(func(L *lua.LState) int {
		format := L.OptString(1, "%c")
		L.Push(lua.LString(time.Now().Format(format)))
		return 1
	}))

	// This is pretty crude, but it provides a way to save some data from the
	// Lua scripts, that we can actually persist and make thread safe in the
	// future.
	L.SetContext(context.WithValue(context.Background(), contextKey("storage"), make(map[string]lua.LValue)))
	memory := L.NewTable()
	memory.RawSetString("write", L.NewFunction(func(L *lua.LState) int {
		ctx := L.Context()
		key := L.CheckString(1)
		ctx.Value(contextKey("storage")).(map[string]lua.LValue)[key] = L.Get(2)
		return 0
	}))
	memory.RawSetString("read", L.NewFunction(func(L *lua.LState) int {
		ctx := L.Context()
		key := L.CheckString(1)
		value, ok := ctx.Value(contextKey("storage")).(map[string]lua.LValue)[key]
		if ok {
			L.Push(value)
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))
	bot.RawSetString("memory", memory)

	// Register the Message usertype
	mmt := L.NewTypeMetatable(luaMessageTypeName)
	L.SetGlobal(luaMessageTypeName, mmt)
	L.SetField(mmt, "__index", L.SetFuncs(L.NewTable(), messageMethods))
	mmt.RawSetString("__tostring", L.NewFunction(messageToString))

	// Register the User usertype
	umt := L.NewTypeMetatable(luaUserTypeName)
	L.SetGlobal(luaUserTypeName, umt)
	L.SetField(umt, "__index", L.SetFuncs(L.NewTable(), userMethods))
	umt.RawSetString("__tostring", L.NewFunction(userToString))

	return L
}
