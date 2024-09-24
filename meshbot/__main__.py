import sys
import os
import time
import threading
import logging
from dotenv import dotenv_values

from .meshwrapper import MeshtasticClient, Message, MeshtasticConnectionLost
from .message_box import handle as message_box
from .radio_commands import handle as radio_commands

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("Meshbot")

config = {
    **dotenv_values(".env"),
    **dotenv_values("production.env"),
    **dotenv_values("development.env"),
    **os.environ,
}


# Define event handlers


def connectionHandler(meshtasticClient: MeshtasticClient):
    logger.info("Connection established!")
    logger.info(meshtasticClient.nodelist())


def messageHandler(message: Message, meshtasticClient: MeshtasticClient):
    logger.info(message)  # So we can actually see messages coming in on the terminal

    if radio_commands(message, meshtasticClient):
        return
    if message_box(message, meshtasticClient):
        return

    # If someone sends us a direct message that's not handled above, reply
    if message.type != "TEXT_MESSAGE_APP":
        return

    if message.text.upper() in ["/ABOUT", "/HELP", "/MESHBOT"]:
        return message.reply(
            "🤖👋 Hello! I'm your friendly neighbourhood Meshbot. My code is available at https://github.com/timendus/meshbot. Send me a direct message to see what I can do!"
        )

    if message.toNode.is_self():
        message.reply(
            """🤖👋 Hey there! I understand these message box commands:

- INBOX: Check your inbox
- NEW: Get new messages
- OLD: Get old messages
- CLEAR: Clear old messages
- SEND <id> <message>: Leave a message

[1/2]
"""
        )
        message.reply(
            """As well as these slash commands:

- /NODES: Get a summary of nodes
- /NODELIST: Get a list of the nodes I see
- /SIGNAL [<id>]: Get signal report on a node

[2/2]
"""
        )


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
        message=lambda message: messageHandler(message, meshtasticClient),
        debug=DEBUG,
    )
elif config["TRANSPORT"] == "net":
    host = "meshtastic.local" if config["DEVICE"] == "detect" else config["DEVICE"]
    logger.info(f"Attempting to connect to {host}...")
    meshtasticClient = MeshtasticClient(
        hostname=host,
        connected=lambda: connectionHandler(meshtasticClient),
        message=lambda message: messageHandler(message, meshtasticClient),
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
