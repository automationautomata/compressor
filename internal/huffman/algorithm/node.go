package algorithm

import (
	"sync"
)

type node struct {
	values    map[string]struct{}
	frequency uint64
	height    uint64
	left      *node
	right     *node
	mu        sync.RWMutex
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
	n.mu.RLock()
	defer n.mu.RUnlock()
	_, ok := n.values[string(bytes)]
	return ok
}

func (n *node) getHeight() uint64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.height
}

func (n *node) getFrequency() uint64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
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

func combine(left, right *node) *node {
	left.mu.RLock()
	defer left.mu.RUnlock()

	right.mu.RLock()
	defer right.mu.RUnlock()

	values := make(map[string]struct{})
	for s := range left.values {
		values[s] = struct{}{}
	}
	for s := range right.values {
		values[s] = struct{}{}
	}

	return &node{
		left:      left,
		right:     right,
		height:    max(left.height, right.height) + 1,
		frequency: left.frequency + right.frequency,
		values:    values,
	}
}

type nodesHeap []*node

func (n nodesHeap) Len() int { return len(n) }

func (n nodesHeap) Less(i, j int) bool { return n[i].compare(n[j]) < 0 }

func (n nodesHeap) Swap(i, j int) { n[i], n[j] = n[j], n[i] }

func (n *nodesHeap) Push(x any) { *n = append(*n, x.(*node)) }

func (n *nodesHeap) Pop() any {
	old := *n
	size := len(old)
	x := old[size-1]
	*n = old[0 : size-1]
	return x
}
