package insecsock

import (
	"bytes"
	"context"
	"testing"

	"github.com/matryer/is"
)

func TestLayer_Encode(t *testing.T) {
	tests := []struct {
		name      string
		nCiphers  int
		cipherStr []byte
		data      []byte
		want      []byte
	}{
		{
			name:      "should encode data correctly with xor",
			nCiphers:  1,
			cipherStr: []byte{0x02, 0x01, 0x00},
			data:      []byte("hello"),
			want:      []byte{0x69, 0x64, 0x6d, 0x6d, 0x6e},
		},
		{
			name:      "should encode data correctly with xor and reversebit",
			nCiphers:  2,
			cipherStr: []byte{0x02, 0x01, 0x01, 0x00},
			data:      []byte("hello"),
			want:      []byte{0x96, 0x26, 0xb6, 0xb6, 0x76},
		},
		{
			name:      "should encode data correctly with addpos",
			nCiphers:  1,
			cipherStr: []byte{0x05, 0x00},
			data:      []byte("hello"),
			want:      []byte{0x68, 0x66, 0x6e, 0x6f, 0x73},
		},
		{
			name:      "should encode data correctly with 2 addpos",
			nCiphers:  2,
			cipherStr: []byte{0x05, 0x05, 0x00},
			data:      []byte("hello"),
			want:      []byte{0x68, 0x67, 0x70, 0x72, 0x77},
		},
		{
			name:      "should work with pos based ciphers as well",
			nCiphers:  3,
			cipherStr: []byte{0x02, 0x7b, 0x05, 0x01, 0x00},
			data:      []byte("4x dog,5x car\n"),
			want: []byte{
				0xf2, 0x20, 0xba, 0x44, 0x18,
				0x84, 0xba, 0xaa, 0xd0, 0x26,
				0x44, 0xa4, 0xa8, 0x7e,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			var rw bytes.Buffer

			rw.Write(tt.cipherStr)

			lr, err := NewLayer(context.Background(), &rw)
			is.NoErr(err)

			is.Equal(len(lr.ciphers), tt.nCiphers)
			t.Logf("%#v", lr.ciphers)

			n, err := lr.Write(tt.data)
			is.NoErr(err)
			is.Equal(n, len(tt.want))

			is.Equal(rw.Bytes(), tt.want)
		})
	}
}

func TestLayer_Decode(t *testing.T) {
	tests := []struct {
		name      string
		nCiphers  int
		cipherStr []byte
		want      []byte
		data      []byte
	}{
		{
			name:      "should encode data correctly with xor",
			nCiphers:  1,
			cipherStr: []byte{0x02, 0x01, 0x00},
			data:      []byte{0x69, 0x64, 0x6d, 0x6d, 0x6e},
			want:      []byte("hello"),
		},
		{
			name:      "should encode data correctly with xor and reversebit",
			nCiphers:  2,
			cipherStr: []byte{0x02, 0x01, 0x01, 0x00},
			data:      []byte{0x96, 0x26, 0xb6, 0xb6, 0x76},
			want:      []byte("hello"),
		},
		{
			name:      "should encode data correctly with addpos",
			nCiphers:  1,
			cipherStr: []byte{0x05, 0x00},
			data:      []byte{0x68, 0x66, 0x6e, 0x6f, 0x73},
			want:      []byte("hello"),
		},
		{
			name:      "should encode data correctly with 2 addpos",
			nCiphers:  2,
			cipherStr: []byte{0x05, 0x05, 0x00},
			data:      []byte{0x68, 0x67, 0x70, 0x72, 0x77},
			want:      []byte("hello"),
		},
		{
			name:      "should work with pos based ciphers as well",
			nCiphers:  3,
			cipherStr: []byte{0x02, 0x7b, 0x05, 0x01, 0x00},
			data: []byte{
				0xf2, 0x20, 0xba, 0x44, 0x18,
				0x84, 0xba, 0xaa, 0xd0, 0x26,
				0x44, 0xa4, 0xa8, 0x7e,
			},
			want: []byte("4x dog,5x car\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			var rw bytes.Buffer

			rw.Write(tt.cipherStr)
			rw.Write(tt.data)

			lr, err := NewLayer(context.Background(), &rw)
			is.NoErr(err)

			is.Equal(len(lr.ciphers), tt.nCiphers)
			t.Logf("%#v", lr.ciphers)

			buf := make([]byte, 100)
			n, err := lr.Read(buf)
			is.NoErr(err)
			is.Equal(n, len(tt.data))

			is.Equal(buf[:n], tt.want)
		})
	}
}
