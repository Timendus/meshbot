from datetime import datetime

from .meshwrapper import MeshtasticClient, Message
from .meshwrapper.time_helper import time_ago


messageStore = {}


def handle(message: Message, meshtasticClient: MeshtasticClient) -> bool:
    # Only reply to text messages
    if message.type != "TEXT_MESSAGE_APP":
        return False
    # Only reply if the message is sent to us
    if not message.toNode.is_self():
        return False

    # Give the current user an inbox and a welcome message if they are new
    if message.fromNode.id not in messageStore:
        messageStore[message.fromNode.id] = [
            {
                "sender": "🤖 Meshbot",
                "contents": f"Welcome to this Meshtastic answering machine, {message.fromNode.longName}! You can leave messages for other users, and they can leave messages for you! Hope you like it 😄",
                "read": False,
                "timestamp": datetime.now(),
            },
        ]

    # Store new messages
    if message.text.upper().startswith("SEND"):
        parts = message.text.split(" ")
        msg = " ".join(parts[2:])

        if len(msg) == 0:
            message.reply("🤖🧨 I'm sorry, I can't send an empty message.")
            return True

        # Figure out who the recipient is
        id = parts[1]
        recipientId = meshtasticClient.nodelist().find_id(id)
        if not recipientId:
            message.reply(
                "🤖🧨 I don't know who that is. The message was not stored.\n\nI need the short name of a node I have seen before (example: TDRP), or the node ID of the recipient (example: !8e92a31f)."
            )
            return True

        # Store the message
        if recipientId not in messageStore:
            messageStore[recipientId] = []
        messageStore[recipientId].append(
            {
                "sender": message.fromNode.to_succinct_string(),
                "contents": msg,
                "read": False,
                "timestamp": datetime.now(),
            }
        )
        message.reply(f"🤖📨 Saved this message for node `{id}`:\n\n{msg}")
        return True

    # Collect some stats
    messages = messageStore[message.fromNode.id]
    numUnread = sum(1 for msg in messages if not msg["read"])
    totalMessages = len(messages)
    numRead = totalMessages - numUnread

    # Handle other commands
    match message.text.upper():

        case "INBOX":
            if totalMessages == 0:
                message.reply("🤖📭 You have no messages in your inbox")
            else:
                icon = "📬" if numUnread > 0 else "📭"
                message.reply(
                    f"🤖{icon} You have {numUnread} unread {'message' if numUnread == 1 else 'messages'}, and a grand total of {totalMessages} {'message' if totalMessages == 1 else 'messages'} in your inbox. Send `NEW` or `OLD` to fetch your messages."
                )

        case "NEW":
            if numUnread == 0:
                message.reply(
                    "🤖📭 You have no new messages. Send `OLD` to read your older messages."
                )
                return True

            message.reply(
                f"🤖📬 You have {numUnread} new {'message' if numUnread == 1 else 'messages'}. Sending {'it' if numUnread == 1 else 'them'} now...",
                wantAck=True,
            )
            for msg in messages:
                if not msg["read"]:
                    msg["read"] = True
                    message.reply(
                        f"🤖✉️ From {msg['sender']}, {time_ago(msg['timestamp'])} ago:\n\n{msg['contents']}"
                    )

        case "OLD":
            if numRead == 0:
                message.reply(
                    "🤖📭 You have no old messages. Send `NEW` to read your new messages."
                )
                return True

            message.reply(
                f"🤖📬 You have {numRead} old {'message' if numRead == 1 else 'messages'}. Sending {'it' if numRead == 1 else 'them'} now...",
                wantAck=True,
            )
            for msg in messages:
                if msg["read"]:
                    message.reply(
                        f"🤖✉️ From {msg['sender']}, {time_ago(msg['timestamp'])} ago:\n\n{msg['contents']}"
                    )

        case "CLEAR":
            messageStore[message.fromNode.id] = [
                msg for msg in messageStore[message.fromNode.id] if not msg["read"]
            ]
            message.reply(
                f"🤖🗑️ I removed {numRead} old {'message' if numRead == 1 else 'messages'}. You have {numUnread} new {'message' if numUnread == 1 else 'messages'} left in your inbox."
            )

        case "NODES":
            message.reply(meshtasticClient.nodeList.to_minimal_string())

        case _:
            return False

    return True
