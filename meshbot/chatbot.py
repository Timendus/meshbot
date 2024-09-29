from itertools import groupby
from typing import Callable

from .meshwrapper import MeshtasticClient, Message


class Chatbot:
    """
    Helper class for defining the states and commands that a chatbot
    understands, and routing the incoming messages to the proper commands
    """

    def __init__(self):
        self.states = ["MAIN"]
        self.state = "MAIN"
        self.commands = []
        self.CATCH_ALL_EVENTS = 1
        self.CATCH_ALL_TEXT = 2

    def add_state(self, *states):
        for state in states:
            self.states.append(state)

    def add_command(self, *commands):
        for command in commands:
            self.commands.append(command)

    def handle(self, message: Message, client: MeshtasticClient) -> None:
        is_text_message = message.type == "TEXT_MESSAGE_APP"
        is_private_message = message.toNode.is_self()
        is_channel_message = message.toNode.is_broadcast()

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
            catch_all_events = next(
                (
                    c
                    for c in relevant_commands
                    if self._matching(c, self.CATCH_ALL_EVENTS)
                ),
                None,
            )
            if catch_all_events:
                self._run_function(catch_all_events["function"], message, client)
            return

        # Messages that are text messages are evaluated specific first, catch
        # all later
        specific = next(
            (c for c in relevant_commands if self._matching(c, message.text)),
            None,
        )
        if specific:
            return self._run_function(specific["function"], message, client)

        catch_all = next(
            (c for c in relevant_commands if self._matching(c, self.CATCH_ALL_TEXT)),
            None,
        )
        if catch_all:
            return self._run_function(catch_all["function"], message, client)

        # No matches, just ignore this message
        return

    def _run_function(
        self,
        function: Callable[Message, MeshtasticClient],
        message: Message,
        client: MeshtasticClient,
    ) -> None:
        assert function is not None, "Can't call a nonexistant function"
        self.state = function(message, client) or self.state

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
            if module is not None:
                description += f"\n\n{module}"
            for command in commands:
                if self._visible(command):
                    cmd = (
                        command["command"]
                        if type(command["command"]) != list
                        else ", ".join(command["command"])
                    )
                    description += f"\n- {cmd}: {command['description']}"

        return description

    def _matching(self, command, input):
        if type(command) == dict:
            return self._matching(command["command"], input)
        if type(command) == list:
            return any(self._matching(c, input) for c in command)
        if type(command) == str and type(input) == str:
            return command.upper().strip() == input.upper().strip()
        return command == input

    def _visible(self, command):
        return (
            command["command"] is not self.CATCH_ALL_EVENTS
            and command["command"] is not self.CATCH_ALL_TEXT
            and command["description"] is not None
        )
