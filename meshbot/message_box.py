from datetime import datetime

from .meshwrapper import Message, Node
from .meshwrapper.time_helper import time_ago
from .chatbot import Chatbot


def register(bot: Chatbot):
    bot.add_command(
        {
            "command": "INBOX",
            "module": "✉️ Message box",
            "description": "Check your inbox",
            "function": send_inbox,
        },
        {
            "command": "NEW",
            "module": "✉️ Message box",
            "description": "Get new messages",
            "function": send_new_messages,
        },
        {
            "command": "OLD",
            "module": "✉️ Message box",
            "description": "Get old messages",
            "function": send_old_messages,
        },
        {
            "command": "CLEAR",
            "module": "✉️ Message box",
            "description": "Clear old messages",
            "function": clear_old_messages,
        },
        {
            "prefix": "SEND",
            "module": "✉️ Message box",
            "description": "SEND <id> <message>: Leave a message",
            "function": store_message,
        },
        {
            "command": Chatbot.CATCH_ALL_EVENTS,
            "module": "✉️ Message box",
            "function": notify_user,
        },
    )


messageStore = {}


def send_inbox(message: Message):
    _store_welcome_message(message.fromNode)
    stats = _user_stats(message.fromNode)

    if stats["totalMessages"] == 0:
        message.fromNode.send("🤖📭 You have no messages in your inbox")
        return

    icon = "📬" if stats["numUnread"] > 0 else "📭"
    message.fromNode.send(
        f"🤖{icon} You have {stats['numUnread']} unread {_pluralize('message', stats['numUnread'])}, and a grand total of {stats['totalMessages']} {_pluralize('message', stats['totalMessages'])} in your inbox. Send `NEW` or `OLD` to fetch your messages."
    )


def send_new_messages(message: Message):
    _store_welcome_message(message.fromNode)
    stats = _user_stats(message.fromNode)

    if stats["numUnread"] == 0:
        old_messages = (
            " Send `OLD` to read your older messages." if stats["numRead"] > 0 else ""
        )
        message.fromNode.send(f"🤖📭 You have no new messages.{old_messages}")
        return

    message.fromNode.send(
        f"🤖📬 You have {stats['numUnread']} new {_pluralize('message', stats['numUnread'])}. Sending {_pluralize('it', stats['numUnread'])} now..."
    )
    _send_messages(message.fromNode, read=False)


def send_old_messages(message: Message):
    _store_welcome_message(message.fromNode)
    stats = _user_stats(message.fromNode)

    if stats["numRead"] == 0:
        new_messages = (
            " Send `NEW` to read your new messages." if stats["numUnread"] > 0 else ""
        )
        message.fromNode.send(f"🤖📭 You have no old messages.{new_messages}")
        return

    message.fromNode.send(
        f"🤖📬 You have {stats['numRead']} old {_pluralize('message', stats['numRead'])}. Sending {_pluralize('it', stats['numRead'])} now..."
    )
    _send_messages(message.fromNode, read=True)


def clear_old_messages(message: Message):
    _store_welcome_message(message.fromNode)
    stats = _user_stats(message.fromNode)

    messageStore[message.fromNode.id] = [
        msg for msg in messageStore[message.fromNode.id] if not msg["read"]
    ]
    message.fromNode.send(
        f"🤖🗑️ I removed {stats['numRead']} old {_pluralize('message', stats['numRead'])}. You have {stats['numUnread']} new {_pluralize('message', stats['numUnread'])} left in your inbox."
    )


def store_message(message: Message):
    """
    Store new messages when requested by the user
    """

    _store_welcome_message(message.fromNode)

    parts = message.text.split(" ")
    msg = " ".join(parts[2:])

    if len(msg) == 0:
        message.fromNode.send("🤖🧨 I'm sorry, I can't send an empty message.")
        return

    # Figure out who the recipient is
    id = parts[1]
    recipientId = message.nodelist.find_id(id)
    if not recipientId:
        message.fromNode.send(
            "🤖🧨 I don't know who that is. The message was not stored.\n\nI need the short name of a node I have seen before (example: TDRP), or the node ID of the recipient (example: !8e92a31f)."
        )
        return

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
    message.fromNode.send(f"🤖📨 Saved this message for node `{id}`:\n\n{msg}")


def notify_user(message: Message):
    """
    Check to see if one of our recipients came in range, and has new messages.
    """

    # If they are messaging us first, they will probably quickly find out that
    # they have messages, and it just breaks the flow. So only check for all
    # other message types.
    if message.type == "TEXT_MESSAGE_APP" and message.toNode.is_self():
        return

    # We get routing messages for each Ack, so ignore those or we get a royal
    # clusterfuck.
    if message.type == "ROUTING_APP":
        return

    # Do we have a message box?
    if message.fromNode.id not in messageStore:
        return

    # Do we have new messages?
    stats = _user_stats(message.fromNode)
    if stats["numUnread"] == 0:
        return

    # Send this user their new messages
    message.fromNode.send(
        f"🤖📬 I have {stats['numUnread']} new {_pluralize('message', stats['numUnread'])} for you! Sending {_pluralize('it', stats['numUnread'])} now..."
    )
    _send_messages(message.fromNode, read=False)


def _store_welcome_message(node: Node):
    """
    Give the current user an inbox and a welcome message if they are new
    """
    if node.id not in messageStore:
        messageStore[node.id] = [
            {
                "sender": "🤖 Meshbot",
                "contents": f"Welcome to this Meshtastic answering machine, {node.longName}! You can leave messages for other users, and they can leave messages for you! Hope you like it 😄",
                "read": False,
                "timestamp": datetime.now(),
            },
        ]


def _send_messages(node: Node, read: bool = False):
    for msg in messageStore.get(node.id, []):
        if msg["read"] != read:
            continue
        if node.send(
            f"🤖✉️ From {msg['sender']}, {time_ago(msg['timestamp'])} ago:\n\n{msg['contents']}"
        ):
            msg["read"] = True


def _user_stats(node: Node) -> dict:
    messages = messageStore.get(node.id, [])
    numUnread = sum(1 for msg in messages if not msg["read"])
    totalMessages = len(messages)
    return {
        "totalMessages": totalMessages,
        "numUnread": numUnread,
        "numRead": totalMessages - numUnread,
    }


def _pluralize(word: str, count: int) -> str:
    if count == 1:
        return word
    if word == "it":
        return "them"
    return word + "s"
