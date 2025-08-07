package utiles

import (
	"fmt"
	"io"
	"os"
	"strings"
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

type TableParams struct {
	ColSep      string
	RowSep      string
	VerticalSep bool
	IndentSize  int
	Writer      io.Writer
}

func ShowTable(titles []string, rows [][]string, params TableParams) {
	w := params.Writer
	if w == nil {
		w = os.Stdout
	}
	indent := ""
	if params.IndentSize != 0 {
		indent = strings.Repeat(" ", params.IndentSize)
	}

	colSep, rowSep := params.ColSep, params.RowSep
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
		for j, fileItem := range rows[i] {
			rowWidths[j] = max(rowWidths[j], len(fileItem))
		}
	}

	titlesRow := formatRow(titles, rowWidths, colSep)
	fmt.Fprintf(w, "%s%s\n", indent, titlesRow)
	if params.VerticalSep {
		fmt.Fprintf(w, "%s%s\n", indent, strings.Repeat(rowSep, len(titlesRow)))
	}
	for i := range rows {
		fmtRow := formatRow(rows[i], rowWidths, colSep)
		fmt.Fprintf(w, "%s%s\n", indent, fmtRow)
		if params.VerticalSep {
			fmt.Fprintf(w, "%s%s\n", indent, strings.Repeat("-", len(fmtRow)))
		}
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
