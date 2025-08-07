package huffman

import (
	comp "compressor/internal/compressing"
	"compressor/internal/utiles"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"slices"
)

type ErrNoSymbol struct{ code []byte }

func (e *ErrNoSymbol) Error() string {
	return fmt.Sprintf("The code %s doesn't match any symbol", utiles.HexBytes(e.code))
}

type Decompressor struct{}

func NewDecompressor() *Decompressor { return &Decompressor{} }

func (d *Decompressor) FooterBodyType() comp.Body {
	m := make(map[string][]byte)
	return &m
}

func (d *Decompressor) Preprocessing(_ comp.Body, _ io.ReadSeeker) error { return nil }

func (d *Decompressor) DecompressFile(dd *comp.DecompressionInput, prog *utiles.Progress[int64]) error {
	src, dst := dd.SourceFile, dd.DestFile

	codes := *(dd.Body.(*map[string][]byte))
	// showFooterData(codes)
	codesToSymbols := make(map[string][]byte)
	minCodeLen, maxCodeLen := (1<<32)-1, 0
	for symb, code := range codes {
		codesToSymbols[string(code)] = []byte(symb)
		minCodeLen = min(minCodeLen, len(code))
		maxCodeLen = max(maxCodeLen, len(code))
	}

	buf := make([]byte, maxCodeLen)
	for {
		var matched []byte
		found := false

		if err := binary.Read(src, binary.NativeEndian, buf[:minCodeLen-1]); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil
			}
			return err
		}
		var newByte byte
		for i := minCodeLen - 1; i < maxCodeLen; i++ {
			if err := binary.Read(src, binary.NativeEndian, &newByte); err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					return nil
				}
				return err
			}

			buf[i] = newByte
			if symb, ok := codesToSymbols[string(buf[:i+1])]; ok {
				matched, found = symb, true
				break
			}
		}
		if !found {
			return &ErrNoCode{buf}
		}
		if _, err := dst.Write(matched); err != nil {
			return err
		}
		clear(buf)
		prog.Write(int64(len(matched)))
	}
}

func showFooterData(codes map[string][]byte) {
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
	t, _ := os.Create("Table.txt")
	defer t.Close()
	p := utiles.TableParams{
		ColSep:      "",
		RowSep:      "",
		VerticalSep: true,
		IndentSize:  0,
		Writer:      t,
	}
	utiles.ShowTable(titles, rows, p)
}
