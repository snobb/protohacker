package app_test

import (
	"testing"

	"github.com/matryer/is"

	"proto/task07/pkg/lrcp/app"
)

func TestReverser_Reverse(t *testing.T) {
	tests := []struct {
		name    string
		data    []string
		want    string
		remains string // needs to be reversed and with trailing new line
	}{
		{
			name: "Should parse and reverse a single line in multiple buffers",
			data: []string{
				"hello ",
				"world\n",
			},
			want: "dlrow olleh\n",
		},
		{
			name: "Should parse and reverse a multiple line in multiple buffers",
			data: []string{
				"hello ",
				"world\n",
				"foo ",
				"bar",
			},
			want:    "dlrow olleh\n",
			remains: "rab oof\n",
		},
		{
			name: "Should parse and reverse a multiple line in multiple buffers",
			data: []string{
				"hello ",
				"world\n",
				"bazbaz\n",
				"foo ",
				"bar",
			},
			want:    "dlrow olleh\nzabzab\n",
			remains: "rab oof\n",
		},

		{
			name: "Should parse and reverse a multiple line in multiple buffers with tracing newline",
			data: []string{
				"hello ",
				"world",
				"\n",
				"foo",
				"bar\n",
			},
			want: "dlrow olleh\nraboof\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)

			app := app.App{}

			for _, data := range tt.data {
				app.Write([]byte(data))
			}

			buf := make([]byte, 100)
			n, _ := app.Read(buf)

			is.Equal(string(buf[:n]), tt.want)
		})
	}
}
