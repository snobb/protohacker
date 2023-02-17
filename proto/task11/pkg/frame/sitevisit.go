package frame

import (
	"bytes"
	"fmt"
)

// SiteVisit represents counts of species per site.
type SiteVisit struct {
	Site        uint32
	Populations map[string]uint32
}

// NewSiteVisit creates a new sitevisit message with defaults
func NewSiteVisit() *SiteVisit {
	return &SiteVisit{}
}

// Kind == KindSiteVisit
func (sv *SiteVisit) Kind() uint8 {
	return KindSiteVisit
}

// Read loads the contents of the message from byte buffer
func (sv *SiteVisit) Read(data []byte) error {
	buf := bytes.NewReader(data)
	var err error

	if sv.Site, err = ReadU32(buf); err != nil {
		return err
	}

	size, err := ReadU32(buf)
	if err != nil {
		return err
	}

	// one population record is at least 8bytes (with empty string), so size*8 must be less than
	// the size of the payload.
	if int(size)*8 > len(data)-8 {
		return fmt.Errorf("SiteVisit: size too big")
	}

	sv.Populations = make(map[string]uint32)
	for i := 0; i < int(size); i++ {
		name, err := ReadString(buf)
		if err != nil {
			return fmt.Errorf("SiteVisit: ReadString: %w", err)
		}

		count, err := ReadU32(buf)
		if err != nil {
			return fmt.Errorf("SiteVisit: ReadU32: %w", err)
		}

		if oldcnt, ok := sv.Populations[name]; ok && count != oldcnt {
			return fmt.Errorf("Sitevisit: conflicting counts")
		}

		sv.Populations[name] = count
	}

	if buf.Len() > 0 {
		return fmt.Errorf("SiteVisit: invalid message - too much payload")
	}

	return nil
}

// String implements fmt.Stringer interface
func (sv *SiteVisit) String() string {
	return fmt.Sprintf("SiteVisit:%d:%#v", sv.Site, sv.Populations)
}
