package echo_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/matryer/is"

	"proto/common/pkg/tcpserver"
	"proto/task00/pkg/echo"
)

func TestHandle(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "should echo the received payload",
			input: "foo bar baz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			conn := &tcpserver.TestConn{
				Recorder: tcpserver.Recorder{
					In: bytes.NewBufferString(tt.input),
				},
			}
			echo.Handle(context.Background(), conn)
			is.Equal(tt.input, conn.Recorder.Out.String())
		})
	}
}
