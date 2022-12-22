package insecsock

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"proto/pkg/insecsock/app"
	"proto/pkg/insecsock/ciphers"
)

// Layer (insecure)
type Layer struct {
	rw io.ReadWriter

	posIn  int
	posOut int

	// ciphers is a list of ciphers to encode/decode the incoming stream.
	ciphers []ciphers.Cipher
}

// NewLayer creates a new Insecure Sockets layer stream
func NewLayer(ctx context.Context, rw io.ReadWriter) (*Layer, error) {
	ciphers, err := readCiphers(ctx, rw)
	if err != nil {
		return nil, fmt.Errorf("Invalid cipher spec: %w", err)
	}

	if err := verifyCiphers(ciphers); err != nil {
		return nil, err
	}

	return &Layer{
		rw:      rw,
		ciphers: ciphers,
	}, nil
}

// Read is a io.Read implementation
func (l *Layer) Read(p []byte) (n int, err error) {
	n, err = l.rw.Read(p)
	if err != nil {
		return n, err
	}

	l.Decode(p[:n])
	return n, err
}

// Write is a io.Writer implementation
func (l *Layer) Write(p []byte) (n int, err error) {
	l.Encode(p)
	return l.rw.Write(p)
}

// Encode will encode all bytes of the provided buffer in place
func (l *Layer) Encode(p []byte) {
	for i := range p {
		for _, cph := range l.ciphers {
			p[i] = cph.Do(p[i], byte(l.posOut%256))
		}

		l.posOut++
	}
}

// Decode will decode all bytes of the provided buffer in place
func (l *Layer) Decode(p []byte) {
	for i := range p {
		for j := len(l.ciphers) - 1; j >= 0; j-- {
			p[i] = l.ciphers[j].Undo(p[i], byte(l.posIn%256))
		}

		l.posIn++
	}
}

// Handle handles the connection with the InSecureLayer
func (l *Layer) Handle(ctx context.Context) {
	scanner := bufio.NewScanner(l)

	for scanner.Scan() {
		line := scanner.Text()

		resp := app.HandleLine(line)

		log.Printf("line: %s -> %s", line, resp)

		if _, err := l.Write([]byte(fmt.Sprintf("%s\n", resp))); err != nil {
			log.Printf("couldnot encode/write the response: %s", err.Error())
		}
	}
}

func readCiphers(ctx context.Context, r io.Reader) ([]ciphers.Cipher, error) {
	cc := []ciphers.Cipher{}
	var buf [1]byte

	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("cancelled")
		default:
		}

		_, err := r.Read(buf[:])
		if err != nil {
			return nil, fmt.Errorf("Could not read byte: %w", err)
		}

		switch buf[0] {
		case 0x00:
			log.Printf("ciphers: %+#v", cc)
			return cc, nil

		case 0x01: // reversebits
			cc = append(cc, &ciphers.ReverseBits{})

		case 0x02: // xor(N)
			_, err := r.Read(buf[:])
			if err != nil {
				return nil, fmt.Errorf("Could not read byte: %w", err)
			}

			cc = append(cc, &ciphers.Xor{N: buf[0]})

		case 0x03: // xorpos
			cc = append(cc, &ciphers.XorPos{})

		case 0x04: // add(N)
			_, err := r.Read(buf[:])
			if err != nil {
				return nil, fmt.Errorf("Could not read byte: %w", err)
			}

			cc = append(cc, &ciphers.Add{N: buf[0]})

		case 0x05: // addpos
			cc = append(cc, &ciphers.AddPos{})

		default:
			return nil, fmt.Errorf("invalid cipher id: %d", buf[0])
		}
	}
}

func verifyCiphers(ciphers []ciphers.Cipher) error {
	test := "foobar"

	pos := 0
	buf := []byte(test)
	for i := range buf {
		for _, cph := range ciphers {
			buf[i] = cph.Do(buf[i], byte(pos))
		}
		pos++
	}

	if string(buf) == test {
		return errors.New("no-op spec")
	}

	return nil
}
