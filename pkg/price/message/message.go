package message

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	TypeInsert = byte('I')
	TypeQuery  = byte('Q')

	MsgLength = 9
)

type Payload struct {
	Time int32
	Data int32
}

type Msg struct {
	Type    byte
	Payload Payload
}

func (m *Msg) ReadFrom(in io.Reader) (int64, error) {
	if err := binary.Read(in, binary.BigEndian, m); err != nil {
		if errors.Is(err, io.EOF) {
			return 0, err // EOF
		}

		return 0, errors.New("Corrupted data")
	}

	return MsgLength, nil
}

func (m *Msg) Len() int64 {
	return MsgLength
}
