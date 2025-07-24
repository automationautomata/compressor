package compressing

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
)

type HeaderInfo struct {
	OrigName string
	Type     string
	CheckSum string
}

type HeaderData any

type Header struct {
	Info HeaderInfo
	Data HeaderData
}

func NewHeader(origName, checkSum, compType string, data HeaderData) *Header {
	return &Header{
		HeaderInfo{
			OrigName: origName,
			Type:     compType,
			CheckSum: checkSum,
		},
		data,
	}
}

func write(headerPart any, file io.Writer) (size int64, err error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	gob.Register(&headerPart)
	if err := encoder.Encode(headerPart); err != nil {
		return -1, err
	}
	size = int64(buffer.Len())
	if err := binary.Write(file, binary.LittleEndian, size); err != nil {
		return -1, err
	}
	if err := binary.Write(file, binary.LittleEndian, buffer.Bytes()); err != nil {
		return -1, err
	}
	return size + 8, nil
}

func WriteHeader(h *Header, file io.WriteSeeker) (size int64, err error) {
	infoSize, err := write(h.Info, file)
	if err != nil {
		return -1, err
	}
	dataSize, err := write(h.Data, file)
	if err != nil {
		return -1, err
	}
	return infoSize + dataSize, nil
}

func read(file io.Reader, headerDataType any) (size int64, err error) {
	var headerSize int64
	if err = binary.Read(file, binary.LittleEndian, &headerSize); err != nil {
		return -1, err
	}

	headerBuf := make([]byte, headerSize)
	if err = binary.Read(file, binary.LittleEndian, headerBuf); err != nil {
		return -1, err
	}

	dec := gob.NewDecoder(bytes.NewBuffer(headerBuf))
	gob.Register(headerDataType)
	if err = dec.Decode(headerDataType); err != nil {
		return -1, err
	}

	return headerSize, nil
}

func ReadHeaderData(file io.ReadSeeker, headerDataType HeaderData) (size int64, err error) {
	return read(file, headerDataType)
}

func ReadHeaderInfo(file io.ReadSeeker) (info *HeaderInfo, size int64, err error) {
	info = &HeaderInfo{}
	size, err = read(file, info)
	if err != nil {
		return nil, -1, err
	}
	return info, size, nil
}
