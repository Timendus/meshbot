package meshwrapper

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/timendus/meshbot/meshwrapper/helpers"
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

func (n *nodeList) Neighbours() string {
	nodes := ""
	for _, node := range n.sortedNodes() {
		nodeIsValid := node.Id != Broadcast.Id && node.Id != Unknown.Id
		nodeIsNeighbour := node.HopsAway == 0
		nodeHeardInLastHour := int(time.Since(node.LastHeard).Seconds()) < 3600

		if nodeIsValid && nodeIsNeighbour && nodeHeardInLastHour && !node.IsSelf() {
			nodes += " - " + node.String()
			nodes += fmt.Sprintf(" - %s ago", helpers.TimeAgo(node.LastHeard))
			if node.Snr != 0 {
				nodes += fmt.Sprintf(", %.2fdB", node.Snr)
			}
			nodes += "\n"
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

func (n *nodeList) findNode(needle string) *Node {
	needle = strings.TrimSpace(needle)
	needleBytes := []byte(needle)

	// Check if we have a specific, full hexadecimal id
	fullHexId, _ := regexp.Compile("![0-9a-fA-F]{8}")
	if fullHexId.Match(needleBytes) {
		id, err := strconv.ParseUint(needle[1:], 16, 32)
		node, ok := n.nodes[uint32(id)]
		if ok && err == nil {
			return node
		}
	}
	shortHexId, _ := regexp.Compile("[0-9a-fA-F]{8}")
	if shortHexId.Match(needleBytes) {
		id, err := strconv.ParseUint(needle, 16, 32)
		node, ok := n.nodes[uint32(id)]
		if ok && err == nil {
			return node
		}
	}

	// Check if we have a shortName
	for _, node := range n.nodes {
		if strings.EqualFold(node.ShortName, needle) {
			return node
		}
	}

	// Check if we have a decimal id
	numericId, _ := regexp.Compile("[0-9]+")
	if numericId.Match(needleBytes) {
		id, err := strconv.ParseUint(needle, 10, 32)
		node, ok := n.nodes[uint32(id)]
		if ok && err == nil {
			return node
		}
	}

	// Check if we have an abbreviated hexadecimal id
	abbreviatedHexId, _ := regexp.Compile("[0-9a-fA-F]{4}")
	if abbreviatedHexId.Match(needleBytes) {
		for _, node := range n.nodes {
			if strings.HasSuffix(node.GetIDExpression(), needle) {
				return node
			}
		}
	}

	// Check is needle is a substring of a longname
	for _, node := range n.nodes {
		if strings.Contains(strings.ToUpper(node.LongName), strings.ToUpper(needle)) {
			return node
		}
	}

	return nil
}
