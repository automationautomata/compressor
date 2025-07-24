package huffman

import (
	comp "archiver/pkg/compressing"
	alg "archiver/pkg/huffman/algorithm"
	"archiver/pkg/utiles"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
)

const (
	minBlockSize    = 1
	maxBlockSize    = 65536
	CompressionType = "HUFF"
)

type ErrNoCode struct {
	code []byte
}

func (e *ErrNoCode) Error() string {
	return fmt.Sprintf("Code not found for block %s", utiles.HexBytes(e.code))
}

type Compressor struct {
	BlockSize int
	FilePath  string
	Header    *comp.Header
}

func NewCompressor(blockSize int, filePath string) *Compressor {
	info, _ := os.Stat(filePath)
	if blockSize == 0 {
		blockSize = computeBlockSize(info.Size())
	}
	return &Compressor{blockSize, filePath, nil}
}

// ComputeBlockSize вычисляет размер блока алфавита в зависимости от размера файла
func computeBlockSize(fileSize int64) int {
	if fileSize <= 0 {
		return minBlockSize
	}

	kb := float64(fileSize) / 1024.0

	size := int(minBlockSize * math.Sqrt(2*kb))

	// Ограничиваем размер сверху и снизу
	if size < minBlockSize {
		size = minBlockSize
	}
	if size > maxBlockSize {
		size = maxBlockSize
	}

	return size
}

func (c *Compressor) Preprocessing(srcFile io.ReadSeeker) error {
	freq, err := alg.CountFrequencies(srcFile, c.BlockSize)
	if err != nil {
		return err
	}

	huff := alg.NewHuffmanTree(c.BlockSize)
	huff.BuildTree(freq)
	codes := huff.EncodeTable()

	bytesFromCodes := make(map[string][]byte)
	for block := range codes {
		bytesFromCodes[block] = codes[block].Bytes()
	}

	sum, _ := comp.CalcCheckSum(srcFile)
	c.Header = comp.NewHeader(
		filepath.Base(c.FilePath),
		sum,
		CompressionType,
		bytesFromCodes,
	)

	return nil
}

func (c *Compressor) GetHeader() *comp.Header {
	return c.Header
}

func (c *Compressor) CompressFile(srcFile io.ReadSeeker, dstFile io.WriteSeeker, progreessChan chan int64) error {
	buf := make([]byte, c.BlockSize)
	codes := c.Header.Data.(map[string][]byte)
	for {
		n, err := srcFile.Read(buf)
		if len(progreessChan) != 0 && n > 0 {
			progreessChan <- int64(n)
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			if err != io.ErrUnexpectedEOF {
				return err
			}
		}

		code, ok := codes[string(buf[:n])]
		if !ok {
			return &ErrNoCode{buf[:n]}
		}
		if _, err := dstFile.Write(code); err != nil {
			return err
		}
	}
	return nil
}

func showEncodeTable(codes map[string]*alg.Code) {
	symbolTitle := "symbol"
	byteTitle := "byte"
	codeTitle := "code"
	codeBytesTitle := "code bytes"

	var rows [][]string

	for block, code := range codes {
		symbol := fmt.Sprintf("%q", block)
		byteRepr := utiles.HexBytes([]byte(block))
		codeStr := code.String()
		row := []string{
			symbol,
			byteRepr,
			codeStr,
			utiles.HexBytes(code.Bytes()),
		}
		rows = append(rows, row)
	}
	slices.SortFunc(rows, func(row1, row2 []string) int {
		if row1[1] > row2[1] {
			return 1
		}
		if row1[1] < row2[1] {
			return -1
		}
		return 0
	})
	titles := []string{symbolTitle, byteTitle, codeTitle, codeBytesTitle}
	utiles.ShowTable(titles, rows, "", "")

}
