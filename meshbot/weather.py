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
    if hasattr(message.fromNode, "location"):
        location = message.fromNode.location
        location_text = "Here's the current weather at your location:"
    else:
        location = message.nodelist.get_self().location
        location_text = "I can't see your location, so I'll give you the current weather at my location:"

    weather = fetch_weather(location)
    if weather:
        message.reply(f"🤖🌂 {location_text}\n\n{weather}")
    else:
        message.reply(f"🤖🌂 I can't get a weather report at this time.")


def get_forecast(message: Message):
    if message.fromNode.location:
        location = message.fromNode.location
        location_text = "Here's the weather forecast for your location:"
    else:
        location = message.nodelist.get_self().location
        location_text = "I can't see your location, so I'll give you the weather forecast for my location:"

    forecast = fetch_forecast(location)
    if forecast:
        message.reply(f"🤖🌂 {location_text}\n\n{forecast}")
    else:
        message.reply(f"🤖🌂 I can't get a weather forecast at this time.")
