#!/usr/bin/env python3

from datetime import datetime

from meshbot.meshwrapper import Node, Nodelist, Message
from meshbot.chatbot import Chatbot


# Create a bot

bot = Chatbot()


# Import desired modules and register them with the bot

for module in [
    "about",
    "message_box",
    "ollama_llm",
    "radio_commands",
]:
    exec(f"from meshbot.{module} import register as register_{module}")
    exec(f"register_{module}(bot)")


def output(response: str) -> bool:
    print(response)
    return True


print(bot)


# Create fake domain model


fromNode = Node()
fromNode.num = 1
fromNode.id = "!00000001"
fromNode.shortName = "USER"
fromNode.longName = "User"
fromNode.snr = 5.0
fromNode.rssi = -80
fromNode.hopsAway = 0
fromNode.send = output
fromNode.lastHeard = datetime.timestamp(datetime.now())

toNode = Node()
toNode.num = 2
toNode.id = "!00000002"
toNode.is_self = lambda: True
toNode.shortName = "MBOT"
toNode.longName = "Meshbot"
toNode.snr = 6.0
toNode.rssi = -75
toNode.hopsAway = 0
toNode.lastHeard = datetime.timestamp(datetime.now())

nodelist = Nodelist()
nodelist.add(fromNode)
nodelist.add(toNode)


# Take input from the user and run it through the bot

while True:
    try:
        message = Message()
        message.text = input(">>> ")
        message.type = "TEXT_MESSAGE_APP"
        message.reply = output
        message.fromNode = fromNode
        message.toNode = toNode
        message.nodelist = nodelist

        bot.handle(message)
    except KeyboardInterrupt:
        break
    except EOFError:
        break

print()
