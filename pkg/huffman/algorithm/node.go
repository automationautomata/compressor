package algorithm

import (
	"fmt"
	"strings"
	"sync"
)

type node struct {
	values    map[string]struct{}
	left      *node
	right     *node
	frequency uint64
	height    uint64
	sync.RWMutex
}

func newNode(bytes []byte, frequency uint64) *node {
	values := make(map[string]struct{})
	values[string(bytes)] = struct{}{}
	return &node{
		values:    values,
		frequency: frequency,
	}
}

func (n *node) hasValue(bytes []byte) bool {
	n.RLock()
	defer n.RUnlock()
	_, ok := n.values[string(bytes)]
	return ok
}

func (n *node) getHeight() uint64 {
	n.RLock()
	defer n.RUnlock()
	return n.height
}

func (n *node) getFrequency() uint64 {
	n.RLock()
	defer n.RUnlock()
	return n.frequency
}

func (n *node) compare(m *node) int {
	nFreq, mFreq := n.getFrequency(), m.getFrequency()

	if nFreq > mFreq {
		return 1
	}
	if nFreq < mFreq {
		return -1
	}

	nHeight, mHeight := n.getHeight(), m.getHeight()
	if nHeight > mHeight {
		return 1
	}
	if nHeight < mHeight {
		return -1
	}

	return 0
}

func combineNodes(left, right *node) *node {
	left.RLock()
	defer left.RUnlock()

	right.RLock()
	defer right.RUnlock()

	values := make(map[string]struct{})
	for k := range left.values {
		values[k] = struct{}{}
	}
	for k := range right.values {
		values[k] = struct{}{}
	}

	return &node{
		left:      left,
		right:     right,
		height:    max(left.height, right.height) + 1,
		frequency: left.frequency + right.frequency,
		values:    values,
	}
}

func printNode(node *node, step, sep, space, side string) {
	if node == nil {
		return
	}
	node.RLock()
	defer node.RUnlock()
	symbols := make([]string, 0, len(node.values))
	for k := range node.values {
		symbols = append(symbols, k)
	}
	out := fmt.Sprint(space, sep, step, strings.Join(symbols, "."), " ", node.frequency, side)
	fmt.Println(out)
	printNode(node.left, step+step[0:1], sep, space, " L")
	printNode(node.right, step+step[0:1], sep, space, " R")
}

type nodesHeap []*node

func (n nodesHeap) Len() int {
	return len(n)
}

func (n nodesHeap) Less(i, j int) bool {
	return n[i].compare(n[j]) < 0
}

func (n nodesHeap) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n *nodesHeap) Push(x any) {
	*n = append(*n, x.(*node))
}

func (n *nodesHeap) Pop() any {
	old := *n
	size := len(old)
	x := old[size-1]
	*n = old[0 : size-1]
	return x
}
