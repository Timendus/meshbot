import requests
import os
from dotenv import dotenv_values

from .meshwrapper import MeshtasticClient, Message

config = {
    **dotenv_values(".env"),
    **dotenv_values("production.env"),
    **dotenv_values("development.env"),
    **os.environ,
}

system_prompt = f"""
You are Meshbot, a helpful chatbot on the Meshtastic network. You talk a bit
like a radio HAM. The primary language of this part of the network is
{config["OLLAMA_LANGUAGE"]}. But users can talk to you in any language. Just be
kind and reply in the same language.

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

tools = [
    {
        "type": "function",
        "function": {
            "name": "get_signal_strength",
            "description": "Get the signal strength for a node",
            "parameters": {
                "type": "object",
                "properties": {
                    "node": {
                        "type": "string",
                        "description": "The ID or short name of the node for which to get the signal strength, e.g. !9a34ed2b or R3NL",
                    },
                },
                "required": ["node"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "get_hops",
            "description": "Get the number of hops for a node",
            "parameters": {
                "type": "object",
                "properties": {
                    "node": {
                        "type": "string",
                        "description": "The ID or short name of the node for which to get the number of hops, e.g. !9a34ed2b or R3NL",
                    },
                },
                "required": ["node"],
            },
        },
    },
]

conversations = {}
client = None


def handle(message: Message, meshtasticClient: MeshtasticClient) -> bool:
    # Store for later use
    client = meshtasticClient

    # Only reply to text messages
    if message.type != "TEXT_MESSAGE_APP":
        return False

    # Only reply if the required Ollama settings have been configured
    if not (
        "OLLAMA_API" in config
        and "OLLAMA_MODEL" in config
        and "OLLAMA_LANGUAGE" in config
    ):
        return False

    # Check if model is available
    # models = requests.get(config["OLLAMA_API"] + "/tags").text
    # print(models)

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
        conversations[identifier].append(
            {
                "role": "user",
                "content": f"Node {message.fromNode.id} says: {message.text}",
            }
        )
        reply = reply_from_ollama(conversations[identifier])
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
        reply = reply_from_ollama(conversations[identifier])
        message.reply("🤖🧠 Started LLM conversation")
        if reply != "" and reply != "NOTHING":
            message.reply("🤖 " + reply)
        return True

    return False


def reply_from_ollama(conversation: list):
    request = {
        "model": config["OLLAMA_MODEL"],
        "messages": conversation,
        "stream": False,
        "tools": tools,
    }

    working = True
    while working:
        try:
            result = requests.post(config["OLLAMA_API"] + "/chat", json=request)
        except ConnectionError:
            return f"Could not reach the Ollama server at this time."
        if not result.ok:
            return f"Did not get a valid result from Ollama :/ Status: {result.status_code}"

        result = result.json()
        tool_calls = result.get("message", {}).get("tool_calls", False)

        if tool_calls:
            print("tool call!")
            print(tool_calls)
            for call in tool_calls:
                function = call.get("function", {})
                arguments = function.get("arguments", {})
                match function.get("name", None):
                    case "get_signal_strength":
                        return_value = get_signal_strength(**arguments)
                    case "get_hops":
                        return_value = get_hops(**arguments)
                    case _:
                        assert False, "Invalid function name in function call from LLM"
                conversation.append({"role": "tool", "content": return_value})

            print(request)
        else:
            working = False

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
            "shortName": message.fromNode.shortName,
            "longName": message.fromNode.longName,
            "id": message.fromNode.id,
        }
    else:
        stats["users"] = [
            {
                "shortName": node.shortName,
                "longName": node.longName,
                "id": node.id,
            }
            for node in meshtasticClient.nodelist().nodes.values()
        ]
    return stats


def get_signal_strength(node: str) -> str:
    node = client.nodelist().find(node)
    return str(
        {
            "snr": node.snr,
            "rssi": node.rssi,
        }
    )


def get_hops(node: str) -> str:
    node = client.nodelist().find(node)
    return str({"hopsAway": node.hopsAway})
