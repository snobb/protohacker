package frame

import (
	"bytes"
	"fmt"
)

const (
	proto          = "pestcontrol"
	version uint32 = 1
)

// Hello represents hello message for handshake.
type Hello struct {
	proto   string
	version uint32
}

// NewHello creates a new hello message with defaults
func NewHello() *Hello {
	return &Hello{}
}

// Kind == KindHello
func (h *Hello) Kind() uint8 {
	return KindHello
}

// Read loads the contents of the message from byte buffer
func (he *Hello) Read(p []byte) error {
	buf := bytes.NewReader(p)

	var err error
	he.proto, err = ReadString(buf)
	if err != nil {
		return err
	}

	he.version, err = ReadU32(buf)
	if err != nil {
		return err
	}

	if buf.Len() > 0 {
		return fmt.Errorf("invalid message - too much payload")
	}

	if err := he.validate(); err != nil {
		return err
	}

	return nil
}

// Write puts the contents of the message into a byte buffer.
func (he *Hello) Write() ([]byte, error) {
	var buf bytes.Buffer

	he.proto = proto
	he.version = version

	if err := WriteString(&buf, proto); err != nil {
		return nil, err
	}

	if err := WriteU32(&buf, version); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// String implements fmt.Stringer interface
func (he *Hello) String() string {
	return fmt.Sprintf("hello:%s:%d", he.proto, he.version)
}

func (he *Hello) validate() error {
	if he.proto != proto {
		return fmt.Errorf("error: proto != 'pestcontrol': got %s", he.proto)
	}

	if he.version != version {
		return fmt.Errorf("error: version != 1: got %d", he.version)
	}

	return nil
}
