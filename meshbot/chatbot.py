from itertools import groupby
from typing import Callable

from .meshwrapper import Message


class Chatbot:
    """
    Helper class for defining the states and commands that a chatbot
    understands, and routing the incoming messages to the proper commands.

    This is the structure of the commands that this class understands:

    {
        "module": "Test module",        # Name of the module this command belongs to
        "command": "/TEST",             # Can be a single string, a list of commands or one of the catch alls
        "description": "Test command",  # If omitted, command will not be listed
        "state": "MAIN",                # State in which the command is valid (default: "MAIN")
        "private": True,                # Is command valid in private messages? (default: True)
        "channel": False,               # Is command valid in channel messages? (default: False)

        # Function to call when the command is received. Can optionally return a
        # string with the name of the next state to change to
        "function": lambda m, c: m.reply("Hello!"),
    }
    """

    CATCH_ALL_TEXT = 1  # Get all text messages
    CATCH_ALL_EVENTS = 2  # Get all packets

    def __init__(self):
        self.states = ["MAIN"]
        self.state = "MAIN"
        self.commands = []

    def add_state(self, *states):
        for state in states:
            self.states.append(state)

    def add_command(self, *commands):
        for command in commands:
            self.commands.append(command)

    def handle(self, message: Message) -> None:
        is_text_message = message.type == "TEXT_MESSAGE_APP"
        is_private_message = message.private_message()
        is_channel_message = not is_private_message

        # Find commands that are valid in this state and are of the right type
        relevant_commands = [
            c
            for c in self.commands
            if c.get("state", "MAIN") is self.state
            and (
                c.get("private", True) == is_private_message
                or c.get("channel", False) == is_channel_message
            )
        ]

        # Bail early if we have no relevant commands at all
        if len(relevant_commands) == 0:
            return

        # Messages that are not text messages can only be handled by
        # CATCH_ALL_EVENTS commands
        if not is_text_message:
            catch_all_events = [
                c
                for c in relevant_commands
                if self._matching(c, Chatbot.CATCH_ALL_EVENTS)
            ]
            for cmd in catch_all_events:
                self._run_function(cmd["function"], message)
            return

        # Messages that are text messages are evaluated specific first, catch
        # all later
        specific = [c for c in relevant_commands if self._matching(c, message.text)]
        for cmd in specific:
            self._run_function(cmd["function"], message)

        # Have we now handled this message?
        if len(specific) > 0:
            return

        # No specific command matched, try catch all
        catch_all = [
            c
            for c in relevant_commands
            if self._matching(c, Chatbot.CATCH_ALL_TEXT)
            or self._matching(c, Chatbot.CATCH_ALL_EVENTS)
        ]
        for cmd in catch_all:
            self._run_function(cmd["function"], message)
        return

    def _run_function(
        self,
        function: Callable[[Message], str | None],
        message: Message,
    ) -> None:
        assert function is not None, "Can't call a nonexistant function"
        new_state = function(message)
        if type(new_state) == str:
            self.state = new_state

    def __str__(self):
        description = "🤖👋 Hey there! I understand these commands:\n"

        self.commands.sort(key=lambda c: c.get("module", ""))
        for module, commands in groupby(
            self.commands,
            key=lambda c: c.get("module", None),
        ):
            commands = list(commands)
            if not any(self._visible(c) for c in commands):
                continue
            module = module or "General commands"
            description += f"\n{module}\n"
            for command in commands:
                if self._visible(command):
                    if "command" in command:
                        cmd = (
                            command["command"]
                            if type(command["command"]) != list
                            else ", ".join(command["command"])
                        )
                        description += f"- {cmd}: {command['description']}\n"
                    elif "prefix" in command:
                        description += f"- {command['description']}\n"

        return description

    def _matching(self, command, input):
        if "command" in command:
            if type(command["command"]) == list:
                return any(self._same(c, input) for c in command["command"])
            return self._same(command["command"], input)

        if "prefix" in command:
            if type(command["prefix"]) == list:
                return any(self._startsWith(c, input) for c in command["prefix"])
            return self._startsWith(command["prefix"], input)

        return False

    def _same(self, command, input):
        if type(command) == str and type(input) == str:
            return command.upper().strip() == input.upper().strip()
        return command is input

    def _startsWith(self, prefix, input):
        if type(prefix) == str and type(input) == str:
            return input.upper().strip().startswith(prefix.upper().strip())
        return False

    def _visible(self, command):
        return (
            (
                "command" in command
                and command["command"] is not Chatbot.CATCH_ALL_EVENTS
                and command["command"] is not Chatbot.CATCH_ALL_TEXT
            )
            or (
                "prefix" in command
                and command["prefix"] is not Chatbot.CATCH_ALL_EVENTS
                and command["prefix"] is not Chatbot.CATCH_ALL_TEXT
            )
        ) and command.get("description", None) is not None
