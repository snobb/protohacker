package message

import (
	"encoding/binary"
	"errors"
	"io"
)

// Constants
const (
	TypeInsert = byte('I')
	TypeQuery  = byte('Q')

	MsgLength = 9
)

// Payload stores the Payload data.
type Payload struct {
	Time int32
	Data int32
}

// Msg represends the message storage.
type Msg struct {
	Type    byte
	Payload Payload
}

// ReadFrom reads a message from the provided io.Reader
func (m *Msg) ReadFrom(in io.Reader) (int64, error) {
	if err := binary.Read(in, binary.BigEndian, m); err != nil {
		if errors.Is(err, io.EOF) {
			return 0, err // EOF - only if no data was read as per binary.Read docs.
		}

		return 0, errors.New("Corrupted data")
	}

	return MsgLength, nil
}

// Len return the length of the message
func (m *Msg) Len() int64 {
	return MsgLength
}
