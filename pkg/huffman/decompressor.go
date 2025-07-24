package huffman

import (
	comp "archiver/pkg/compressing"
	"archiver/pkg/utiles"
	"fmt"
	"io"
	"slices"
)

type ErrNoSymbol struct {
	code []byte
}

func (e *ErrNoSymbol) Error() string {
	hexCode := utiles.HexBytes(e.code)
	return fmt.Sprintf("The code %s doesn't match any symbol", hexCode)
}

type Decompressor struct{}

func NewDecompressor() *Decompressor {
	return &Decompressor{}
}

func (d *Decompressor) HeaderDataType() comp.HeaderData {
	return make(map[string][]byte)
}

func (d *Decompressor) Preprocessing(_ io.ReadSeeker) error {
	return nil
}

func (d *Decompressor) DecompressFile(dd *comp.DecompressionInput, progressChan chan int64) (err error) {
	srcFile, dstFile := dd.SourceFile, dd.DestFile

	codes := dd.Data.(map[string][]byte)
	codesToSymbols := make(map[string][]byte)
	minCodeLen, maxCodeLen := (1<<32)-1, 0
	for symb, code := range codes {
		codesToSymbols[string(code)] = []byte(symb)
		minCodeLen = min(minCodeLen, len(code))
		maxCodeLen = max(maxCodeLen, len(code))
	}

	buf := make([]byte, maxCodeLen)
	var i int
	for {
		matchedSymb := []byte{}
		for i = minCodeLen; i <= maxCodeLen; i++ {
			_, err := srcFile.Read(buf[i-1 : i])
			if err != nil {
				if err == io.EOF {
					return nil
				}
				if err != io.ErrUnexpectedEOF {
					return err
				}
			}
			if symb, ok := codesToSymbols[string(buf[:i])]; ok {
				matchedSymb = symb
				break
			}
		}
		if len(matchedSymb) == 0 {
			return &ErrNoCode{buf}
		}
		if _, err = dstFile.Write(matchedSymb); err != nil {
			return err
		}
		clear(buf)

		if _, ok := <-progressChan; ok {
			progressChan <- int64(i)
		}
	}
}

func showHeaderData(codes map[string][]byte) {
	titles := []string{"symbol", "symbol bytes", "code bytes", "bin code"}

	rows := make([][]string, 0, len(codes))
	for symb, code := range codes {
		row := []string{
			fmt.Sprintf("%q", symb),
			utiles.HexBytes([]byte(symb)),
			utiles.HexBytes(code),
			utiles.BinBytes(code),
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
	utiles.ShowTable(titles, rows, "", "")
}
