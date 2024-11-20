package meshtastic

import (
	"fmt"
	"time"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"github.com/timendus/meshbot/meshtastic/helpers"
)

type Neighbor struct {
	Node         *Node
	Snr          float32
	LastReported time.Time
}

func (n *Neighbor) String() string {
	return fmt.Sprintf("%s \033[90m(last reported %s ago, SNR %.2f)\033[0m", n.Node.String(), helpers.TimeAgo(n.LastReported), n.Snr)
}

type NeighborList []Neighbor

func NewNeighbourList(nodelist *nodeList, timestamp uint32, neighbors []*meshtastic.Neighbor) NeighborList {
	properTimestamp := time.Unix(int64(timestamp), 0)
	neighbourList := make([]Neighbor, 0)
	for _, neighbor := range neighbors {
		node := nodelist.nodes[neighbor.NodeId]
		if node == nil {
			node = NewNode(&meshtastic.NodeInfo{
				Num: neighbor.NodeId,
			})
			nodelist.nodes[neighbor.NodeId] = node
		}
		neighbourList = append(neighbourList, Neighbor{
			Node:         node,
			Snr:          neighbor.Snr,
			LastReported: properTimestamp,
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
