package frame

import (
	"bytes"
	"fmt"
)

// Target represents target counts for species.
type Target struct {
	Min, Max uint32
}

// TargetPopulations represents the target counts per species for a site.
type TargetPopulations struct {
	Site    uint32
	Targets map[string]Target
}

// NewTargetPopulations creates a new TargetPopulations message
func NewTargetPopulations() *TargetPopulations {
	return &TargetPopulations{}
}

// Kind = KindTargetPopulation
func (tp *TargetPopulations) Kind() uint8 {
	return KindTargetPopulation
}

// Read loads the contents of the message from byte buffer
func (tp *TargetPopulations) Read(data []byte) error {
	buf := bytes.NewReader(data)
	var err error

	if tp.Site, err = ReadU32(buf); err != nil {
		return err
	}

	size, err := ReadU32(buf)
	if err != nil {
		return err
	}

	tp.Targets = make(map[string]Target)
	for i := 0; i < int(size); i++ {
		tgt := Target{}

		name, err := ReadString(buf)
		if err != nil {
			return err
		}

		if tgt.Min, err = ReadU32(buf); err != nil {
			return err
		}

		if tgt.Max, err = ReadU32(buf); err != nil {
			return err
		}

		tp.Targets[name] = tgt
	}

	if buf.Len() > 0 {
		return fmt.Errorf("invalid message - too much payload")
	}

	return nil
}

// String implements fmt.Stringer interface
func (tp *TargetPopulations) String() string {
	return fmt.Sprintf("TargetPopulations:%d:%#v", tp.Site, tp.Targets)
}
