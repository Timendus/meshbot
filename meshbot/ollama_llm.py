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
like a radio HAM. Users can talk to you in any language. Just be kind and reply
in the same language if you can.

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

Messages will be prepended with "Node XXXXXX says:". This is for your
convenience only, so you can see who said what. This is not part of the original
message. You should not prepend a similar message to your replies.

If you are in a channel (a group chat) and you think you can't answer the
question or you are quite sure you are not being addressed, just say nothing by
answering with the string "NOTHING". Only "NOTHING", no more. If you are talking
one-on-one you are always expected to reply.

Do not hallucinate things, only use the information below and the available
tools/functions that you can call when answering radio reception specific
questions. Otherwise just answer that you do not know, or that you do not know
what to say. Feel free to talk generally about unrelated topics when asked.

Information:

"""

tools = [
    {
        "type": "function",
        "function": {
            "name": "get_signal_strength",
            "description": "Get the signal strength for a node, in SNR and RSSI if available",
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
            "description": "Get the number of hops to reach a node in the mesh network",
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


def handle(message: Message, meshtasticClient: MeshtasticClient) -> bool:
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
        reply = reply_from_ollama(conversations[identifier], meshtasticClient)
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
        reply = reply_from_ollama(conversations[identifier], meshtasticClient)
        message.reply("🤖🧠 Started LLM conversation")
        if reply != "" and reply != "NOTHING":
            message.reply("🤖 " + reply)
        return True

    return False


def reply_from_ollama(conversation: list, meshtasticClient: MeshtasticClient):
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
        except requests.exceptions.ConnectionError:
            return f"Could not reach the Ollama server at this time."
        if not result.ok:
            return f"Did not get a valid result from Ollama :/ Status: {result.status_code} - {result.text}"

        result = result.json()
        tool_calls = result.get("message", {}).get("tool_calls", False)

        if tool_calls:
            print("Tool call! " + str(tool_calls))
            for call in tool_calls:
                function = call.get("function", {})
                arguments = function.get("arguments", {})
                node = meshtasticClient.nodelist().find(arguments.get("node", ""))
                assert node, "The tool should have been called with a node parameter"
                match function.get("name", None):
                    case "get_signal_strength":
                        return_value = f"Node {node.to_succinct_string()} is being received with an SNR of {node.snr} and an RSSI of {node.rssi}"
                    case "get_hops":
                        return_value = f"Node {node.to_succinct_string()} is {node.hopsAway} hops away"
                    case _:
                        assert False, "Invalid function name in function call from LLM"
                print("Return value: " + return_value)
                conversation.append({"role": "tool", "content": return_value})
        else:
            working = False

    return result.get("message", {}).get(
        "content", "Did not get a valid result from Ollama :/"
    )


def gather_relevant_stats(message: Message, meshtasticClient: MeshtasticClient) -> dict:
    nodelist = meshtasticClient.nodelist()
    meshbot_node = next((n for n in nodelist.nodes.values() if n.is_self()), None)
    stats = {
        "in_channel": not message.private_message(),
        "meshbot": {
            "shortName": meshbot_node.shortName,
            "longName": meshbot_node.longName,
            "id": meshbot_node.id,
        },
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
            for node in nodelist.nodes.values()
        ]
    return stats
