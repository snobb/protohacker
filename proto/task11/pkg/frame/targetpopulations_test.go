package frame_test

import (
	"proto/task11/pkg/frame"
	"testing"

	"github.com/matryer/is"
)

func TestTargetPopulation_Read(t *testing.T) {
	is := is.New(t)

	tests := map[string]struct {
		payload  []byte
		wantSite uint32
		wantPops int
		checkFn  func(s *frame.TargetPopulations)
		wantErr  bool
	}{
		"Should read TargetPopulations correctly": {
			payload: []byte{
				0x00, 0x00, 0x30, 0x39, // site: 12345,
				0x00, 0x00, 0x00, 0x02, // populations: (length 2) [
				0x00, 0x00, 0x00, 0x03, //   species: (length 3)
				0x64, 0x6f, 0x67, //           "dog",
				0x00, 0x00, 0x00, 0x01, //       min: 1,
				0x00, 0x00, 0x00, 0x03, //       max: 3,
				0x00, 0x00, 0x00, 0x03, //   species: (length 3)
				0x72, 0x61, 0x74, //           "rat",
				0x00, 0x00, 0x00, 0x00, //       min: 0,
				0x00, 0x00, 0x00, 0x0a, //       max: 10,
			},
			checkFn: func(tp *frame.TargetPopulations) {
				is.Equal(tp.Site, uint32(12345))
				is.Equal(len(tp.Targets), 2)
				is.Equal(tp.Targets["dog"].Min, uint32(1))
				is.Equal(tp.Targets["dog"].Max, uint32(3))
				is.Equal(tp.Targets["rat"].Min, uint32(0))
				is.Equal(tp.Targets["rat"].Max, uint32(10))
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var tp frame.TargetPopulations
			err := tp.Read(tt.payload)
			if tt.wantErr && err != nil {
				return
			}

			is.NoErr(err)
			tt.checkFn(&tp)
		})
	}
}
