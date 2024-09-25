import requests

from .meshwrapper import MeshtasticClient, Message

system_prompt = """
You are a helpful chatbot on the Meshtastic network. You talk a bit like a radio
HAM, but not too much. The primary language of this part of the network is
Dutch. But users can talk to you in any language. Just be kind and reply in the
same language.

Remember that Meshtastic is unlicensed and does not use call signs. Also
remember that Meshtastic uses the LoRa protocol, which can work reliably with
very noisy messages. Typical LoRa SNR values are between -20 dB and +10 dB. RSSI
can go all the way down to -128 and rarely gets above -30. Messages can hop
through the mesh via other nodes.

These icons are often used in long names of nodes:

🏠 - Base node
📟 - Mobile node
✈ - Node on board of a plane
🎈 - Node carried by a balloon
☀️ - Solar powered node
🔌 - Net powered node
🌐 - Node connected to MQTT (for sharing locations and passing messages)
🕐 - Node using a yagi antenna
🛰️ - Node with GPS/GNSS on board

Keep your replies polite and friendly, but short and to the point, since
bandwidth is very limited. Preferably under 234 characters, so they can be
transmitted in a single packet.

If you are in a channel (a group chat) and you think you can't answer the
question or you are quite sure you are not being addressed, just say nothing by
answering with the string "NOTHING". Only "NOTHING", no more. If you are talking
one-on-one you are always expected to reply.

Do not hallucinate things, only use the information below when answering
specific questions. Otherwise just answer that you do not know, or that you do
not know what to say. Feel free to talk generally about unrelated topics when
asked.

Information:

"""

conversations = {}


def handle(
    message: Message, meshtasticClient: MeshtasticClient, endpoint: str, model: str
) -> bool:
    # Only reply to text messages
    if message.type != "TEXT_MESSAGE_APP":
        return False

    # Only reply if an endpoint and a model have been configured
    if not endpoint or not model:
        return False

    if message.private_message():
        identifier = message.fromNode.id
    else:
        identifier = f"Channel {message.channel}"

    # Stopping conversations
    if identifier in conversations and message.text.upper() in [
        "STOP",
        "QUIT",
        "EXIT",
        "/STOP",
        "/QUIT",
        "/EXIT",
    ]:
        del conversations[identifier]
        message.reply("🤖🧠 Ended LLM conversation")
        return True

    # Are we in a conversation?
    if identifier in conversations:
        conversations[identifier].append({"role": "user", "content": message.text})
        reply = reply_from_ollama(conversations[identifier], endpoint, model)
        if reply != "" and reply != "NOTHING":
            message.reply("🤖 " + reply)
        return True

    # Starting conversations
    if message.text.upper() == "/LLM":
        conversations[identifier] = [
            {
                "role": "system",
                "content": system_prompt
                + str(gather_relevant_stats(message, meshtasticClient)),
            }
        ]
        reply = reply_from_ollama(conversations[identifier], endpoint, model)
        message.reply("🤖🧠 Started LLM conversation")
        if reply != "" and reply != "NOTHING":
            message.reply("🤖 " + reply)
        return True

    return False


def reply_from_ollama(conversation: list, endpoint: str, model: str):
    request = {
        "model": model,
        "messages": conversation,
        "stream": False,
    }

    try:
        result = requests.post(endpoint + "/chat", json=request)
    except ConnectionError:
        return f"Could not reach the Ollama server at this time."
    if not result.ok:
        return f"Did not get a valid result from Ollama :/ Status: {result.status}"

    return (
        result.json()
        .get("message", {})
        .get("content", "Did not get a valid result from Ollama :/")
    )


def gather_relevant_stats(message: Message, meshtasticClient: MeshtasticClient) -> dict:
    stats = {
        "in_channel": not message.private_message(),
    }
    if message.private_message():
        stats["user"] = {
            "rssi": message.fromNode.rssi,
            "snr": message.fromNode.snr,
            "shortName": message.fromNode.shortName,
            "longName": message.fromNode.longName,
            "id": message.fromNode.id,
            "hops": message.fromNode.hopsAway,
        }
    else:
        stats["users"] = [
            {
                "rssi": node.rssi,
                "snr": node.snr,
                "shortName": node.shortName,
                "longName": node.longName,
                "id": node.id,
                "hops": node.hopsAway,
            }
            for node in meshtasticClient.nodelist().nodes.values()
        ]
    return stats
