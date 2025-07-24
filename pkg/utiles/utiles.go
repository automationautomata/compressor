package utiles

import (
	"fmt"
	"strings"

	"github.com/cheggaaa/pb/v3"
)

const (
	defaultWidth           = 0
	defaultColumnSeparator = " | "
	defaultRowSeparator    = "-"
)

func formatRow(row []string, rowWidths []int, sep string) string {
	fmtRows := make([]string, len(row))
	for i := range row {
		width := defaultWidth
		if i < len(rowWidths) {
			width = rowWidths[i]
		}
		fmtRows[i] = fmt.Sprintf("%-*s", width, row[i])
	}
	return strings.Join(fmtRows, sep)
}

func ShowTable(titles []string, rows [][]string, colSep, rowSep string) {
	if colSep == "" {
		colSep = defaultColumnSeparator
	}
	if rowSep == "" {
		rowSep = defaultRowSeparator
	}

	rowWidths := make([]int, len(titles))
	for i, title := range titles {
		rowWidths[i] = max(rowWidths[i], len(title))
	}
	for i := range rows {
		if len(rows[i]) > len(rowWidths) {
			extending := make([]int, len(rows[i])-len(rowWidths))
			rowWidths = append(rowWidths, extending...)
		}
		for j, item := range rows[i] {
			rowWidths[j] = max(rowWidths[j], len(item))
		}
	}

	titlesRow := formatRow(titles, rowWidths, colSep)
	fmt.Println(titlesRow)

	fmt.Println(strings.Repeat(rowSep, len(titlesRow)))

	for i := range rows {
		fmtRow := formatRow(rows[i], rowWidths, colSep)
		fmt.Println(fmtRow)
		fmt.Println(strings.Repeat("-", len(fmtRow)))
	}
}

// FormatBytes возвращает строку с hex-представлением байтов
func HexBytes(b []byte) string {
	bytes := make([]string, len(b))
	for i, v := range b {
		bytes[i] = fmt.Sprintf("%X", v)
	}
	return strings.Trim(fmt.Sprintf("%02X", b), "[]")
}

// BinBytes возвращает строку из нулей и единиц
func BinBytes(b []byte) string {
	n := len(b)
	bits := make([]string, n)
	for i, v := range b {
		bits[i] = fmt.Sprintf("%08b", v)
	}
	return strings.Join(bits, " ")
}

func ShowProgress64(total int64, progressChan chan int64) {
	bar := pb.Start64(total)
	for incr := range progressChan {
		bar.Add64(incr)
		if bar.Current() >= total {
			break
		}
	}
	bar.Finish()
}

func ShowProgress(total int, progressChan chan int) {
	int64Chan := make(chan int64)
	go func() {
		defer close(int64Chan)
		for v := range progressChan {
			int64Chan <- int64(v)
		}
	}()
	ShowProgress64(int64(total), int64Chan)
}
