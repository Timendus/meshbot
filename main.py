import sys
from time import sleep

from meshwrapper import MeshtasticClient, Message, MeshtasticConnectionLost
import message_box
import signal_reporter


# Define event handlers


def connectionHandler(meshtasticClient: MeshtasticClient):
    print("Connection established!")
    print("")
    print(meshtasticClient.nodeList)


def messageHandler(message: Message, meshtasticClient: MeshtasticClient):
    print(message)  # So we can actually see messages coming in on the terminal

    if signal_reporter.handle(message, meshtasticClient):
        return
    if message_box.handle(message, meshtasticClient):
        return

    # If someone sends us a direct message that's not handled above, reply
    if message.type == "TEXT_MESSAGE_APP" and message.toNode.isSelf:
        message.reply(
            """🤖👋 Hey there! Available commands:

- INBOX: Check your inbox
- NEW: Get new messages
- OLD: Get old messages
- CLEAR: Clear old messages
- SEND <id> <message>: Leave a message
- /SIGNAL [<id>]: Get signal report
"""
        )


# Find hostname parameter


if len(sys.argv) < 2:
    print(f"Usage: {sys.argv[0]} <hostname or IP>")
    sys.exit(1)

hostname = sys.argv[len(sys.argv) - 1]


# Start the connection to the Meshtastic node


print("Attempting to connect...")
meshtasticClient = MeshtasticClient(
    hostname,
    connected=lambda: connectionHandler(meshtasticClient),
    message=lambda message: messageHandler(message, meshtasticClient),
    debug=False,
)


# Keep the connection open until the user presses Ctrl+C or the device
# disconnects on the other side


try:
    while True:
        sleep(1)
except KeyboardInterrupt:
    print("Closing connection...")
    meshtasticClient.close()
except MeshtasticConnectionLost:
    print("Connection lost!")
finally:
    print("Done!")
