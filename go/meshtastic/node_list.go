package meshtastic

import (
	"cmp"
	"slices"
)

type nodeList struct {
	nodes map[uint32]*Node
}

var Broadcast = Node{
	Id:        0xFFFFFFFF,
	ShortName: "CAST",
	LongName:  "Everyone",
}

var Unknown = Node{
	Id:        0x00000000,
	ShortName: "UNKN",
	LongName:  "Unknown",
}

func NewNodeList() nodeList {
	list := nodeList{
		nodes: make(map[uint32]*Node),
	}

	list.nodes[Broadcast.Id] = &Broadcast
	list.nodes[Unknown.Id] = &Unknown

	return list
}

func (n *nodeList) String() string {
	nodes := ""
	for _, node := range n.sortedNodes() {
		if node.Id != Broadcast.Id && node.Id != Unknown.Id {
			nodes += node.VerboseString() + "\n"
		}
	}
	return nodes
}

func (n *nodeList) sortedNodes() []Node {
	nodes := make([]Node, 0, len(n.nodes))
	for _, node := range n.nodes {
		nodes = append(nodes, *node)
	}
	slices.SortFunc(nodes, func(a, b Node) int {
		return cmp.Or(
			cmp.Compare(a.HopsAway, b.HopsAway),
			-cmp.Compare(a.LastHeard.Unix(), b.LastHeard.Unix()),
		)
	})
	return nodes
}
