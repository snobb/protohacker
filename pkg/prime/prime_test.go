package prime_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"proto/pkg/prime"
	"proto/pkg/tcpserver"

	"github.com/matryer/is"
)

func TestHandle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOut string
	}{
		{
			name:    "should return valid response on valid request with Prime number",
			input:   `{"method":"isPrime","number":41}`,
			wantOut: `{"method":"isPrime","prime":true}`,
		},
		{
			name:    "should return valid response on valid request with not a Prime number",
			input:   `{"method":"isPrime","number":42}`,
			wantOut: `{"method":"isPrime","prime":false}`,
		},
		{
			name: "should return valid response on valid request with pipelined requests",
			input: fmt.Sprintf("%s\n%s",
				`{"method":"isPrime","number":42}`,
				`{"method":"isPrime","number":41}`),
			wantOut: fmt.Sprintf("%s\n%s",
				`{"method":"isPrime","prime":false}`,
				`{"method":"isPrime","prime":true}`),
		},
		{
			name:    "should return invalid response on invalid request",
			input:   `{"method":"isPrime","number":"42"}`,
			wantOut: `{"method":"error","prime":false}`,
		},
		{
			name: "should return invalid response on invalid request with pipelined requests",
			input: fmt.Sprintf("%s\n%s",
				`{"method":"isPrime","number":true}`,
				`{"method":"isPrime","number":41}`),
			wantOut: fmt.Sprintf("%s\n%s",
				`{"method":"error","prime":false}`,
				`{"method":"isPrime","prime":true}`),
		},
		{
			name:    "should return invalid response when method is missing",
			input:   `{"number":"42"}`,
			wantOut: `{"method":"error","prime":false}`,
		},
		{
			name:    "should return invalid response when number is missing",
			input:   `{"method":"isPrime"}`,
			wantOut: `{"method":"error","prime":false}`,
		},
		{
			name:    "should return invalid response on invalid JSON",
			input:   `{"method":"isPr`,
			wantOut: `{"method":"error","prime":false}`,
		},
		{
			name:    "should return invalid response on blank request",
			input:   `{}`,
			wantOut: `{"method":"error","prime":false}`,
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
			prime.Handle(context.Background(), conn)

			is.Equal(conn.Recorder.Out.String(), fmt.Sprintf("%s\n", tt.wantOut))
		})
	}
}
