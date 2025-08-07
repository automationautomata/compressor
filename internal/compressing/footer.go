package compressing

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
)

type ErrFooterRead struct{ Cause error }

func (e *ErrFooterRead) Error() string { return fmt.Sprintf("footer can't be readed: %v", e.Cause) }
func (e *ErrFooterRead) Unwrap() error { return e.Cause }

type ErrFooterWrite struct{ Cause error }

func (e *ErrFooterWrite) Error() string { return fmt.Sprintf("footer can't be written: %v", e.Cause) }
func (e *ErrFooterWrite) Unwrap() error { return e.Cause }

type File struct {
	Path     string // relative path
	Checksum string
	Offset   int64
	Size     int64
}

type Metadata struct {
	Type    string
	FileMap []File
}

type Body any

type Footer struct {
	Metadata
	Body
}

func newFooter(compType string, fileMap []File, body Body) *Footer {
	return &Footer{Metadata{compType, fileMap}, body}
}

func write(data any, file io.Writer) (size int64, err error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	gob.Register(&data)
	if err := encoder.Encode(data); err != nil {
		return 0, err
	}

	size = int64(buffer.Len())
	if err := binary.Write(file, binary.LittleEndian, size); err != nil {
		return 0, err
	}

	if err := binary.Write(file, binary.LittleEndian, buffer.Bytes()); err != nil {
		return 0, err
	}
	return size + 8, nil
}

func writeFooter(footer *Footer, file io.Writer) (size int64, err error) {
	mdSize, err := write(footer.Metadata, file)
	if err != nil {
		return 0, &ErrFooterWrite{err}
	}
	bodySize, err := write(footer.Body, file)
	if err != nil {
		return 0, &ErrFooterWrite{err}
	}

	return bodySize + mdSize, nil
}

func read(file io.ReadSeeker, dataType any) (size int64, err error) {
	if err = binary.Read(file, binary.LittleEndian, &size); err != nil {
		return 0, err
	}

	buf := make([]byte, size)
	if err = binary.Read(file, binary.LittleEndian, buf); err != nil {
		return 0, err
	}

	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	gob.Register(dataType)
	if err = dec.Decode(dataType); err != nil {
		return 0, err
	}

	return size + 8, nil
}

func readFooterBody(file io.ReadSeeker, footerDataType Body) (size int64, err error) {
	return read(file, footerDataType)
}

func readFooterMetadata(file io.ReadSeeker) (md *Metadata, size int64, err error) {
	size, err = read(file, &md)
	if err != nil {
		return nil, 0, err
	}
	return md, size, nil
}
