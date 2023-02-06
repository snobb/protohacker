package ciphers_test

import (
	"testing"

	"github.com/matryer/is"

	"proto/task08/pkg/insecsock/ciphers"
)

func TestXor_DoUndo(t *testing.T) {
	tests := []struct {
		name string
		n    byte
		in   byte
		want byte
	}{
		{
			name: "should convert byte",
			n:    255,
			in:   0,
			want: 255,
		},
		{
			name: "should convert byte",
			n:    0,
			in:   255,
			want: 255,
		},
		{
			name: "should convert byte",
			n:    0b10101010,
			in:   0b01010101,
			want: 255,
		},
		{
			name: "should convert byte",
			n:    0b11111111,
			in:   0b10000000,
			want: 0b01111111,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)

			xor := ciphers.Xor{N: tt.n}
			got := xor.Do(tt.in)
			is.Equal(got, tt.want)
			is.Equal(xor.Undo(got), tt.in)
		})
	}
}

func TestXorPos_DoUndo(t *testing.T) {
	tests := []struct {
		name string
		pos  byte
		in   byte
		want byte
	}{
		{
			name: "should convert byte",
			pos:  255,
			in:   0,
			want: 255,
		},
		{
			name: "should convert byte",
			pos:  0,
			in:   255,
			want: 255,
		},
		{
			name: "should convert byte",
			pos:  0b10101010,
			in:   0b01010101,
			want: 255,
		},
		{
			name: "should convert byte",
			pos:  0b11111111,
			in:   0b10000000,
			want: 0b01111111,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)

			xor := ciphers.XorPos{}
			got := xor.Do(tt.in, tt.pos)
			is.Equal(got, tt.want)
			is.Equal(xor.Undo(got, tt.pos), tt.in)
		})
	}
}
