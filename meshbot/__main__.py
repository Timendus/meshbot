import sys
import os
import time
import threading
import logging
from dotenv import dotenv_values

from .meshwrapper import MeshtasticClient, Message, MeshtasticConnectionLost
from .chatbot import Chatbot

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("Meshbot")

config = {
    **dotenv_values(".env"),
    **dotenv_values("production.env"),
    **dotenv_values("development.env"),
    **os.environ,
}


# Create a bot and register the desired modules with it

bot = Chatbot()
for module in [
    "about",
    "radio_commands",
    "weather",
    "message_box",
    "ollama_llm",
]:
    exec(f"from meshbot.{module} import register as register_{module}")
    exec(f"register_{module}(bot)")


# Define event handlers


def connectionHandler(meshtasticClient: MeshtasticClient):
    logger.info("Connection established!")
    logger.info(meshtasticClient.nodelist())


def messageHandler(message: Message):
    logger.info(message)  # So we can actually see messages coming in on the terminal
    bot.handle(message)


# Start the connection to the Meshtastic node


DEBUG = False

if config["TRANSPORT"] == "serial":
    if config["DEVICE"] == "detect":
        logger.info("Trying to find serial device...")
    else:
        logger.info(f"Attempting to open serial connection on {config['DEVICE']}...")
    meshtasticClient = MeshtasticClient(
        device=None if config["DEVICE"] == "detect" else config["DEVICE"],
        connected=lambda: connectionHandler(meshtasticClient),
        message=messageHandler,
        debug=DEBUG,
    )
elif config["TRANSPORT"] == "net":
    host = "meshtastic.local" if config["DEVICE"] == "detect" else config["DEVICE"]
    logger.info(f"Attempting to connect to {host}...")
    meshtasticClient = MeshtasticClient(
        hostname=host,
        connected=lambda: connectionHandler(meshtasticClient),
        message=messageHandler,
        debug=DEBUG,
    )
else:
    raise Exception(f"Unknown transport: {config['TRANSPORT']}")


# Output the node list every half hour


class setInterval:
    def __init__(self, interval, action):
        self.interval = interval
        self.action = action
        self.stopEvent = threading.Event()
        thread = threading.Thread(target=self.__setInterval)
        thread.start()

    def __setInterval(self):
        nextTime = time.time() + self.interval
        while not self.stopEvent.wait(nextTime - time.time()):
            nextTime += self.interval
            self.action()

    def cancel(self):
        self.stopEvent.set()


interval = setInterval(30 * 60, lambda: logger.info(meshtasticClient.nodelist()))


# Keep the connection open until the user presses Ctrl+C or the device
# disconnects on the other side


try:
    while True:
        time.sleep(1000)
except KeyboardInterrupt:
    logger.info("Closing connection...")
    meshtasticClient.close()
except MeshtasticConnectionLost:
    logger.error("Connection lost!")
finally:
    logger.info("Done!")
    interval.cancel()
