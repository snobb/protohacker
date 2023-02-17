package frame

import (
	"bytes"
	"fmt"
)

// PolicyResult represents a result of creating a policy
type PolicyResult struct {
	Policy uint32
}

// NewPolicyResult creates a new instance of PolicyResult message
func NewPolicyResult() *PolicyResult {
	return &PolicyResult{}
}

// Kind == KindPolicyResult
func (pr *PolicyResult) Kind() uint8 {
	return KindPolicyResult
}

// Read loads the contents of the message from byte buffer
func (pr *PolicyResult) Read(data []byte) error {
	buf := bytes.NewReader(data)
	var err error

	pr.Policy, err = ReadU32(buf)
	if err != nil {
		return err
	}

	if buf.Len() > 0 {
		return fmt.Errorf("PolicyResult: invalid message - too much payload")
	}

	return nil
}

// String implements fmt.Stringer interface
func (pr *PolicyResult) String() string {
	return fmt.Sprintf("Policy:%d", pr.Policy)
}
