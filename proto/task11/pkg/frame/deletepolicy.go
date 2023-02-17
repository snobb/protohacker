package frame

import (
	"bytes"
	"fmt"
)

// DeletePolicy represents the Delete policy message
type DeletePolicy struct {
	Policy uint32
}

// NewDeletePolicy create new instance of DeletePolicy with given policy id
func NewDeletePolicy(policy uint32) *DeletePolicy {
	return &DeletePolicy{Policy: policy}
}

// Kind == DeletePolicy
func (dp *DeletePolicy) Kind() uint8 {
	return KindDeletePolicy
}

// Write puts the contents of the message into a byte buffer.
func (dp *DeletePolicy) Write() ([]byte, error) {
	var buf bytes.Buffer

	if err := WriteU32(&buf, dp.Policy); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// String implements fmt.Stringer
func (dp *DeletePolicy) String() string {
	return fmt.Sprintf("DeletePolicy:%d", dp.Policy)
}
