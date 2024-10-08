from meshbot.chatbot import Chatbot
from meshbot.meshwrapper import Node, Message


def test_registration():
    bot = Chatbot()

    my_state = "MY_STATE"
    my_command = {
        "command": "TEST",
        "description": "Test command",
        "function": lambda m, c: "TEST",
        "state": "MAIN",
    }

    bot.add_state(my_state)
    bot.add_command(my_command)

    assert bot.states == ["MAIN", my_state]
    assert bot.commands == [my_command]


def test_multiple_registrations():
    bot = Chatbot()

    state1 = "MY_STATE_1"
    state2 = "MY_STATE_2"
    command1 = {
        "command": "TEST1",
        "description": "Test command 1",
        "function": lambda m, c: "TEST1",
        "state": "MAIN",
    }
    command2 = {
        "command": "TEST2",
        "description": "Test command 2",
        "function": lambda m, c: "TEST2",
        "state": "MAIN",
    }

    bot.add_state(state1, state2)
    bot.add_command(command1, command2)

    assert bot.states == ["MAIN", state1, state2]
    assert bot.commands == [command1, command2]

    bot.add_state(state1)
    bot.add_command(command1)

    assert bot.states == ["MAIN", state1, state2, state1]
    assert bot.commands == [command1, command2, command1]


def test_to_string():
    bot = Chatbot()
    bot.add_command(
        {
            "command": "TEST1",
            "module": "Test Module",
            "description": "Test command 1",
            "function": lambda m, c: "TEST2",
            "state": "MAIN",
        }
    )
    bot.add_command(
        {
            "command": "TEST2",
            "description": "Test command 2",
            "function": lambda m, c: "TEST",
            "state": "MAIN",
        }
    )

    assert (
        str(bot)
        == """🤖👋 Hey there! I understand these commands:

Test Module
- TEST1: Test command 1

General commands
- TEST2: Test command 2
"""
    )


def test_simple_message_handling():
    bot = Chatbot()
    called = 0

    message = Message()
    message.text = "test"
    message.type = "TEXT_MESSAGE_APP"
    message.toNode = Node()

    def callback(m):
        nonlocal called
        assert m == message
        called += 1

    bot.add_command(
        {
            "command": "TEST",
            "description": "Test command",
            "function": callback,
            "state": "MAIN",
        }
    )

    bot.handle(message)
    bot.handle(message)

    assert called == 2, "Test message should have been handled by test command twice"


def test_specific_before_catch_all_message_handling():
    bot = Chatbot()
    called = False

    message = Message()
    message.text = "TEST"
    message.type = "TEXT_MESSAGE_APP"
    message.toNode = Node()

    def callback1(m):
        nonlocal called
        assert m == message
        called = True

    def callback2(m):
        assert False, "This should not be called"

    bot.add_command(
        {
            "command": Chatbot.CATCH_ALL_TEXT,
            "description": "Test command",
            "function": callback2,
            "state": "MAIN",
        },
        {
            "command": "TEST",
            "description": "Test command",
            "function": callback1,
            "state": "MAIN",
        },
    )

    bot.handle(message)

    assert called, "Test message should have been handled by test command"


def test_catch_all_message_handling():
    bot = Chatbot()
    called = False

    message = Message()
    message.text = "TEST"
    message.type = "TEXT_MESSAGE_APP"
    message.toNode = Node()

    def callback(m):
        nonlocal called
        assert m == message
        called = True

    bot.add_command(
        {
            "command": Chatbot.CATCH_ALL_TEXT,
            "description": "Test command",
            "function": callback,
            "state": "MAIN",
        }
    )

    bot.handle(message)

    assert called, "Test message should have been handled by catch all command"


def test_ignore_other_events_message_handling():
    bot = Chatbot()

    message = Message()
    message.text = "TEST"
    message.type = "TELEMETRY_APP"
    message.toNode = Node()

    def callback(m):
        assert False, "This should not be called"

    bot.add_command(
        {
            "command": Chatbot.CATCH_ALL_TEXT,
            "description": "Test command",
            "function": callback,
            "state": "MAIN",
        },
        {
            "command": "TEST",
            "description": "Test command",
            "function": callback,
            "state": "MAIN",
        },
    )

    bot.handle(message)

    assert True, "Telemetry packet should have been ignored"


def test_catch_all_events_message_handling():
    bot = Chatbot()
    called = False

    message = Message()
    message.text = "TEST"
    message.type = "TELEMETRY_APP"
    message.toNode = Node()

    def callback(m):
        nonlocal called
        assert m == message
        called = True

    bot.add_command(
        {
            "command": Chatbot.CATCH_ALL_EVENTS,
            "description": "Test command",
            "function": callback,
            "state": "MAIN",
        }
    )

    bot.handle(message)

    assert (
        called
    ), "Telemetry packet should have been handled by catch all events command"


def test_multiple_commands_message_handling():
    bot = Chatbot()
    called = False

    message = Message()
    message.text = "TEST"
    message.type = "TEXT_MESSAGE_APP"
    message.toNode = Node()

    def callback(m):
        nonlocal called
        assert m == message
        called = True

    bot.add_command(
        {
            "command": ["THINGS", "TEST"],
            "description": "Test command",
            "function": callback,
            "state": "MAIN",
        }
    )

    bot.handle(message)

    assert called, "Test message should have been handled by multi-command test command"


def test_multiple_handlers_message_handling():
    bot = Chatbot()
    called = 0

    message = Message()
    message.text = "TEST"
    message.type = "TEXT_MESSAGE_APP"
    message.toNode = Node()

    def callback(m):
        nonlocal called
        assert m == message
        called += 1

    bot.add_command(
        {
            "command": "TEST",
            "description": "Test command 1",
            "function": callback,
            "state": "MAIN",
        },
        {
            "command": "TEST",
            "description": "Test command 2",
            "function": callback,
            "state": "MAIN",
        },
    )

    bot.handle(message)

    assert called == 2, "Test message should have been handled by both commands"


def test_multiple_catch_all_handlers_message_handling():
    bot = Chatbot()
    called = 0

    message = Message()
    message.text = "TEST"
    message.type = "TEXT_MESSAGE_APP"
    message.toNode = Node()

    def callback(m):
        nonlocal called
        assert m == message
        called += 1

    bot.add_command(
        {
            "command": Chatbot.CATCH_ALL_EVENTS,
            "description": "Test command 1",
            "function": callback,
            "state": "MAIN",
        },
        {
            "command": Chatbot.CATCH_ALL_TEXT,
            "description": "Test command 2",
            "function": callback,
            "state": "MAIN",
        },
    )

    bot.handle(message)

    assert called == 2, "Test message should have been handled by both commands"
