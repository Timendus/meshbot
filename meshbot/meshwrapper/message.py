from datetime import datetime
from .node import Node, Everyone


class Message:
    """Class representing a message that was received over the LoRa mesh"""

    def __init__(self):
        pass

    @staticmethod
    def from_packet(data):
        message = Message()
        message.data = data

        message.id = data.get("id")
        message.channel = int(data.get("channel", 0))
        message.timestamp = datetime.fromtimestamp(data.get("rxTime", 0))
        message.type = data.get("decoded", {}).get("portnum")
        message.text = data.get("decoded", {}).get("text", "")

        message.telemetry = data.get("decoded", {}).get("telemetry", {})
        if "raw" in message.telemetry:
            del message.telemetry["raw"]

        position = data.get("decoded", {}).get("position", None)
        message.position_request = data.get("decoded", {}).get("wantResponse", False)
        if position and not message.position_request:
            message.position = [
                position.get("latitudeI", 0) / pow(10, 7),
                position.get("longitudeI", 0) / pow(10, 7),
                position.get("altitude", 0),
            ]

        message.neighborInfo = data.get("decoded", {}).get("neighborinfo", {})
        if "raw" in message.neighborInfo:
            del message.neighborInfo["raw"]

        message.user = data.get("decoded", {}).get("user", {})
        if "raw" in message.user:
            del message.user["raw"]

        message.routing = data.get("decoded", {}).get("routing", {})
        if "raw" in message.routing:
            del message.routing["raw"]

        message.admin = data.get("decoded", {}).get("admin", {})
        if "raw" in message.admin:
            del message.admin["raw"]

    def private_message(self):
        return self.toNode != Everyone

    def reply(self, message: str, **kwargs) -> bool:
        if self.toNode == Everyone:
            # This was a message in a channel, respond in the same channel
            return Everyone.send(message, channelIndex=self.channel, **kwargs)
        if self.fromNode:
            # This was a direct message, respond to the right node
            return self.fromNode.send(message, channelIndex=self.channel, **kwargs)
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
                if self.position_request:
                    content = f"position request"
                else:
                    content = f"updated location to {self.position}"
            case "NEIGHBORINFO_APP":
                content = f"I'm seeing these neighbours: {self.neighborInfo}"
            case "NODEINFO_APP":
                content = f"updated node info to: {self.user}"
            case "ROUTING_APP":
                content = f"new routing info: {self.routing}"
            case "ADMIN_APP":
                content = f"administrating: {self.admin}"
            case "TRACEROUTE_APP":
                content = f"traceroute request"

        return f"{self.fromNode} --> {self.toNode}: {content}"
