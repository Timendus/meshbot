import re

from .node import Node, Everyone, Unknown


fullHexId = re.compile("![0-9a-fA-F]{8}")
shortHexId = re.compile("[0-9a-fA-F]{8}")


class Nodelist:
    """Class representing a collection of Meshtastic nodes"""

    def __init__(self):
        self.nodes = {}

    def add(self, node: Node):
        self.nodes[node.num] = node

    def update(self, node: Node):
        self.nodes[node.num] = node

    def get(self, num) -> Node | None:
        """Returns Node object"""
        if num == 0xFFFFFFFF:
            return Everyone
        elif num in self.nodes.keys():
            return self.nodes[num]
        return Unknown

    def find(self, needle: str) -> Node | None:
        """Figure out which node the user intends. Returns Node object or None"""
        id = self.find_id(needle)
        if id:
            return self.nodes.get(int(id[1:], 16), None)
        else:
            return None

    def find_id(self, needle: str) -> str | None:
        """Figure out which node the user intends. Returns full HEX id string or None"""

        if fullHexId.match(needle):
            # needle is a HEX notation node number
            return needle

        elif shortHexId.match(needle):
            # needle is a HEX notation node number, but we're missing the exclamation mark
            return f"!{needle}"

        elif len(needle) <= 4 and needle.upper() in [
            node.shortName.upper() for node in self.nodes.values()
        ]:
            # needle is a known short name
            return next(
                node.id
                for node in self.nodes.values()
                if node.shortName.upper() == needle.upper()
            )

        elif needle.isnumeric() and int(needle) > 0:
            # needle is a decimal number
            return "!" + hex(needle)[2:]

        return None

    def __str__(self):
        output = "Node list\n"
        output += "---------\n"
        nodes = sorted(self.nodes.values(), key=lambda n: n.hopsAway)
        for node in nodes:
            output += f"{node.to_verbose_string()}\n"
        return output

    def to_succinct_string(self):
        """Used when sending the node list in Meshtastic messages"""
        return "\n".join(node.to_succinct_string() for node in self.nodes.values())
