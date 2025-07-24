package main

import (
	comp "archiver/pkg/compressing"
	"archiver/pkg/huffman"
	"archiver/pkg/utiles"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

const CompressedExt = ".dedal"

func makeCompressedFileName(path, origName string) string {
	if path != "" && !strings.HasSuffix(path, "/") && !strings.HasSuffix(path, "\\") {
		return path
	}
	baseName := strings.TrimSuffix(filepath.Base(origName), filepath.Ext(origName))
	return filepath.Join(path, baseName+CompressedExt)
}

func restoreOriginalName(path, origName string) string {
	if path != "" && !strings.HasSuffix(path, "/") && !strings.HasSuffix(path, "\\") {
		return path
	}
	dir := filepath.Dir(path)
	return filepath.Join(dir, origName)
}

func printError(err error) {
	if err == nil {
		return
	}
	color.Red("ERROR:")
	fmt.Println("\t", err)
	fmt.Println()
}

// selectDecompressor возвращает декомпрессор по типу сжатия.
func selectDecompressor(compType string) comp.Decompressor {
	switch compType {
	case huffman.CompressionType:
		return huffman.NewDecompressor()
	default:
		return nil
	}
}

// decompressFile выполняет распаковку файла srcPath в dstPath.
func decompressFile(srcPath, dstPath string, showProgress bool) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		printError(err)
		return
	}
	defer srcFile.Close()

	dstFile, err := os.CreateTemp(".", "tmp-decompress-*.tmp")
	if err != nil {
		printError(err)
		return
	}

	var progressChan chan int64
	if showProgress {
		info, err := srcFile.Stat()
		if err == nil {
			progressChan = make(chan int64)
			go utiles.ShowProgress64(info.Size(), progressChan)
			defer close(progressChan)
		}
	}

	origName, oldChecksum, err := comp.Decompress(selectDecompressor, srcFile, dstFile, progressChan)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Original file name: %s\n", origName)

	newChecksum, err := comp.CalcCheckSum(dstFile)
	if err != nil {
		printError(err)
		return
	}

	if newChecksum != oldChecksum {
		color.Red("Checksum mismatch! Decompression failed.")
		return
	}

	tempName := dstFile.Name()
	if err := dstFile.Close(); err != nil {
		printError(err)
		return
	}

	finalPath := restoreOriginalName(dstPath, origName)
	if err := os.Rename(tempName, finalPath); err != nil {
		printError(err)
		return
	}

	color.Green("Decompression succeeded!")
	fmt.Printf("Output file: %s\n", finalPath)
}

// compressFile выполняет сжатие файла filePath в dstPath с указанным типом и размером блока.
func compressFile(filePath, dstPath, compType string, blockSize int, showProgress bool) {
	var compressor comp.Compressor

	switch compType {
	case "huff":
		compressor = huffman.NewCompressor(blockSize, filePath)
	default:
		color.Red("Unsupported compression type: %s\n", compType)
		return
	}

	srcFile, err := os.Open(filePath)
	if err != nil {
		printError(err)
		return
	}
	defer srcFile.Close()

	dstPath = makeCompressedFileName(dstPath, filepath.Base(filePath))

	dstFile, err := os.Create(dstPath)
	if err != nil {
		printError(err)
		return
	}
	defer dstFile.Close()

	var progressChan chan int64
	if showProgress {
		info, err := srcFile.Stat()
		if err == nil {
			fmt.Printf("File size: %d bytes\n", info.Size())
			progressChan = make(chan int64)
			go utiles.ShowProgress64(info.Size(), progressChan)
			defer close(progressChan)
		}
	}

	headerSize, err := comp.Compress(compressor, srcFile, dstFile, progressChan)
	if err != nil {
		printError(err)
		return
	}

	if showProgress {
		info, err := dstFile.Stat()
		if err == nil {
			fmt.Printf("\nOutput file: %s\n", color.GreenString(dstPath))
			fmt.Printf("Header size: %d bytes\n", headerSize)
			fmt.Printf("Output file size: %d bytes\n", info.Size())
		}
	}
	color.Green("Compression succeeded!")
}
