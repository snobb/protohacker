package frame

import (
	"bytes"
	"fmt"
)

// DialAuth represents a message setting site for the authority session.
type DialAuth struct {
	Site uint32
}

// NewDialAuth create new instance of DialAuth with site id
func NewDialAuth(site uint32) *DialAuth {
	return &DialAuth{Site: site}
}

// Kind == DialAuth
func (da *DialAuth) Kind() uint8 {
	return KindDialAuthority
}

// Write puts the contents of the message into a byte buffer.
func (da *DialAuth) Write() ([]byte, error) {
	var buf bytes.Buffer
	if err := WriteU32(&buf, da.Site); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// String implements fmt.Stringer
func (da *DialAuth) String() string {
	return fmt.Sprintf("DialAuth:%d", da.Site)
}
