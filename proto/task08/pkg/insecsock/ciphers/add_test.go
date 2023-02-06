package ciphers_test

import (
	"testing"

	"github.com/matryer/is"

	"proto/task08/pkg/insecsock/ciphers"
)

func TestAdd_DoUndo(t *testing.T) {
	tests := []struct {
		name string
		n    byte
		in   byte
		want byte
	}{
		{
			name: "should add to a byte",
			n:    5,
			in:   0,
			want: 5,
		},
		{
			name: "should wrap around correctly",
			n:    5,
			in:   253,
			want: 2,
		},
		{
			name: "should wrap around correctly",
			n:    15,
			in:   250,
			want: 9,
		},
		{
			name: "should do nothing on N == 0",
			n:    0,
			in:   250,
			want: 250,
		},
		{
			name: "should wrap around correctly",
			n:    255,
			in:   250,
			want: 249,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			a := ciphers.Add{N: tt.n}
			got := a.Do(tt.in)
			is.Equal(got, tt.want)       // Do
			is.Equal(a.Undo(got), tt.in) // Undo
		})
	}
}

func TestAddPos_DoUndo(t *testing.T) {
	tests := []struct {
		name string
		pos  byte
		in   byte
		want byte
	}{
		{
			name: "should add to a byte",
			pos:  5,
			in:   0,
			want: 5,
		},
		{
			name: "should wrap around correctly",
			pos:  5,
			in:   253,
			want: 2,
		},
		{
			name: "should wrap around correctly",
			pos:  15,
			in:   250,
			want: 9,
		},
		{
			name: "should do nothing on N == 0",
			pos:  0,
			in:   250,
			want: 250,
		},
		{
			name: "should wrap around correctly",
			pos:  255,
			in:   250,
			want: 249,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			a := ciphers.AddPos{}
			got := a.Do(tt.in, tt.pos)
			is.Equal(got, tt.want)               // Do
			is.Equal(a.Undo(got, tt.pos), tt.in) // Undo
		})
	}
}
