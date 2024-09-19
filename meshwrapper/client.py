import meshtastic
import meshtastic.tcp_interface
from pubsub import pub

from .node import Node
from .nodelist import Nodelist
from .message import Message


class MeshtasticConnectionLost(Exception):
    """Thrown when the Meshtastic node disconnects from your project"""

    pass


class MeshtasticClient:
    """The class used to connect your project to a Meshtastic node."""

    def __init__(self, hostname: str, connected=None, message=None, debug=False):
        pub.subscribe(self._on_receive, "meshtastic.receive")
        pub.subscribe(
            self._on_connection_established, "meshtastic.connection.established"
        )
        pub.subscribe(self._on_connection_lost, "meshtastic.connection.lost")
        pub.subscribe(self._on_node, "meshtastic.node")
        pub.subscribe(self._on_node_updated, "meshtastic.node.updated")
        if debug:
            pub.subscribe(self._debug, "meshtastic")

        self.connected = False
        self.closing = False
        self.nodeList = Nodelist()

        self._myNodeNum = None
        self._hostname = hostname
        self._connectedCallback = connected
        self._messageCallback = message
        self._interface = meshtastic.tcp_interface.TCPInterface(hostname=self._hostname)

    def close(self):
        self.closing = True
        self._interface.close()

    def _on_receive(self, packet, interface):
        message = Message(
            packet, self.nodeList.get(packet["from"]), self.nodeList.get(packet["to"])
        )
        if self._messageCallback:
            self._messageCallback(message)

    def _on_connection_established(self, interface, topic=pub.AUTO_TOPIC):
        self.connected = True
        self._myNodeNum = interface.myInfo.my_node_num
        self.nodeList.get(self._myNodeNum).mark_as_self()
        if self._connectedCallback:
            self._connectedCallback()

    def _on_connection_lost(self, interface, topic=pub.AUTO_TOPIC):
        self.connected = False
        if not self.closing:
            print("ERROR: Connection to node lost")
            raise MeshtasticConnectionLost("Connection to node was lost")

    def _on_node(self, node, interface, topic=pub.AUTO_TOPIC):
        node = Node(node, interface)
        if node.num is self._myNodeNum:
            node.mark_as_self()
        self.nodeList.add(node)

    def _on_node_updated(self, node, interface, topic=pub.AUTO_TOPIC):
        node = Node(node, interface)
        if node.num is self._myNodeNum:
            node.mark_as_self()
        self.nodeList.update(node)

    def _debug(self, interface=None, *args, **kwargs):
        for arg in args:
            print("Argument:", arg)
        for key, value in kwargs.items():
            print("Key, value:", key, value)
