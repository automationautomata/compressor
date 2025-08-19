package cmd

import (
	comp "compressor/internal/compressing"
	"compressor/internal/huffman"
	"compressor/internal/utiles"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// selectDecompressor возвращает декомпрессор по типу сжатия.
func selectDecompressor(compType string) comp.Decompressor {
	switch compType {
	case huffman.CompressionType:
		return huffman.NewDecompressor()
	default:
		return nil
	}
}

var (
	decompDest  string
	decompQuiet bool
)
var uncompressCmd = &cobra.Command{
	Use:   "uncompress [flags] <file>",
	Short: "Decompress file",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if ok, err := isDir(decompDest); !(ok || os.IsNotExist(err)) {
			cmd.Println(color.RedString("--dest must be directory"))
			return err
		}

		ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		srcPath := args[0]
		showProgress := !decompQuiet
		dstDir := decompDest

		srcFile, err := os.Open(srcPath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// if there is only one compressed file, then give it a base name
		dstDir, err = filepath.Abs(getUniqueName(dstDir))
		if err != nil {
			cmd.Println("Failed to create destination directory")
			return err
		}
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			cmd.Println("Failed to create destination directory")
			return err
		}
		go cleanup(dstDir, ctx, srcFile)

		prog := utiles.NewProgress[int64](0)
		if info, err := srcFile.Stat(); showProgress && err == nil {
			prog.ShowProgress(info.Size())
			defer prog.Close()
		} else {
			prog.Close()
		}

		output, err := comp.Decompress(selectDecompressor, srcFile, dstDir, prog)
		if err != nil {
			cmd.Println(color.RedString("File can't be uncompressed! Decompression failed."))
			return err
		}
		if showProgress {
			cmd.Println("files: ")
			for _, out := range output {
				relPath := strings.TrimPrefix(out.Path, filepath.Clean(dstDir))

				if out.NewChecksum != out.OldChecksum {
					cmd.Println(color.RedString("Decompression failed."))
					return fmt.Errorf("Checksum for %s mismatch! Decompression failed.\n", relPath)
				}
				cmd.Println(color.GreenString(".%s", relPath))
			}
		}

		cmd.Println(color.GreenString("\nDecompression succeeded!"))
		return nil
	},
}

func init() {
	uncompressCmd.Flags().StringVar(&decompDest, "dest", "", "output file or directory path")
	uncompressCmd.Flags().BoolVarP(&decompQuiet, "quiet", "q", false, "quiet mode (no progress output)")
}
