package frame_test

import (
	"proto/task11/pkg/frame"
	"testing"

	"github.com/matryer/is"
)

func TestSiteVisit_Read(t *testing.T) {
	is := is.New(t)

	tests := map[string]struct {
		payload  []byte
		wantSite uint32
		checkFn  func(sv *frame.SiteVisit)
		wantErr  bool
	}{
		"Should read SiteVisit correctly": {
			payload: []byte{
				0x00, 0x00, 0x30, 0x39, // site: 12345,
				0x00, 0x00, 0x00, 0x02, // populations: (length 2) [
				0x00, 0x00, 0x00, 0x03, //   species: (length 3)
				0x64, 0x6f, 0x67, //           "dog",
				0x00, 0x00, 0x00, 0x01, //       count: 1,
				0x00, 0x00, 0x00, 0x03, //   species: (length 3)
				0x72, 0x61, 0x74, //           "rat",
				0x00, 0x00, 0x00, 0x05, //       count: 5,
			},
			checkFn: func(sv *frame.SiteVisit) {
				is.Equal(sv.Site, uint32(12345))
				is.Equal(len(sv.Populations), 2)
				is.Equal(sv.Populations["dog"], uint32(1))
				is.Equal(sv.Populations["rat"], uint32(5))
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var sv frame.SiteVisit
			err := sv.Read(tt.payload)
			if tt.wantErr && err != nil {
				return
			}

			is.NoErr(err)
			tt.checkFn(&sv)
		})
	}
}
