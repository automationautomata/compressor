package compressing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// checksum computes the SHA-256 checksum of the data read from r and returns it as a hex string.
func checksum(r io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func formatPathes(fileMap []File) error {
	if len(fileMap) == 1 {
		fileMap[0].Path = filepath.Base(fileMap[0].Path)
		return nil
	}
	splittedPathes := make([][]string, len(fileMap))
	for i, f := range fileMap {
		cleaned := filepath.Clean(f.Path)
		splittedPathes[i] = strings.Split(cleaned, string(filepath.Separator))
	}

	shortest := splittedPathes[0]
	for _, p := range splittedPathes {
		if len(p) < len(shortest) {
			shortest = p
		}
	}

	prefixLen := len(shortest)
	for _, p := range splittedPathes {
		for i := 0; i < prefixLen; i++ {
			if p[i] != shortest[i] {
				prefixLen = i
				break
			}
		}
		if prefixLen == 0 {
			break
		}
	}

	if prefixLen == 0 {
		return fmt.Errorf("no common prefix found")
	}

	for i, p := range splittedPathes {
		suffix := p[prefixLen:]
		fileMap[i].Path = filepath.Join(suffix...)
	}
	return nil
}

func removeFiles(files []*os.File) {
	for _, f := range files {
		f.Close()
		os.Remove(f.Name())
	}
}
