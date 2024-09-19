from meshwrapper import MeshtasticClient, Message


def handle(message: Message, meshtasticClient: MeshtasticClient) -> bool:
    # Only reply to text messages
    if message.type != "TEXT_MESSAGE_APP":
        return False
    # Only reply if the message is an explicit signal command
    if not message.text.upper().startswith("/SIGNAL"):
        return False

    # Figure out who we're requesting a signal report about
    parts = message.text.split(" ")
    if len(parts) == 1:
        # Send a signal report on the sender
        subject = message.fromNode
    else:
        # Send a signal report on the specified node
        subject = meshtasticClient.nodeList.find(" ".join(parts[1:]))

    if not subject:
        message.reply(
            "🤖🧨 I don't know who that is. Sorry!\n\nI need the short name (example: TDRP), or node ID (example: !8e92a31f) of a node that I know."
        )
        return True

    message.reply(
        f"🤖📶 I'm reading {subject.to_succinct_string()} with an SNR of {subject.snr}"
    )
    return True
