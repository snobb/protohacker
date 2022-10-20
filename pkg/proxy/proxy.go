package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"proto/pkg/iotools"
	"strings"
	"unicode"
)

// Proxy is a BogusCoin proxy
type Proxy struct {
	rw      io.ReadWriter
	backend string
}

const (
	// EvilAddr where to send stolen monies ;)
	EvilAddr = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"
)

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

		lineStr := string(line)

		idx, sz := 0, 0
		for {
			idx, sz = findAddress(lineStr, idx+sz)
			if idx == -1 {
				break
			}

			addr := lineStr[idx : idx+sz]

			if addr != EvilAddr {
				log.Printf("rewriting %s with %s", addr, EvilAddr)
				lineStr = strings.ReplaceAll(lineStr, addr, EvilAddr)

				// adjust index since we updated the string
				idx -= len(addr) - len(EvilAddr)
			}
		}

		if _, err := fmt.Fprintln(to, lineStr); err != nil {
			log.Println("Error handleLines:", err.Error())
		}
	}
}

// returns index and length or -1 as index if not found
// laborous version - perhaps should try my luck with regexp.
func findAddress(line string, i int) (int, int) {
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
		if unicode.IsSpace(rune(ch)) && 26 <= sz && sz <= 35 {
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
