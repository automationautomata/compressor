package compressing

import (
	"fmt"
	"io"
)

type ErrHeaderRead struct {
	Cause error
}

func (e *ErrHeaderRead) Error() string {
	return fmt.Sprintf("Header can't be readed: %s", e.Cause.Error())
}

func (e *ErrHeaderRead) Unwrap() error {
	return e.Cause
}

type Decompressor interface {
	HeaderDataType() HeaderData
	Preprocessing(srcPath io.ReadSeeker) error
	DecompressFile(dd *DecompressionInput, progreessChan chan int64) error
}

type DecompressionInput struct {
	Data       HeaderData
	SourceFile io.ReadSeeker
	DestFile   io.WriteSeeker
}

type DecompressionOutput struct {
	OrigName    string
	OldCheckSum string
}

type DecompressorFactory func(compType string) (d Decompressor)

func Decompress(
	decompFactory DecompressorFactory,
	srcFile io.ReadSeeker,
	dstFile io.WriteSeeker,
	progressChan chan int64,
) (origName string, oldCheckSum string, err error) {

	info, infoSize, err := readHeaderInfo(srcFile)
	if err != nil {
		return "", "", &ErrHeaderRead{err}
	}

	decomp := decompFactory(info.Type)
	data, dataSize, err := readHeaderData(srcFile, decomp)
	if err != nil {
		return "", "", &ErrHeaderRead{err}
	}

	if err := decomp.Preprocessing(srcFile); err != nil {
		return "", "", err
	}
	srcFile.Seek(infoSize+dataSize, io.SeekStart)
	if _, ok := <-progressChan; ok {
		progressChan <- infoSize + dataSize
	}

	err = decomp.DecompressFile(&DecompressionInput{data, srcFile, dstFile}, progressChan)
	if err != nil {
		return "", "", err
	}
	return info.OrigName, info.CheckSum, nil
}

func readHeaderInfo(file io.ReadSeeker) (*HeaderInfo, int64, error) {
	info, size, err := ReadHeaderInfo(file)
	if err != nil {
		return nil, -1, err
	}
	return info, size, nil
}

func readHeaderData(file io.ReadSeeker, decomp Decompressor) (HeaderData, int64, error) {
	data := decomp.HeaderDataType()
	size, err := ReadHeaderData(file, data)
	if err != nil {
		return nil, -1, err
	}
	return data, size, nil
}
