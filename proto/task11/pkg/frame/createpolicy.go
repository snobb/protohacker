package frame

import (
	"bytes"
	"fmt"
)

const (
	ActionCull     = 0x90
	ActionConserve = 0xa0
)

// Action represents a policy action
type Action uint8

// String implements fmt.Stringer for Action
func (ac Action) String() string {
	switch ac {
	case ActionCull:
		return "cull"
	case ActionConserve:
		return "conserve"
	default:
		return "invalid"
	}
}

// CreatePolicy represents the Create policy message
type CreatePolicy struct {
	Species string
	Action  Action
}

// NewCreatePolicy create new instance of CreatePolicy with species and action
func NewCreatePolicy(species string, action Action) *CreatePolicy {
	return &CreatePolicy{Species: species, Action: action}
}

// Kind == CreatePolicy
func (cp *CreatePolicy) Kind() uint8 {
	return KindCreatePolicy
}

// Write puts the contents of the message into a byte buffer.
func (cp *CreatePolicy) Write() ([]byte, error) {
	var buf bytes.Buffer

	if err := WriteString(&buf, cp.Species); err != nil {
		return nil, err
	}

	if err := WriteU8(&buf, uint8(cp.Action)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// String implements fmt.Stringer
func (cp *CreatePolicy) String() string {
	return fmt.Sprintf("CreatePolicy:%s:%s", cp.Species, cp.Action)
}
