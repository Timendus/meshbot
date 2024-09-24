import logging
import textwrap

from .time_helper import time_ago

logger = logging.getLogger("Meshbot")

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

    def __init__(self, data, interface):
        self.data = data
        self.interface = interface
        self.isSelf = False

        self.num = data.get("num")
        if not self.num:
            logger.error("ERROR: Node has no ID")
            raise Exception("A node should have an ID")

        self.id = data.get("user", {}).get("id", "")
        self.mac = data.get("user", {}).get("macaddr", "")
        self.hardware = data.get("user", {}).get("hwModel", "")
        self.shortName = (
            data.get("user", {}).get("shortName", "") if self.mac else "UNKN"
        )
        self.longName = (
            data.get("user", {}).get("longName", "") if self.mac else "Unknown node"
        )
        self.role = data.get("user", {}).get("role", None)
        self.lastHeard = data.get("lastHeard", 0)
        self.hopsAway = data.get("hopsAway", 0)
        self.snr = data.get("snr", None)

    def mark_as_self(self):
        self.isSelf = True

    def send(self, message: str, **kwargs):
        if self.id and self.interface:
            logger.info(f"Sending to {self}: {message}")
            for msg in self.break_message(message):
                self.interface.sendText(msg, destinationId=self.id, **kwargs)
            return True
        else:
            return False

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
        elif self.isSelf:
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
        hops = (
            f", {self.hopsAway} {'hop' if self.hopsAway == 1 else 'hops'} away"
            if self.hopsAway > 0
            else ""
        )
        return f"{str(self)} \033[90m({hardware}{role}last heard {time_ago(self.lastHeard)} ago{snr}{hops})\033[0m"

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
        self.isSelf = False

    def mark_as_self(self):
        pass


Everyone = SpecialNode("CAST", "Everyone", 0xFFFFFFFF)
Unknown = SpecialNode("UNKN", "Unknown", 0x00000000)
