from .chatbot import Chatbot


def register(bot: Chatbot):
    bot.add_command(
        # This is a hidden command, which is not listed (because it has no
        # description), but might be "guessed" by users, and will result in
        # expected behaviour.
        {
            "command": ["/ABOUT", "/HELP", "/MESHBOT"],
            "channel": True,
            "function": lambda message: message.reply(
                "🤖👋 Hello! I'm your friendly neighbourhood Meshbot. My code is available at https://github.com/timendus/meshbot. Send me a direct message to see what I can do!"
            ),
        },
        # This is the "catch all" command, if no more specific command is
        # matched in the "MAIN" state when receiving a private message, we reply
        # with the capabilities of this bot. This too is not listed because it
        # has no description.
        {
            "command": Chatbot.CATCH_ALL_TEXT,
            "function": lambda message: message.reply(str(bot)),
        },
    )
