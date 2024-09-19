from .time_helper import time_ago


class Node:
    """Class representing a Meshtastic node in the LoRa mesh"""

    def __init__(self, data, interface):
        self.data = data
        self.interface = interface
        self.isSelf = False

        self.num = data.get("num")
        if not self.num:
            print("ERROR: Node has no ID")
            raise Exception("A node should have an ID")

        self.id = data.get("user", {}).get("id", "")
        self.mac = data.get("user", {}).get("macaddr", "")
        self.hardware = data.get("user", {}).get("hwModel", "")
        self.shortName = data.get("user", {}).get("shortName", "")
        self.longName = data.get("user", {}).get("longName", "")
        self.lastHeard = data.get("lastHeard", 0)
        self.hopsAway = data.get("hopsAway", 0)
        self.snr = data.get("snr", None) if self.hopsAway == 0 else None

    def mark_as_self(self):
        self.isSelf = True

    def send(self, message: str, **kwargs):
        if self.id:
            print(f"Sending to {self}: {message}")
            self.interface.sendText(message, destinationId=self.id, **kwargs)
            return True
        else:
            return False

    def __str__(self):
        if self.isSelf:
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
        snr = f", SNR {self.snr:.2f}" if self.snr else ""
        hops = f", {self.hopsAway} {"hop" if self.hopsAway == 1 else "hops"} away" if self.hopsAway > 0 else ""
        return f"{str(self)} \033[90m({hardware}last heard {time_ago(self.lastHeard)} ago{snr}{hops})\033[0m"

    def to_succinct_string(self):
        """Use when indentifying this node in Meshtastic messages"""
        return f"[{self.shortName}] {self.longName} ({self.id})"
