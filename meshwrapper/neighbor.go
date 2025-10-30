package meshwrapper

import (
	"fmt"
	"time"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"github.com/timendus/meshbot/meshwrapper/helpers"
)

type Neighbor struct {
	Node         *Node
	Snr          float32
	LastReported time.Time
}

func (n *Neighbor) String() string {
	return fmt.Sprintf("%s \033[90m(last reported %s ago, SNR %.2f)\033[0m", n.Node.ColorString(), helpers.TimeAgo(n.LastReported), n.Snr)
}

type NeighborList []Neighbor

func NewNeighbourList(connectedNode *ConnectedNode, message IncomingMessage) NeighborList {
	neighbourList := make([]Neighbor, 0)
	for _, neighbor := range message.NeighborInfo.Neighbors {
		node, ok := connectedNode.NodeList.nodes[neighbor.NodeId]
		if !ok {
			node = NewNode(connectedNode, &meshtastic.NodeInfo{
				Num: neighbor.NodeId,
			})
			connectedNode.NodeList.nodes[neighbor.NodeId] = node
		}
		neighbourList = append(neighbourList, Neighbor{
			Node:         node,
			Snr:          neighbor.Snr,
			LastReported: message.Timestamp,
		})
	}
	return neighbourList
}

func (nl NeighborList) String() string {
	nodes := ""
	for _, node := range nl {
		nodes += "\n   - " + node.String()
	}
	return nodes
}
