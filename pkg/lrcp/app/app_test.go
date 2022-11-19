package app_test

import (
	"context"
	"testing"

	"github.com/matryer/is"

	"proto/pkg/lrcp/app"
)

func TestReverser_Reverse(t *testing.T) {
	tests := []struct {
		name    string
		data    []string
		want    []string
		remains string // needs to be reversed and with trailing new line
	}{
		{
			name: "Should parse and reverse a single line in multiple buffers",
			data: []string{
				"hello ",
				"world\n",
			},
			want: []string{"dlrow olleh\n"},
		},
		{
			name: "Should parse and reverse a multiple line in multiple buffers",
			data: []string{
				"hello ",
				"world\n",
				"foo ",
				"bar",
			},
			want:    []string{"dlrow olleh\n"},
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
			want:    []string{"dlrow olleh\n", "zabzab\n"},
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
			want: []string{"dlrow olleh\n", "raboof\n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			app := app.New(ctx)

			for _, data := range tt.data {
				app.Write([]byte(data))
			}

			i := 0
			for line := range app.OutCh() {
				is.Equal(string(line), tt.want[i])
				i++

				if i >= len(tt.want) {
					break
				}
			}

			is.Equal(i, len(tt.want))

			if tt.remains != "" {
				app.Write([]byte("\n"))
				line := <-app.OutCh()

				is.Equal(string(line), tt.remains)
			}
		})
	}
}
