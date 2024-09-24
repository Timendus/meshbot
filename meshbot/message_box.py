from datetime import datetime

from .meshwrapper import MeshtasticClient, Message, Node
from .meshwrapper.time_helper import time_ago


messageStore = {}


def handle(message: Message, meshtasticClient: MeshtasticClient) -> bool:
    # Check to see if one of our recipients came in range, and has new messages
    if notify_user(message, meshtasticClient):
        return False  # We have notified the user, but not really handled the message per say

    # Otherwise, only reply to text messages that were sent to us
    if message.type != "TEXT_MESSAGE_APP" or not message.toNode.is_self():
        return False

    # Give the current user an inbox and a welcome message if they are new
    store_welcome_message(message.fromNode)

    # Store a message if the user wants us to
    if store_message(message, meshtasticClient):
        return True

    # Handle other commands
    match message.text.upper():
        case "INBOX":
            send_inbox(message.fromNode)
        case "NEW":
            send_new_messages(message.fromNode)
        case "OLD":
            send_old_messages(message.fromNode)
        case "CLEAR":
            clear_old_messages(message.fromNode)
        case "NODES":
            message.reply(
                f"🤖👀 I've seen these nodes:\n\n{meshtasticClient.nodelist().to_succinct_string()}"
            )
        case _:
            return False

    return True


def notify_user(message: Message, meshtasticClient: MeshtasticClient) -> bool:
    """
    Check to see if one of our recipients came in range, and has new messages.
    """

    # If they are messaging us first, they will probably quickly find out that
    # they have messages, and it just breaks the flow. So only check for all
    # other message types.
    if message.type == "TEXT_MESSAGE_APP" and message.toNode.is_self():
        return False

    # We get routing messages for each Ack, so ignore those or we get a royal
    # clusterfuck.
    if message.type == "ROUTING_APP":
        return False

    # Send this user their new messages
    messages = messageStore.get(message.fromNode.id, [])
    numUnread = sum(1 for msg in messages if not msg["read"])
    return send_messages(
        message.fromNode,
        f"🤖📬 I have {numUnread} new {'message' if numUnread == 1 else 'messages'} for you! Sending {'it' if numUnread == 1 else 'them'} now...",
        read=False,
    )


def send_messages(node: Node, intro_text: str, read: bool = False) -> bool:
    if user_stats(node)[("numRead" if read else "numUnread")] == 0:
        return False

    node.send(intro_text)
    for msg in messageStore.get(node.id, []):
        if msg["read"] != read:
            continue
        msg["read"] = True
        node.send(
            f"🤖✉️ From {msg['sender']}, {time_ago(msg['timestamp'])} ago:\n\n{msg['contents']}"
        )

    return True


def store_message(message: Message, meshtasticClient: MeshtasticClient) -> bool:
    """
    Store new messages when requested by the user
    """
    if not message.text.upper().startswith("SEND"):
        return False

    parts = message.text.split(" ")
    msg = " ".join(parts[2:])

    if len(msg) == 0:
        message.fromNode.send("🤖🧨 I'm sorry, I can't send an empty message.")
        return True

    # Figure out who the recipient is
    id = parts[1]
    recipientId = meshtasticClient.nodelist().find_id(id)
    if not recipientId:
        message.fromNode.send(
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
    message.fromNode.send(f"🤖📨 Saved this message for node `{id}`:\n\n{msg}")
    return True


def store_welcome_message(node: Node):
    if node.id not in messageStore:
        messageStore[node.id] = [
            {
                "sender": "🤖 Meshbot",
                "contents": f"Welcome to this Meshtastic answering machine, {node.longName}! You can leave messages for other users, and they can leave messages for you! Hope you like it 😄",
                "read": False,
                "timestamp": datetime.now(),
            },
        ]


def user_stats(node: Node) -> dict:
    messages = messageStore.get(node.id, [])
    numUnread = sum(1 for msg in messages if not msg["read"])
    totalMessages = len(messages)
    return {
        "totalMessages": totalMessages,
        "numUnread": numUnread,
        "numRead": totalMessages - numUnread,
    }


def send_inbox(node: Node):
    stats = user_stats(node)
    if stats["totalMessages"] == 0:
        node.send("🤖📭 You have no messages in your inbox")
        return

    icon = "📬" if stats["numUnread"] > 0 else "📭"
    node.send(
        f"🤖{icon} You have {stats['numUnread']} unread {pluralize('message', stats['numUnread'])}, and a grand total of {stats['totalMessages']} {pluralize('message', stats['totalMessages'])} in your inbox. Send `NEW` or `OLD` to fetch your messages."
    )


def send_new_messages(node: Node):
    stats = user_stats(node)
    if not send_messages(
        node,
        f"🤖📬 You have {stats['numUnread']} new {pluralize('message', stats['numUnread'])}. Sending {pluralize('it', stats['numUnread'])} now...",
        read=False,
    ):
        old_messages = (
            " Send `OLD` to read your older messages." if stats["numRead"] > 0 else ""
        )
        node.send(f"🤖📭 You have no new messages.{old_messages}")


def send_old_messages(node: Node):
    stats = user_stats(node)
    if not send_messages(
        node,
        f"🤖📬 You have {stats['numRead']} old {pluralize('message', stats['numRead'])}. Sending {pluralize('it', stats['numRead'])} now...",
        read=True,
    ):
        new_messages = (
            " Send `NEW` to read your new messages." if stats["numUnread"] > 0 else ""
        )
        node.send(f"🤖📭 You have no old messages.{new_messages}")


def clear_old_messages(node: Node):
    stats = user_stats(node)
    messageStore[node.id] = [msg for msg in messageStore[node.id] if not msg["read"]]
    node.send(
        f"🤖🗑️ I removed {stats['numRead']} old {pluralize('message', stats['numRead'])}. You have {stats['numUnread']} new {pluralize('message', stats['numUnread'])} left in your inbox."
    )


def pluralize(word: str, count: int) -> str:
    if count == 1:
        return word
    if word == "it":
        return "them"
    return word + "s"
