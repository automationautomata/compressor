package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "compressor",
		Short: "Compressor is a CLI tool for files or directory compressing and decompressing ",
	}
	rootCmd.AddCommand(compressCmd, decompressCmd, metadataCmd)
	return rootCmd
}

func Execute(rootCmd *cobra.Command) error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
