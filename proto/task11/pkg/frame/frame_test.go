package frame_test

import (
	"bytes"
	"testing"

	"github.com/matryer/is"

	"proto/task11/pkg/frame"
)

func TestMsg_ReadFrom(t *testing.T) {
	is := is.New(t)

	tests := map[string]struct {
		payload  []byte
		wantKind uint8
		wantN    int64
		wantErr  bool
	}{
		"Should read frame correctly": {
			payload: []byte{
				0x50,                   // hello
				0x00, 0x00, 0x00, 0x19, // len: 25
				0x00, 0x00, 0x00, 0x0b, // proto len: 11
				0x70, 0x65, 0x73, 0x74, // "pest
				0x63, 0x6f, 0x6e, 0x74, // cont
				0x72, 0x6f, 0x6c, // rol"
				0x00, 0x00, 0x00, 0x01, // version: 1
				0xce, // chksum: 0xce
			},
			wantKind: frame.KindHello,
			wantN:    25,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buf := bytes.NewBuffer(tt.payload)

			msg := &frame.Frame{}
			gotN, err := msg.ReadFrom(buf)
			if tt.wantErr && err != nil {
				return
			}

			is.NoErr(err)

			is.Equal(msg.Kind, tt.wantKind)
			is.Equal(msg.Payload, tt.payload[5:len(tt.payload)-1])
			is.Equal(gotN, tt.wantN)
		})
	}
}

func TestMsg_WriteTo(t *testing.T) {
	is := is.New(t)

	tests := map[string]struct {
		msg         *frame.Frame
		wantPayload []byte
		wantN       int64
		wantErr     bool
	}{
		"Should read frame correctly": {
			msg: frame.New(frame.KindHello, []byte{
				0x00, 0x00, 0x00, 0x0b, // proto len: 11
				0x70, 0x65, 0x73, 0x74, // "pest
				0x63, 0x6f, 0x6e, 0x74, // cont
				0x72, 0x6f, 0x6c, // rol"
				0x00, 0x00, 0x00, 0x01, // version: 1
			}),
			wantPayload: []byte{
				0x50,                   // hello
				0x00, 0x00, 0x00, 0x19, // len: 25
				0x00, 0x00, 0x00, 0x0b, // proto len: 11
				0x70, 0x65, 0x73, 0x74, // "pest
				0x63, 0x6f, 0x6e, 0x74, // cont
				0x72, 0x6f, 0x6c, // rol"
				0x00, 0x00, 0x00, 0x01, // version: 1
				0xce, // chksum: 0xce
			},
			wantN: 25,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer

			gotN, err := tt.msg.WriteTo(&buf)
			if tt.wantErr && err != nil {
				return
			}

			is.NoErr(err)
			is.Equal(gotN, int64(buf.Len()))
			is.Equal(buf.Bytes(), tt.wantPayload)
		})
	}
}
