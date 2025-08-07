package cmd

import (
	comp "compressor/internal/compressing"
	"sort"

	"compressor/internal/utiles"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var metadataCmd = &cobra.Command{
	Use:   "metadata <file>",
	Short: "Print file metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		defer func() {
			if r := recover(); r != nil {
				cobra.CompErrorln("Metadata not found")
			}
		}()

		path := args[0]

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		md, size, err := comp.ReadFooterMetadata(file)
		if err != nil {
			cmd.Println(color.RedString("File doesn't contain meatadata"))
			return err
		}
		cmd.Printf("Size: %d bytes\n", size)

		titles := []string{"File", "Size", "Checksum"}
		rows := make([][]string, len(md.FileMap))
		files := md.FileMap
		for i := range rows {
			rows[i] = []string{
				files[i].Path,
				fmt.Sprintf("%d bytes", files[i].Size),
				files[i].Checksum,
			}
		}
		sort.Slice(rows, func(i, j int) bool {
			return rows[i][0] < rows[j][0]
		})
		cmd.Println()
		tp := utiles.TableParams{
			ColSep:      "   ",
			RowSep:      "   ",
			VerticalSep: false,
			Writer:      cmd.OutOrStdout(),
		}
		utiles.ShowTable(titles, rows, tp)
		cmd.Println()
		return nil
	},
}
