from datetime import datetime
from .node import Node


class Message:
    """Class representing a message that was received over the LoRa mesh"""

    def __init__(self, data, fromNode: Node, toNode: Node):
        self.data = data

        self.id = data.get("id")
        self.timestamp = datetime.fromtimestamp(data.get("rxTime", 0))
        self.type = self.data.get("decoded", {}).get("portnum")
        self.text = self.data.get("decoded", {}).get("text", "")

        self.telemetry = self.data.get("decoded", {}).get("telemetry", {})
        if "raw" in self.telemetry:
            del self.telemetry["raw"]

        position = self.data.get("decoded", {}).get("position", None)
        if position:
            self.position = [
                position["latitudeI"] / pow(10, 7),
                position["longitudeI"] / pow(10, 7),
                position["altitude"] if "altitude" in position else 0,
            ]

        self.neighborInfo = self.data.get("decoded", {}).get("neighborinfo", {})
        if "raw" in self.neighborInfo:
            del self.neighborInfo["raw"]

        self.user = self.data.get("decoded", {}).get("user", {})
        if "raw" in self.user:
            del self.user["raw"]

        self.routing = self.data.get("decoded", {}).get("routing", {})
        if "raw" in self.routing:
            del self.routing["raw"]

        self.fromNode = fromNode
        if toNode:
            self.toNode = toNode
        elif data.get("to") == 0xFFFFFFFF:
            self.toNode = "\033[95mEveryone\033[0m"
        else:
            self.toNode = "\033[32mUnknown\033[0m"

    def reply(self, message: str, **kwargs):
        if self.fromNode:
            return self.fromNode.send(message, **kwargs)
        else:
            return False

    def __str__(self):
        content = str(self.data)
        match self.type:
            case "TELEMETRY_APP":
                content = f"new telemetry: {self.telemetry}"
            case "TEXT_MESSAGE_APP":
                content = self.text
            case "POSITION_APP":
                content = f"updated location to {self.position}"
            case "NEIGHBORINFO_APP":
                content = f"I'm seeing these neighbours: {self.neighborInfo}"
            case "NODEINFO_APP":
                content = f"updated node info to: {self.user}"
            case "ROUTING_APP":
                content = f"new routing info: {self.routing}"

        return f"{self.fromNode} --> {self.toNode}: {content}"
