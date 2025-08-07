package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Context-aware temp directory cleanup
func cleanup(pathToRemove string, ctx context.Context, filesToClose ...*os.File) {
	select {
	case <-ctx.Done():
		for _, f := range filesToClose {
			f.Close()
		}
		if pathToRemove == "" {
			return
		}
		if err := os.RemoveAll(pathToRemove); err != nil {
			fmt.Fprintf(os.Stderr, "Error while removing temp files: %v\n", err)
		}
	}
}

func getUniqueName(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	ext := filepath.Ext(path)
	name := strings.TrimSuffix(path, ext)
	for i := 1; ; i++ {
		newName := fmt.Sprintf("%s (%d)%s", name, i, ext)
		if _, err := os.Stat(newName); os.IsNotExist(err) {
			return newName
		}
	}
}

func totalSize(pathes []string) (int64, error) {
	totalSize := int64(0)
	for i := range pathes {
		info, err := os.Stat(pathes[i])
		if err != nil {
			return 0, err
		}
		totalSize += info.Size()
	}
	return totalSize, nil
}

func isDir(path string) (bool, error) {
	info, err := os.Stat(path)
	return err == nil && info.IsDir(), err
}
