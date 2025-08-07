package compressing

import (
	"bufio"
	"compressor/internal/utiles"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"golang.org/x/sync/errgroup"
)

type ErrCompression struct{ Cause error }

func (e *ErrCompression) Error() string { return fmt.Sprintf("Compression failed: %v", e.Cause) }
func (e *ErrCompression) Unwrap() error { return e.Cause }

type CompressionBase interface {
	CompressorData() (name string, data Body)
	CompressFile(src io.Reader, dst io.Writer, prog *utiles.Progress[int64]) (size int64, err error)
}

type SimpleCompressor interface {
	Preprocessing(srcs []io.Reader) error
	CompressionBase
}

type FastCompressor interface {
	Preprocessing(srcs []io.Reader) (sizes []int64, err error)
	CompressionBase
}

// CompressFiles compresses the given files using the specified compressor.
func CompressFiles(
	c CompressionBase, pathes []string, dst *os.File, prog *utiles.Progress[int64],
) (contentSize int64, footerSize int64, err error) {
	srcs, err := utiles.OpenFiles(pathes...)
	if err != nil {
		return 0, 0, err
	}
	defer utiles.CloseFiles(srcs)

	var fileMap []File
	switch comp := c.(type) {
	case FastCompressor:
		if fileMap, err = fastCompress(comp, srcs, dst, prog); err != nil {
			return 0, 0, &ErrCompression{err}
		}
	case SimpleCompressor:
		if fileMap, err = simpleCompress(comp, srcs, dst, prog); err != nil {
			return 0, 0, &ErrCompression{err}
		}
	default:
		return 0, 0, fmt.Errorf("unsupported compressor type")
	}

	if err := formatPathes(fileMap); err != nil {
		return 0, 0, &ErrCompression{err}
	}
	name, data := c.CompressorData()
	footer := newFooter(name, fileMap, data)

	footerSize, err = writeFooter(footer, dst)
	if err != nil {
		return 0, 0, &ErrCompression{err}
	}

	// Write footer size at the end of the file
	if err = binary.Write(dst, binary.LittleEndian, footerSize); err != nil {
		return 0, 0, &ErrCompression{fmt.Errorf("error while writing footer size: %v", err)}
	}

	lastFile := fileMap[len(fileMap)-1]
	contentSize = lastFile.Offset + lastFile.Size
	return contentSize, footerSize, nil
}

// compress handles compression for SimpleCompressor implementations.
func simpleCompress(
	c SimpleCompressor, srcs []*os.File, dst *os.File, prog *utiles.Progress[int64],
) ([]File, error) {
	bufReaders := make([]*bufio.Reader, len(srcs))
	readers := make([]io.Reader, len(srcs))

	for i, f := range srcs {
		bufReaders[i] = bufio.NewReader(f)
		readers[i] = bufReaders[i]
	}

	if err := c.Preprocessing(readers); err != nil {
		return nil, &ErrCompression{err}
	}

	var (
		offset  int64
		fileMap = make([]File, len(srcs))
	)

	for i, f := range srcs {
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		bufReaders[i].Reset(f)

		size, err := c.CompressFile(bufReaders[i], dst, prog)
		if err != nil {
			return nil, &ErrCompression{err}
		}

		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return nil, &ErrCompression{err}
		}
		checksum, _ := checksum(f)

		fileMap[i] = File{
			Path:     f.Name(),
			Checksum: checksum,
			Offset:   offset,
			Size:     size,
		}
		offset += size
	}
	return fileMap, nil
}

// fastCompress handles compression for FastCompressor implementations
// with concurrent writes.
func fastCompress(
	c FastCompressor, srcs []*os.File, dst *os.File, prog *utiles.Progress[int64],
) ([]File, error) {
	bufReaders := make([]*bufio.Reader, len(srcs))
	readers := make([]io.Reader, len(srcs))

	for i, f := range srcs {
		bufReaders[i] = bufio.NewReader(f)
		readers[i] = bufReaders[i]
	}

	sizes, err := c.Preprocessing(readers)
	if err != nil {
		return nil, err
	}

	var (
		offset  int64
		fileMap = make([]File, len(srcs))
	)

	for i, f := range srcs {
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		checksum, _ := checksum(f)

		fileMap[i] = File{
			Path:     f.Name(),
			Checksum: checksum,
			Offset:   offset,
			Size:     sizes[i],
		}
		offset += sizes[i]
	}

	if err := dst.Truncate(offset); err != nil {
		return nil, err
	}

	var eg errgroup.Group
	for i, f := range srcs {
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		bufReaders[i].Reset(f)
		writer := io.NewOffsetWriter(dst, fileMap[i].Offset)
		eg.Go(func() error {
			_, err := c.CompressFile(bufReaders[i], writer, prog)
			return err
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if _, err := dst.Seek(0, io.SeekEnd); err != nil {
		return nil, err
	}
	return fileMap, nil
}
