package app

import (
	"testing"

	"github.com/matryer/is"
)

func TestHandleLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "should choose the the only choice",
			line: "5x car",
			want: "5x car",
		},
		{
			name: "should choose the the max choice",
			line: "4x dog,5x car",
			want: "5x car",
		},
		{
			name: "should choose the the max choice regardless of order",
			line: "3x rat,2x cat",
			want: "3x rat",
		},
		{
			name: "should try return a valid entry",
			line: "xx rat,2x cat",
			want: "2x cat",
		},
		{
			name: "should try return a valid entry",
			line: "xx rat,yy cat",
			want: "xx rat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			got := HandleLine(tt.line)
			is.Equal(got, tt.want)
		})
	}
}
