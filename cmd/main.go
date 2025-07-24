package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: archiver <compress|decompress> [options] <file>")
		return
	}

	// Определяем флаги для подкоманды compress
	compressCmd := flag.NewFlagSet("compress", flag.ExitOnError)
	compBlockSize := compressCmd.Int("block", 0, "block size for compression")
	compType := compressCmd.String("type", "huff", "compression type (default huff)")
	compDest := compressCmd.String("dest", "", "output file or directory path")
	compQuiet := compressCmd.Bool("q", false, "quiet mode (no progress output)")

	// Флаги для подкоманды decompress
	decompressCmd := flag.NewFlagSet("decompress", flag.ExitOnError)
	decompDest := decompressCmd.String("dest", "", "output file or directory path")
	decompQuiet := decompressCmd.Bool("q", false, "quiet mode (no progress output)")

	switch os.Args[1] {
	case "compress":
		if err := compressCmd.Parse(os.Args[2:]); err != nil {
			printError(err)
			return
		}
		args := compressCmd.Args()
		if len(args) < 1 {
			fmt.Println("Error: input file path required for compress")
			compressCmd.Usage()
			return
		}
		inputPath := args[0]
		showProgress := !*compQuiet
		compressFile(inputPath, *compDest, *compType, *compBlockSize, showProgress)

	case "decompress":
		if err := decompressCmd.Parse(os.Args[2:]); err != nil {
			printError(err)
			return
		}
		args := decompressCmd.Args()
		if len(args) < 1 {
			fmt.Println("Error: input file path required for decompress")
			decompressCmd.Usage()
			return
		}
		inputPath := args[0]
		showProgress := !*decompQuiet
		decompressFile(inputPath, *decompDest, showProgress)

	default:
		fmt.Println("Expected 'compress' or 'decompress' subcommands")
	}
}
