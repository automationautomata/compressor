package huffman

import (
	comp "compressor/internal/compressing"
	alg "compressor/internal/huffman/algorithm"
	"compressor/internal/utiles"
	"fmt"
	"io"
	"math"

	"golang.org/x/sync/errgroup"
)

const (
	minBlockSize    = 1
	maxBlockSize    = 65536
	CompressionType = "HUFF"
)

type ErrNoCode struct{ code []byte }

func (e *ErrNoCode) Error() string {
	return fmt.Sprintf("Code not found for block %s", utiles.HexBytes(e.code))
}

// ComputeBlockSize вычисляет размер блока алфавита в зависимости от размера файла
func computeBlockSize(size int64) int {
	if size <= 0 {
		return minBlockSize
	}

	kb := float64(size) / 1024.0

	blockSize := int(minBlockSize * math.Sqrt(2*kb))

	// Ограничиваем размер сверху и снизу
	if blockSize < minBlockSize {
		blockSize = minBlockSize
	}
	blockSize = min(blockSize, maxBlockSize)

	return blockSize
}

type Compressor struct {
	BlockSize int
	codes     map[string][]byte
}

func NewCompressor(blockSize int, totalSize int64) *Compressor {
	blockSize = computeBlockSize(totalSize)
	return &Compressor{blockSize, nil}
}

func calcSizes(codes map[string][]byte, srcSymbols []map[string]uint64) []int64 {
	sizes := make([]int64, len(srcSymbols))
	for i := range srcSymbols {
		for symb, freq := range srcSymbols[i] {
			sizes[i] += int64(uint64(len(codes[symb])) * freq)
		}
	}
	return sizes
}

func (c *Compressor) Preprocessing(srcs []io.Reader) ([]int64, error) {
	var eg errgroup.Group
	freqs := make([]map[string]uint64, len(srcs))
	for i, src := range srcs {
		eg.Go(func() (err error) {
			freqs[i], err = alg.CountFrequencies(src, c.BlockSize)
			if err != nil {
				return err
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	generalFreq := make(map[string]uint64)
	for i := range freqs {
		for symb, freq := range freqs[i] {
			generalFreq[symb] += freq
		}
	}

	huff := alg.NewHuffmanTree(c.BlockSize)
	if err := huff.BuildTree(generalFreq); err != nil {
		return nil, err
	}
	c.codes = huff.EncodeTable()
	return calcSizes(c.codes, freqs), nil
}

func (c *Compressor) CompressorData() (string, comp.Body) { return CompressionType, c.codes }

func (c *Compressor) CompressFile(
	src io.Reader, dst io.Writer, prog *utiles.Progress[int64],
) (size int64, err error) {
	buf := make([]byte, c.BlockSize)
	n := 0
	for {
		n, err = src.Read(buf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return 0, err
		}

		code, ok := c.codes[string(buf[:n])]
		if !ok {
			return 0, &ErrNoCode{buf[:n]}
		}

		n, err = dst.Write(code)
		if err != nil {
			return 0, err
		}

		prog.Write(int64(n))
		size += int64(n)
	}
	return size, nil
}

// func showEncodeTable(codes map[string]*[]byte) {
// 	symbolTitle := "symbol"
// 	byteTitle := "byte"
// 	codeTitle := "code"
// 	codeBytesTitle := "code bytes"
// 	var rows [][]string
// 	for block, code := range codes {
// 		symbol := fmt.Sprintf("%q", block)
// 		byteRepr := utiles.HexBytes([]byte(block))
// 		codeStr := code.String()
// 		row := []string{
// 			symbol,
// 			byteRepr,
// 			codeStr,
// 			utiles.HexBytes(code.Bytes()),
// 		}
// 		rows = append(rows, row)
// 	}
// 	slices.SortFunc(rows, func(row1, row2 []string) int {
// 		if row1[1] > row2[1] {
// 			return 1
// 		}
// 		if row1[1] < row2[1] {
// 			return -1
// 		}
// 		return 0
// 	})
// 	titles := []string{symbolTitle, byteTitle, codeTitle, codeBytesTitle}
// 	utiles.ShowTable(titles, rows, "", "")
// }
