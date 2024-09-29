from .chatbot import Chatbot


def register(bot: Chatbot):
    bot.add_command(
        {
            "command": ["/ABOUT", "/HELP", "/MESHBOT"],
            "function": lambda message, client: message.reply(
                "🤖👋 Hello! I'm your friendly neighbourhood Meshbot. My code is available at https://github.com/timendus/meshbot. Send me a direct message to see what I can do!"
            ),
            "channel": True,
        }
    )
