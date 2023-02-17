package frame

import (
	"encoding/binary"
	"io"
)

func ReadU8(r io.Reader) (uint8, error) {
	var data uint8
	err := binary.Read(r, binary.BigEndian, &data)
	return data, err
}

func ReadU32(r io.Reader) (uint32, error) {
	var data uint32
	err := binary.Read(r, binary.BigEndian, &data)
	return data, err
}

func ReadBytes(r io.Reader, sz uint32) ([]byte, error) {
	data := make([]byte, sz)
	err := binary.Read(r, binary.BigEndian, &data)
	return data, err
}

func ReadString(r io.Reader) (string, error) {
	sz, err := ReadU32(r)
	if err != nil {
		return "", err
	}

	data := make([]byte, sz)
	err = binary.Read(r, binary.BigEndian, &data)
	return string(data), err
}

func WriteU8(w io.Writer, data uint8) error {
	return binary.Write(w, binary.BigEndian, data)
}

func WriteU32(w io.Writer, data uint32) error {
	return binary.Write(w, binary.BigEndian, data)
}

func WriteString(w io.Writer, data string) error {
	if err := WriteU32(w, uint32(len(data))); err != nil {
		return err
	}

	return binary.Write(w, binary.BigEndian, []byte(data))
}

func WriteBytes(w io.Writer, data []byte) error {
	return binary.Write(w, binary.BigEndian, data)
}
