package frame

import (
	"bytes"
	"fmt"
)

// Error represents the Error message
type Error struct {
	msg string
}

// NewError creates new Error message
func NewError(err error) *Error {
	return &Error{msg: err.Error()}
}

// Kind == KindError
func (er *Error) Kind() uint8 {
	return KindError
}

// Read loads the contents of the message from byte buffer
func (er *Error) Read(data []byte) error {
	buf := bytes.NewReader(data)

	msg, err := ReadString(buf)
	if err != nil {
		return err
	}
	er.msg = msg

	if buf.Len() > 0 {
		return fmt.Errorf("invalid message - too much payload")
	}

	return nil
}

// Write puts the contents of the message into a byte buffer.
func (er *Error) Write() ([]byte, error) {
	var buf bytes.Buffer

	if err := WriteString(&buf, er.msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// String implements fmt.Stringer interface
func (er *Error) String() string {
	return fmt.Sprintf("Error:%s", er.msg)
}

// Error implements error interface
func (er Error) Error() string {
	return fmt.Sprintf("Error:%s", er.msg)
}
