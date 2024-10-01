import logging
import textwrap
import time
from threading import Timer

from .time_helper import time_ago

logger = logging.getLogger("Meshbot")

# Message reply timeout delay before we give up
MAX_REPLY_DELAY = 5

# Maximum size of a message in UTF-8 bytes that we can send
MAX_SIZE = 234

# Size minus `len(" [i/n]")`.
# Note: if we have to split into more than 9 messages, this does break.
SHORT_SIZE = MAX_SIZE - 6

wrapper = textwrap.TextWrapper(
    width=SHORT_SIZE, replace_whitespace=False, break_long_words=True
)


class Node:
    """Class representing a Meshtastic node in the LoRa mesh"""

    def __init__(self):
        self.transmission = {"sending": None, "last_result": False, "timeout": None}

    @staticmethod
    def from_packet(data, interface):
        node = Node()
        node.data = data
        node.interface = interface

        node.num = data.get("num")
        assert node.num, "Node should at least have an ID"

        node.id = data.get("user", {}).get("id", "")
        node.mac = data.get("user", {}).get("macaddr", "")
        node.hardware = data.get("user", {}).get("hwModel", "")
        node.role = data.get("user", {}).get("role", None)
        node.shortName = data.get("user", {}).get("shortName", "")
        node.longName = data.get("user", {}).get("longName", "")
        if not node.mac:
            node.shortName = "UNKN"
            node.longName = "Unknown node"

        node.lastHeard = data.get("lastHeard", 0)
        node.hopsAway = data.get("hopsAway", 0)
        node.snr = data.get("snr", None)
        node.rssi = None

        return node

    def is_self(self):
        return (
            hasattr(self, "interface")
            and hasattr(self.interface, "myInfo")
            and hasattr(self.interface.myInfo, "my_node_num")
            and self.num == self.interface.myInfo.my_node_num
        )

    def is_broadcast(self):
        return self is Everyone

    def send(self, message: str, **kwargs) -> bool:
        if self.id and self.interface:
            messages = self.break_message(message)
            oneliner = message.replace("\n", "\\n")
            logger.info(
                f"Sending to {self} in {len(messages)} {'part' if len(messages) == 1 else 'parts'}: {oneliner}"
            )
            for msg in messages:
                if not self._send(msg, **kwargs):
                    return False
            return True
        else:
            return False

    def _send(self, message: str, **kwargs) -> bool:
        self.transmission["timeout"] = Timer(MAX_REPLY_DELAY, self.on_timeout)
        self.transmission["timeout"].start()
        self.transmission["sending"] = self.interface.sendText(
            message,
            destinationId=self.id,
            wantAck=True,
            onResponse=self.onAckNak,
            **kwargs,
        )
        while self.transmission["sending"]:
            time.sleep(0.1)
        return self.transmission["last_result"]

    # Don't change the name of this callback
    # https://github.com/meshtastic/python/blob/c696d59b9052361856630c8eb97a061cdb51dc6b/meshtastic/mesh_interface.py#L415-L418
    def onAckNak(self, response):
        if (
            self.transmission["sending"]
            and self.transmission["sending"].id == response["decoded"]["requestId"]
        ):
            # Got a reply to the blocking message! Unblocking...
            self.transmission["timeout"].cancel()
            self.transmission["last_result"] = (
                response["decoded"]["routing"]["errorReason"] == "NONE"
            )
            self.transmission["sending"] = None

    def on_timeout(self):
        logger.info(
            f"Did not get a reply from {self} within {MAX_REPLY_DELAY} seconds, moving on"
        )
        self.transmission["last_result"] = False
        self.transmission["sending"] = None

    def break_message(self, message: str):
        # Keep it as a single message if possible
        if len(message.encode("utf-8")) <= MAX_SIZE:
            return [message]

        # Split message into multiple parts
        words = wrapper._split_chunks(message)
        words.reverse()  # use it as a stack
        words = [w.encode("utf-8") for w in words]
        lines = [b""]
        while words:
            word = words.pop(-1)
            if len(word) > SHORT_SIZE:
                assert False, "we should never be here if the wrapper does its job"
            if len(lines[-1]) + len(word) <= SHORT_SIZE:
                lines[-1] += word
            else:
                lines.append(word)
        return [
            f"{l.decode().rstrip()} [{i+1}/{len(lines)}]" for i, l in enumerate(lines)
        ]

    def __str__(self):
        if type(self) == SpecialNode:
            color = "95"
        elif self.is_self():
            color = "92"
        elif self.hopsAway == 0:
            color = "96"
        else:
            color = "94"

        if len(self.shortName) == 1 and len(self.shortName.encode("utf-8")) == 4:
            # Short name is an emoji
            shortName = f" {self.shortName} "
        else:
            shortName = self.shortName.ljust(4)

        return f"\033[{color}m[{shortName}] {self.longName}\033[0m"

    def to_verbose_string(self):
        """Used when stringifying a Nodelist"""
        hardware = f"{self.hardware}, " if self.hardware != "UNSET" else ""
        role = f"{self.role}, " if self.role else ""
        snr = f", SNR {self.snr:.2f}" if self.snr else ""
        rssi = f", RSSI {self.rssi:.2f}" if self.rssi else ""
        hops = (
            f", {self.hopsAway} {'hop' if self.hopsAway == 1 else 'hops'} away"
            if self.hopsAway > 0
            else ""
        )
        return f"{str(self)} \033[90m({hardware}{role}last heard {time_ago(self.lastHeard)} ago{snr}{rssi}{hops})\033[0m"

    def to_succinct_string(self):
        """Use when indentifying this node in Meshtastic messages"""
        return f"[{self.shortName}] {self.longName} ({self.id})"


class SpecialNode(Node):
    def __init__(self, short, long, id):
        self.shortName = short
        self.longName = long
        self.id = id
        self.interface = None
        self.hardware = "UNSET"
        self.transmission = {"sending": None, "last_result": False, "timeout": None}

    def is_self(self):
        return False


Everyone = SpecialNode("CAST", "Everyone", 0xFFFFFFFF)
Unknown = SpecialNode("UNKN", "Unknown", 0x00000000)
