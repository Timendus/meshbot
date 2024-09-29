#!/usr/bin/env python3

from meshbot.meshwrapper import Node, Message
from meshbot.chatbot import Chatbot

# Create a bot
bot = Chatbot()

# Import and register all modules
for module in [
    "about",
    # "message_box",
    # "ollama_llm",
    # "radio_commands",
]:
    exec(f"from meshbot.{module} import register as register_{module}")
    exec(f"register_{module}(bot)")


def output(response: str):
    print(response)


print(bot)

# Take input from the user and run in through the bot
while True:
    try:
        message = Message()
        message.text = input(">>> ")
        message.type = "TEXT_MESSAGE_APP"
        message.reply = output
        message.toNode = Node()

        bot.handle(message, None)
    except KeyboardInterrupt:
        break
    except EOFError:
        break

print()
