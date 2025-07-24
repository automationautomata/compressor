package compressing

import (
	"crypto/sha256"
	"io"
)

func CalcCheckSum(r io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return "", err
	}
	return string(hasher.Sum(nil)), nil
}
