package compressing

import (
	"compressor/internal/utiles"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type ErrDecompression struct{ Cause error }

func (e *ErrDecompression) Error() string { return fmt.Sprintf("Decompression failed: %v", e.Cause) }
func (e *ErrDecompression) Unwrap() error { return e.Cause }

type Decompressor interface {
	FooterBodyType() Body
	Preprocessing(data Body, src io.ReadSeeker) error
	DecompressFile(dd *DecompressionInput, prog *utiles.Progress[int64]) error
}

type DecompressionInput struct {
	Body       Body
	SourceFile io.Reader
	DestFile   io.Writer
}

type DecompressedFile struct {
	Path        string
	OldChecksum string
	NewChecksum string
}

type DecompressorFactory func(compType string) (d Decompressor)

func Decompress(
	factory DecompressorFactory, src *os.File, dstpath string, prog *utiles.Progress[int64],
) ([]*DecompressedFile, error) {
	md, mdSize, err := ReadFooterMetadata(src)
	if err != nil {
		return nil, &ErrDecompression{err}
	}
	prog.Write(mdSize)

	decomp := factory(md.Type)

	body := decomp.FooterBodyType()
	bodySize, err := readFooterBody(src, body)
	if err != nil {
		return nil, &ErrDecompression{err}
	}
	prog.Write(bodySize)

	if err := decomp.Preprocessing(body, src); err != nil {
		return nil, err
	}

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	files := make([]*os.File, len(md.FileMap))
	for i, file := range md.FileMap {
		dir := filepath.Join(dstpath, filepath.Dir(file.Path))
		if err := os.MkdirAll(dir, 0755); err != nil {
			removeFiles(files[:i])
			return nil, &ErrDecompression{err}
		}
		path := filepath.Join(dstpath, file.Path)
		files[i], err = os.Create(path)
	}
	defer utiles.CloseFiles(files)

	if _, err := src.Seek(bodySize+mdSize, io.SeekStart); err != nil {
		return nil, err
	}
	for i, f := range md.FileMap {
		reader := io.NewSectionReader(src, f.Offset, f.Size)
		input := &DecompressionInput{body, reader, files[i]}
		if err = decomp.DecompressFile(input, prog); err != nil {
			removeFiles(files)
			return nil, &ErrDecompression{err}
		}
	}

	output := make([]*DecompressedFile, len(files))
	for i, f := range files {
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			removeFiles(files)
			return nil, err
		}

		oldChecksum := md.FileMap[i].Checksum
		newChecksum, err := checksum(f)
		if err != nil {
			removeFiles(files)
			return nil, &ErrDecompression{err}
		}
		output[i] = &DecompressedFile{f.Name(), oldChecksum, newChecksum}
	}
	return output, nil
}

func ReadFooterMetadata(file io.ReadSeeker) (md *Metadata, size int64, err error) {
	if _, err := file.Seek(-8, io.SeekEnd); err != nil {
		return nil, 0, &ErrDecompression{fmt.Errorf("error while reading footer size: %v", err)}
	}
	footerSize := int64(0)
	if err = binary.Read(file, binary.LittleEndian, &footerSize); err != nil {
		return nil, 0, &ErrFooterRead{err}
	}

	if _, err = file.Seek(-footerSize-8, io.SeekEnd); err != nil {
		return nil, 0, &ErrFooterRead{err}
	}
	md, size, err = readFooterMetadata(file)
	if err != nil {
		return nil, 0, &ErrFooterRead{err}
	}
	return md, size, nil
}
