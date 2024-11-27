package meshbot

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/timendus/meshbot/meshwrapper"
)

type State string

type Chatbot struct {
	state   State
	plugins []*plugin
}

func NewChatbot() *Chatbot {
	return &Chatbot{
		state: "MAIN",
	}
}

func (c *Chatbot) ReloadPlugins() error {
	plugins := make([]*plugin, 0)
	entries, err := os.ReadDir("plugins")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		plugin, err := LoadPlugin("plugins/" + entry.Name())
		if err != nil {
			return err
		}
		plugins = append(plugins, plugin)
	}
	c.plugins = plugins
	return nil
}

func (c *Chatbot) String() string {
	description := "🤖👋 Hey there! I understand these commands:\n"

	for _, plugin := range c.plugins {
		if plugin.Hidden {
			continue
		}
		if plugin.Name != "nil" && plugin.Description != "nil" {
			description += fmt.Sprintf("\n%s - %s\n", plugin.Name, plugin.Description)
		} else if plugin.Name != "nil" {
			description += fmt.Sprintf("\n%s\n", plugin.Name)
		}
		for _, command := range plugin.Commands {
			if command.Hidden {
				continue
			}
			var commands string
			if len(command.Command) > 0 {
				commands = strings.Join(command.Command, ", ")
			} else if len(command.Prefix) > 0 {
				commands = strings.Join(command.Prefix, ", ")
			} else {
				continue
			}
			if command.Description != "nil" {
				description += fmt.Sprintf("- %s: %s\n", commands, command.Description)
			} else {
				description += fmt.Sprintf("- %s\n", commands)
			}
		}
	}

	return description
}

func (c *Chatbot) HandleMessage(message meshwrapper.Message) {
	// Messages that are not text messages can only be handled by
	// catch all commands
	if message.MessageType != meshwrapper.MESSAGE_TYPE_TEXT_MESSAGE {
		c.handleMessageIf(message, func(cmd command, _ string) bool { return cmd.IsCatchAll })
		return
	}

	// See if we have one or more specific handlers for this text message
	if c.handleMessageIf(message, c.matches) {
		return
	}

	// See if we have one or more catch all text handlers for this text message
	c.handleMessageIf(message, func(cmd command, _ string) bool { return cmd.IsCatchAllText })
}

func (c *Chatbot) handleMessageIf(message meshwrapper.Message, comp func(command, string) bool) bool {
	matchFound := false
	for _, plugin := range c.plugins {
		for _, command := range plugin.Commands {
			validCommand := command.State == c.state &&
				(command.Private == message.IsPrivateMessage() ||
					command.Channel == !message.IsPrivateMessage())
			if validCommand && comp(command, message.Text) {
				matchFound = true
				newState, err := command.Function(&message)
				if err != nil {
					log.Println("We got an error while handling a message:", err)
				} else {
					c.state = newState
				}
			}
		}
	}
	return matchFound
}

func (c *Chatbot) matches(command command, message string) bool {
	for _, command := range command.Command {
		if strings.EqualFold(strings.TrimSpace(message), strings.TrimSpace(command)) {
			return true
		}
	}
	for _, prefix := range command.Prefix {
		if len(strings.TrimSpace(message)) < len(strings.TrimSpace(prefix)) {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(message)[:len(strings.TrimSpace(prefix))], strings.TrimSpace(prefix)) {
			return true
		}
	}
	return false
}
