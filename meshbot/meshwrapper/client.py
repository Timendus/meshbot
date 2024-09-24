from pubsub import pub
from typing import Callable
import logging

import meshtastic
import meshtastic.tcp_interface
import meshtastic.serial_interface

from .node import Node, Everyone
from .nodelist import Nodelist
from .message import Message

logger = logging.getLogger("Meshbot")


class MeshtasticConnectionLost(Exception):
    """Thrown when the Meshtastic node disconnects from your project"""

    pass


class MeshtasticClient:
    """The class used to connect your project to a Meshtastic node."""

    def __init__(
        self,
        hostname: str = None,
        device: str = None,
        connected: Callable[[], None] = None,
        message: Callable[[Message], None] = None,
        debug: bool = False,
    ):
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
        self._connectedCallback = connected
        self._messageCallback = message

        if hostname:
            self._interface = meshtastic.tcp_interface.TCPInterface(hostname=hostname)
        else:
            self._interface = meshtastic.serial_interface.SerialInterface(
                devPath=device
            )

        Everyone.interface = self._interface

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
            logger.error("ERROR: Connection to node lost")
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
            logger.debug("Argument:", arg)
        for key, value in kwargs.items():
            logger.debug("Key, value:", key, value)
