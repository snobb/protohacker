package message_test

import (
	"bytes"
	"io"
	"proto/pkg/price/message"
	"testing"

	"github.com/matryer/is"
)

func TestMessage_ReadFrom(t *testing.T) {
	tests := []struct {
		name    string
		in      io.Reader
		wantMsg message.Msg
		wantErr bool
	}{
		{
			name: "should read data from a reader",
			in: bytes.NewReader([]byte{
				0x49,
				0x00, 0x00, 0x30, 0x39,
				0x00, 0x00, 0x00, 0x65}),
			wantMsg: message.Msg{
				Type: 'I',
				Payload: message.Payload{
					Time: 12345,
					Data: 101,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)

			var msg message.Msg
			gotN, err := msg.ReadFrom(tt.in)
			if tt.wantErr {
				is.True(err != nil)
			} else {
				is.NoErr(err)
			}

			is.Equal(gotN, msg.Len())
			is.Equal(msg, tt.wantMsg)
		})
	}
}
