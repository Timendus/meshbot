package meshtastic

import (
	"cmp"
	"slices"
)

type nodeList struct {
	nodes map[uint32]Node
}

func NewNodeList() nodeList {
	return nodeList{
		nodes: make(map[uint32]Node),
	}
}

func (n *nodeList) String() string {
	nodes := ""
	for _, node := range n.sortedNodes() {
		nodes += node.VerboseString() + "\n"
	}
	return nodes
}

func (n *nodeList) sortedNodes() []Node {
	nodes := make([]Node, 0, len(n.nodes))
	for _, node := range n.nodes {
		nodes = append(nodes, node)
	}
	slices.SortFunc(nodes, func(a, b Node) int {
		return cmp.Or(
			cmp.Compare(a.HopsAway, b.HopsAway),
			-cmp.Compare(a.LastHeard.Unix(), b.LastHeard.Unix()),
		)
	})
	return nodes
}
