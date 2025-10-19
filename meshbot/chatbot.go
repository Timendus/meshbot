package meshbot

import (
	"fmt"
	"log"
	"os"
	"strings"
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
		if !strings.HasSuffix(entry.Name(), ".lua") {
			continue
		}
		plugin, err := LoadPlugin("plugins/"+entry.Name(), c)
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

func (c *Chatbot) HandleMessage(message ChatMessage) {
	// See if we have one or more catch all handlers
	c.handleMessageIf(message, func(cmd command, _ string) bool { return cmd.IsCatchAll })

	// Messages that are not text messages can only be handled by
	// catch all commands, so in that case we're done here.
	if message.GetType() != TEXT_MESSAGE {
		return
	}

	// See if we have one or more specific handlers for this text message
	if c.handleMessageIf(message, matches) {
		return
	}

	// See if we have one or more catch all text handlers for this text message
	c.handleMessageIf(message, func(cmd command, _ string) bool { return cmd.IsCatchAllText })
}

func (c *Chatbot) handleMessageIf(message ChatMessage, comp func(command, string) bool) bool {
	isPrivateMessage := message.IsPrivateMessage()
	matchFound := false
	for _, plugin := range c.plugins {
		for _, cmd := range plugin.Commands {
			validCommand := cmd.State == c.state &&
				(cmd.Private == isPrivateMessage ||
					cmd.Channel == !isPrivateMessage)
			if validCommand && comp(cmd, message.GetText()) {
				matchFound = true
				c.runFunction(cmd, message)
			}
		}
	}
	return matchFound
}

func (c *Chatbot) runFunction(cmd command, message ChatMessage) {
	newState, err := cmd.Function(&message)
	if err != nil {
		log.Println("We got an error while handling a message:", err)
	} else {
		c.state = newState
	}
}

func matches(command command, message string) bool {
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
