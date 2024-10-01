import requests
import os
from dotenv import dotenv_values

from .meshwrapper import Message, Nodelist
from .chatbot import Chatbot

config = {
    **dotenv_values(".env"),
    **dotenv_values("production.env"),
    **dotenv_values("development.env"),
    **os.environ,
}


def register(bot: Chatbot):
    if not ("OLLAMA_API" in config and "OLLAMA_MODEL" in config):
        return

    bot.add_state("LLM")

    bot.add_command(
        {
            "command": "/LLM",
            "module": "Ollama LLM",
            "description": "Start AI conversation",
            "function": start_conversation,
            "channel": True,
        },
        {
            "state": "LLM",
            "command": Chatbot.CATCH_ALL_TEXT,
            "module": "Ollama LLM",
            "function": converse,
            "channel": True,
        },
        {
            "state": "LLM",
            "command": ["/STOP", "/EXIT"],
            "module": "Ollama LLM",
            "description": "End conversation",
            "function": stop_conversation,
            "channel": True,
        },
    )


conversations = {}


def start_conversation(message: Message) -> str:
    message.reply("🤖⏳ Spinning up the LLM, just a moment...")
    conversations[identifier(message)] = [
        {
            "role": "system",
            "content": system_prompt + str(gather_relevant_stats(message)),
        }
    ]
    reply = reply_from_ollama(conversations[identifier(message)], message.nodelist)
    message.reply("🤖🧠 Started LLM conversation")
    reply_if_not_empty(message, reply)
    return "LLM"


def converse(message: Message):
    assert identifier(message) in conversations, "Conversation should have been started"
    conversations[identifier(message)].append(
        {
            "role": "user",
            "content": f"Node {message.fromNode.id}: {message.text}",
        }
    )
    reply_if_not_empty(
        message, reply_from_ollama(conversations[identifier(message)], message.nodelist)
    )


def stop_conversation(message: Message) -> str:
    del conversations[identifier(message)]
    message.reply("🤖🧠 Ended LLM conversation")
    return "MAIN"


def identifier(message: Message) -> str:
    if message.private_message():
        return message.fromNode.id
    else:
        return f"Channel {message.channel}"


def reply_if_not_empty(message: Message, reply: str):
    if reply != "":
        conversations[identifier(message)].append(
            {"role": "assistant", "content": reply}
        )
        message.reply("🤖 " + reply)


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

If you are in a channel (a group chat) and you think you can't answer the
question or you think that you are not being addressed, just say nothing. If you
are talking one-on-one you are always expected to reply.

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


def reply_from_ollama(conversation: list, nodelist: Nodelist):
    request = {
        "model": config["OLLAMA_MODEL"],
        "messages": conversation,
        "stream": False,
    }

    if config.get("OLLAMA_USE_TOOLS") == "True":
        request["tools"] = tools

    working = True
    while working:
        try:
            result = requests.post(config["OLLAMA_API"] + "/chat", json=request)
        except requests.exceptions.ConnectionError as err:
            return f"Could not reach the Ollama server at this time: {err}"
        if not result.ok:
            return f"Did not get a valid result from Ollama. Status: {result.status_code} - {result.text}"

        result = result.json()
        tool_calls = result.get("message", {}).get("tool_calls", False)

        if tool_calls:
            # print("Tool call! " + str(tool_calls))
            for call in tool_calls:
                function = call.get("function", {})
                arguments = function.get("arguments", {})
                node = nodelist.find(arguments.get("node", ""))
                assert node, "The tool should have been called with a node parameter"
                match function.get("name", None):
                    case "get_signal_strength":
                        return_value = f"Node {node.to_succinct_string()} is being received with an SNR of {node.snr} and an RSSI of {node.rssi}"
                    case "get_hops":
                        return_value = f"Node {node.to_succinct_string()} is {node.hopsAway} hops away"
                    case _:
                        assert False, "Invalid function name in function call from LLM"
                # print("Return value: " + return_value)
                conversation.append({"role": "tool", "content": return_value})
        else:
            working = False

    return result.get("message", {}).get(
        "content", "Did not get a valid result from Ollama :/"
    )


def gather_relevant_stats(message: Message) -> dict:
    meshbot_node = next(
        (n for n in message.nodelist.nodes.values() if n.is_self()), None
    )
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
            for node in message.nodelist.nodes.values()
        ]
    return stats
