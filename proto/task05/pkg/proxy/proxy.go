package proxy

import (
	"bytes"
	"context"
	"io"
	"log"
	"net"
	"os"

	"proto/common/pkg/iotools"
)

// Proxy is a BogusCoin proxy
type Proxy struct {
	rw      io.ReadWriter
	backend string
}

// EvilAddr where to send stolen monies ;)
var EvilAddr = []byte("7YWHMfk9JZe0LM0g1ZauHuiSxhI")

// New creates a new Proxy instance.
func New(rw io.ReadWriter) *Proxy {
	backend := os.Getenv("ADDRESS")
	if backend == "" {
		backend = "localhost:8100"
	}

	return &Proxy{rw: rw, backend: backend}
}

// Handle handles the proxy connection
func (p *Proxy) Handle(ctx context.Context, fe io.ReadWriter) {
	be, err := p.connect(ctx)
	if err != nil {
		return
	}
	defer func() {
		_ = be.Close()
	}()

	go handleLines(ctx, be, fe)
	handleLines(ctx, fe, be)
}

func (p *Proxy) connect(ctx context.Context) (net.Conn, error) {
	log.Println("Connecting to", p.backend)
	be, err := net.Dial("tcp", p.backend)
	if err != nil {
		log.Println("Error client dial:", err.Error())
		return nil, err
	}

	return be, nil
}

func handleLines(ctx context.Context, from, to io.ReadWriter) {
	for line := range iotools.GetLineStrict(ctx, from) {
		if len(line) == 0 {
			continue
		}

		idx, sz := 0, 0
		for {
			idx, sz = findAddress(line, idx+sz)
			if idx == -1 {
				break
			}

			addr := line[idx : idx+sz]

			if !bytes.Equal(addr, EvilAddr) {
				log.Printf("rewriting %s with %s", addr, EvilAddr)
				line = bytes.ReplaceAll(line, addr, EvilAddr)

				// adjust index since we updated the string
				idx -= len(addr) - len(EvilAddr)
			}
		}

		if _, err := to.Write(append(line, '\n')); err != nil {
			log.Println("Error handleLines:", err.Error())
		}
	}
}

// returns index and length or -1 as index if not found
// laborous version - perhaps should try my luck with regexp.
func findAddress(line []byte, i int) (int, int) {
	start := -1

	for ; i < len(line); i++ {
		ch := line[i]

		// if not within a token AND found a start of a token - mark it and advance
		if start < 0 && ch == '7' && (i == 0 || line[i-1] == ' ') {
			start = i
			continue
		}

		// if not within a token OR within a token AND on legit chars - advance
		if start < 0 ||
			('a' <= ch && ch <= 'z') ||
			('A' <= ch && ch <= 'Z') ||
			('0' <= ch && ch <= '9') {
			continue
		}

		// if we're here - we have been within a token AND found a non-legit char - handle buffer
		sz := i - start
		if ch == ' ' && 26 <= sz && sz <= 35 {
			return start, sz
		}

		// the collected token isn't valid - reset
		start = -1
	}

	// found the end of line and have a token buffer - deal with it
	if start > -1 {
		sz := i - start
		if 26 <= sz && sz <= 35 {
			return start, sz
		}
	}

	return -1, -1
}
