package main

import (
	"compressor/compressor/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(cmd.NewRootCommand()); err != nil {
		os.Exit(1)
	}
}
