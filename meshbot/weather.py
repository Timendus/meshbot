from .meshwrapper import Message
from .chatbot import Chatbot
from .open_meteo import fetch_weather, fetch_forecast


def register(bot: Chatbot):
    bot.add_command(
        {
            "command": "/WEATHER",
            "module": "🌂 Weather requests",
            "description": "Get the current weather",
            "channel": True,
            "function": get_weather,
        },
        {
            "command": "/FORECAST",
            "module": "🌂 Weather requests",
            "description": "Get a weather forecast",
            "channel": True,
            "function": get_forecast,
        },
    )


def get_weather(message: Message):
    if message.fromNode.position:
        position = message.fromNode.position
        location_text = "Here's the current weather at your location:"
    elif message.nodelist.get_self() and message.nodelist.get_self().position:
        position = message.nodelist.get_self().position
        location_text = "I can't see your location, so I'll give you the current weather at my location:"
    else:
        message.reply(
            f"🤖🧨 I'm sorry! I can't give you a weather report, because I don't know the location of either of us."
        )
        return

    weather = fetch_weather(position)
    if weather:
        message.reply(f"🤖🌂 {location_text}\n\n{weather}")
    else:
        message.reply(f"🤖🌂 I can't get a weather report at this time.")


def get_forecast(message: Message):
    if message.fromNode.position:
        position = message.fromNode.position
        location_text = "Here's the weather forecast for your location:"
    elif message.nodelist.get_self() and message.nodelist.get_self().position:
        position = message.nodelist.get_self().position
        location_text = "I can't see your location, so I'll give you the weather forecast for my location:"
    else:
        message.reply(
            f"🤖🧨 I'm sorry! I can't give you a weather forecast, because I don't know the location of either of us."
        )
        return

    forecast = fetch_forecast(position)
    if forecast:
        message.reply(f"🤖🌂 {location_text}\n\n{forecast}")
    else:
        message.reply(f"🤖🌂 I can't get a weather forecast at this time.")
