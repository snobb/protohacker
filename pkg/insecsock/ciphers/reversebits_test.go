package ciphers_test

import (
	"testing"

	"github.com/matryer/is"

	"proto/pkg/insecsock/ciphers"
)

func TestReverseBits_DoUndo(t *testing.T) {
	tests := []struct {
		name string
		in   byte
		want byte
	}{
		{
			name: "should reverse 0 to zero",
			in:   0,
			want: 0,
		},
		{
			name: "should reverse 255 to 255",
			in:   255,
			want: 255,
		},
		{
			name: "should reverse 1 to 128",
			in:   1,
			want: 128,
		},
		{
			name: "should reverse 2 to 64",
			in:   2,
			want: 64,
		},
		{
			name: "should reverse 3 to 64+128",
			in:   3,
			want: 128 + 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)

			rb := ciphers.ReverseBits{}
			got := rb.Do(tt.in)
			is.Equal(got, tt.want)
			is.Equal(rb.Undo(got), tt.in)
		})
	}
}
