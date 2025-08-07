package algorithm

import (
	"compressor/internal/utiles"
	"container/heap"
	"io"
	"strings"
	"sync"
)

type code struct {
	bits []byte
	sync.RWMutex
}

// NewCode создает новый код длинны length бит
func newCode(length uint64) *code {
	return &code{
		bits: make([]byte, length),
	}
}

func (c *code) writeBit(bit bool) {
	if bit {
		c.bits = append(c.bits, 1)
	} else {
		c.bits = append(c.bits, 0)
	}
}

func (c *code) bytes() []byte {
	bitCount := len(c.bits)
	if bitCount == 0 {
		return nil
	}

	byteCount := (bitCount + 7) / 8
	result := make([]byte, byteCount)

	for i, bit := range c.bits {
		if bit != 0 {
			ind := i / 8
			offset := 7 - (i % 8) // старший бит в байте — первый
			result[ind] |= 1 << offset
		}
	}

	return result
}

func (c *code) String() string {
	var sb strings.Builder
	for _, b := range c.bits {
		sb.WriteByte('0' + b)
	}
	return sb.String()
}

type HuffmanTree struct {
	root      *node
	size      int
	BlockSize int
	Alphabet  [][]byte
}

func NewHuffmanTree(blockSize int) *HuffmanTree {
	return &HuffmanTree{
		root:      nil,
		size:      0,
		BlockSize: blockSize,
		Alphabet:  nil,
	}
}

// CountFrequencies читает файл блоками и возвращает частоты блоков
func CountFrequencies(file io.Reader, blockSize int) (map[string]uint64, error) {
	buf := make([]byte, blockSize)
	frequencies := make(map[string]uint64)

	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			if err != io.ErrUnexpectedEOF {
				return nil, err
			}
		}
		block := string(buf[:n])
		frequencies[block]++
	}

	return frequencies, nil
}

func (huff *HuffmanTree) BuildTree(frequencies map[string]uint64) error {
	if len(frequencies) == 0 {
		return nil
	}
	nodes := make(nodesHeap, 0, len(frequencies))
	alphabet := make([][]byte, 0, len(frequencies))
	for symb, freq := range frequencies {
		bytes := []byte(symb)
		heap.Push(&nodes, newNode(bytes, freq))
		alphabet = append(alphabet, bytes)
	}
	huff.Alphabet = alphabet

	for len(nodes) != 1 {
		node1 := heap.Pop(&nodes).(*node)
		node2 := heap.Pop(&nodes).(*node)
		heap.Push(&nodes, combine(node1, node2))
	}
	huff.root = heap.Pop(&nodes).(*node)
	return nil
}

func (huff *HuffmanTree) EncodeTable() map[string][]byte {
	codes := utiles.NewSafeMap[string, *code]()
	var wg sync.WaitGroup
	wg.Add(len(huff.Alphabet))

	for i := range huff.Alphabet {
		block := huff.Alphabet[i]
		go func(block []byte) {
			defer wg.Done()
			c := newCode(0)
			encodingSearch(huff.root, block, c, true, 0)
			codes.Store(string(block), c)
		}(block)
	}
	wg.Wait()

	bytesFromCodes := make(map[string][]byte)
	codes.Range(func(s string, c *code) (pass bool) {
		bytesFromCodes[s] = c.bytes()
		return true
	})

	return bytesFromCodes
}

func encodingSearch(n *node, value []byte, c *code, bitFlag bool, height uint64) {
	if n == nil || !n.hasValue(value) {
		return
	}

	if height > 0 {
		if bitFlag {
			c.writeBit(true)
		} else {
			c.writeBit(false)
		}
	}
	encodingSearch(n.right, value, c, true, height+1)
	encodingSearch(n.left, value, c, false, height+1)
}
