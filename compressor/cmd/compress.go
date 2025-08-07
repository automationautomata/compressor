package cmd

import (
	comp "compressor/internal/compressing"
	"compressor/internal/huffman"
	"compressor/internal/utiles"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	huffmanCompressionType = "huff"
	OutputExt              = ".dedal"
)

var (
	compBlockSize int
	compType      string
	compDestDir   string
	compQuiet     bool
)

var compressCmd = &cobra.Command{
	Use:   "compress [flags] <files|directory>",
	Short: "Compress files or directories",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		compDestDir, err = filepath.Abs(compDestDir)
		if err != nil {
			return err
		}
		if ok, err := isDir(compDestDir); !(ok || os.IsNotExist(err)) {
			color.Red("--dest must be a directory")
			return err
		}
		if err := os.MkdirAll(compDestDir, 0755); err != nil {
			return err
		}

		pathes := args
		if ok, _ := isDir(pathes[0]); ok && len(pathes) == 1 {
			pathes, err = utiles.GetDirFiles(args[0])
			if err != nil {
				return err
			}
		}

		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		compArgs := map[string]any{"blockSize": compBlockSize}
		result, err := compression(pathes, compArgs, compDestDir, !compQuiet, ctx)
		if err != nil {
			if errors.Is(err, &comp.ErrCompression{}) {
				cmd.Println(color.RedString("Compression failed"))
			}
			return err
		}

		compFilePath, err := makeCompressedFile(compDestDir, pathes[0], result.tempPath)
		if err != nil {
			return err
		}
		if !compQuiet {
			cmd.Printf("\nOutput file: %s\n", color.GreenString(compFilePath))
			cmd.Printf("Footer size: %d bytes\n", result.footerSize)
			cmd.Printf("Output file total size: %d bytes\n", result.compSize+result.footerSize)
		}
		cmd.Println(color.GreenString("Compression succeeded!"))
		return nil
	},
}

func init() {
	compressCmd.Flags().IntVar(&compBlockSize, "block", 0, "block size for compression")
	compressCmd.Flags().StringVar(&compType, "type", huffmanCompressionType, "compression type")
	compressCmd.Flags().StringVar(&compDestDir, "dest", "", "directory of output file")
	compressCmd.Flags().BoolVarP(&compQuiet, "quiet", "q", false, "quiet mode (no progress output)")
}

func makeCompressedFile(path, name string, tempPath string) (finalPath string, err error) {
	name = strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	finalPath = getUniqueName(filepath.Join(path, name) + OutputExt)
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}
	if err := os.Rename(tempPath, finalPath); err != nil {
		return "", err
	}
	return finalPath, nil
}

type compressionOutput struct {
	tempPath   string
	compSize   int64
	footerSize int64
}

func compression(
	pathes []string,
	compArgs map[string]any,
	dstDir string,
	showProgress bool,
	ctx context.Context,
) (*compressionOutput, error) {

	totalSizeVal, err := totalSize(pathes)
	if err != nil {
		return nil, err
	}
	var totalSize int64 = totalSizeVal

	dstFile, err := os.CreateTemp(dstDir, "temp-comp-*.dedal-temp")
	if err != nil {
		return nil, err
	}

	result := &compressionOutput{
		tempPath: dstFile.Name(),
	}

	defer dstFile.Close()

	go cleanup(dstFile.Name(), ctx, dstFile)

	prog := utiles.NewProgress[int64](len(pathes))
	if showProgress {
		prog.ShowProgress(totalSize)
		defer prog.Close()
	} else {
		prog.Close()
	}

	var compSize, footerSize int64
	switch compType {
	case "huff":
		compressor := huffman.NewCompressor(compArgs["blockSize"].(int), totalSize)
		compSize, footerSize, err = comp.CompressFiles(compressor, pathes, dstFile, prog)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported compression type: %s", compType)
	}

	result.tempPath = dstFile.Name()
	result.compSize = compSize
	result.footerSize = footerSize
	return result, nil
}
