from .meshwrapper import MeshtasticClient, Message


def handle(message: Message, meshtasticClient: MeshtasticClient) -> bool:
    # Only reply to text messages
    if message.type != "TEXT_MESSAGE_APP":
        return False

    if message.text.upper().startswith("/SIGNAL"):
        signal_report(message, meshtasticClient)
        return True

    match message.text.upper():
        case "/NODES":
            nodes_info(message, meshtasticClient)
        case "/NODELIST":
            node_list(message, meshtasticClient)
        case _:
            return False

    return True


def signal_report(message: Message, meshtasticClient: MeshtasticClient):
    # Figure out who we're requesting a signal report about
    parts = message.text.split(" ")
    if len(parts) == 1:
        # Send a signal report on the sender
        subject = message.fromNode
    else:
        # Send a signal report on the specified node
        subject = meshtasticClient.nodelist().find(" ".join(parts[1:]))

    if not subject:
        message.reply(
            "🤖🧨 I don't know who that is. Sorry!\n\nI need the short name (example: TDRP), or node ID (example: !8e92a31f) of a node that I know."
        )
        return

    if subject.hopsAway == 0:
        if subject.snr and subject.rssi:
            message.reply(
                f"🤖📶 I'm reading {subject.to_succinct_string()} with an SNR of {subject.snr} and an RSSI of {subject.rssi}."
            )
        elif subject.snr:
            message.reply(
                f"🤖📶 I'm reading {subject.to_succinct_string()} with an SNR of {subject.snr}."
            )
        elif subject.rssi:
            message.reply(
                f"🤖📶 I'm reading {subject.to_succinct_string()} with an RSSI of {subject.rssi}."
            )
        else:
            message.reply(
                f"🤖📶 I don't have any readings for {subject.to_succinct_string()}."
            )
    else:
        rssi = f" and an RSSI of {subject.rssi}" if subject.rssi else ""
        snr = (
            f", with an SNR of {subject.snr}{rssi} on the last hop"
            if subject.snr
            else ""
        )
        message.reply(
            f"🤖📶 {subject.to_succinct_string()} is {subject.hopsAway} {'hop' if subject.hopsAway == 1 else 'hops'} away{snr}."
        )


def nodes_info(message: Message, meshtasticClient: MeshtasticClient):
    message.reply(f"🤖📡 Nodes report!\n\n{meshtasticClient.nodelist().summary()}")


def node_list(message: Message, meshtasticClient: MeshtasticClient):
    message.reply(
        f"🤖👀 I've seen these nodes:\n\n{meshtasticClient.nodelist().to_succinct_string()}"
    )
