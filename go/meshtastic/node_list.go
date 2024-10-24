package meshtastic

import (
	"cmp"
	"slices"
)

type nodeList struct {
	nodes map[uint32]node
}

func NewNodeList() nodeList {
	return nodeList{
		nodes: make(map[uint32]node),
	}
}

func (n *nodeList) String() string {
	nodes := ""
	for _, node := range n.sortedNodes() {
		nodes += node.VerboseString() + "\n"
	}
	return nodes
}

func (n *nodeList) sortedNodes() []node {
	nodes := make([]node, 0, len(n.nodes))
	for _, node := range n.nodes {
		nodes = append(nodes, node)
	}
	slices.SortFunc(nodes, func(a, b node) int {
		return cmp.Or(
			cmp.Compare(a.hopsAway, b.hopsAway),
			-cmp.Compare(a.lastHeard.Unix(), b.lastHeard.Unix()),
		)
	})
	return nodes
}
