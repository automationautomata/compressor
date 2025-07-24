package compressing

import (
	"fmt"
	"io"
)

type ErrHeaderWrite struct {
	Cause error
}

func (e *ErrHeaderWrite) Error() string {
	return fmt.Sprintf("Header can't be written: %s", e.Cause.Error())
}

func (e *ErrHeaderWrite) Unwrap() error {
	return e.Cause
}

type Compressor interface {
	Preprocessing(srcFile io.ReadSeeker) error
	GetHeader() *Header
	CompressFile(srcPath io.ReadSeeker, dstPath io.WriteSeeker, progreessChan chan int64) error
}

func Compress(
	c Compressor, srcFile io.ReadSeeker, dstFile io.WriteSeeker, progreessChan chan int64,
) (headerSize int64, err error) {
	if err = c.Preprocessing(srcFile); err != nil {
		return -1, err
	}
	srcFile.Seek(0, io.SeekStart)

	size, err := WriteHeader(c.GetHeader(), dstFile)
	if err != nil {
		return -1, &ErrHeaderWrite{err}
	}

	err = c.CompressFile(srcFile, dstFile, progreessChan)
	return size, err
}
