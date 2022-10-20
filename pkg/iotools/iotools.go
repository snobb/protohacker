package iotools

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
)

// GetLine is a standard line splitter based on bufio.Scanner
func GetLine(ctx context.Context, r io.Reader) <-chan []byte {
	return getLines(ctx, bufio.NewScanner(r))
}

// GetLine is a modified line splitter based on bufio.Scanner
// By default bufio.Scanner would emit the accumulated buffer even if the next delimiter isn't
// found, which can be undesired behaviour. This version sticks to delimiter strictly and throws
// away the collected buffer if the next delimiter was NOT found.
func GetLineStrict(ctx context.Context, r io.Reader) <-chan []byte {
	scanner := bufio.NewScanner(r)

	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			// We have a full newline-terminated line.
			return i + 1, dropCR(data[0:i]), nil
		}

		// Request more data.
		return 0, nil, nil
	})

	return getLines(ctx, scanner)
}

func getLines(ctx context.Context, scanner *bufio.Scanner) <-chan []byte {
	ch := make(chan []byte)

	go func() {
		defer close(ch)

		for idx := 0; scanner.Scan(); idx++ {
			select {
			case <-ctx.Done():
				log.Println("getLines: got canceled")
				return

			default:
			}

			bytes := scanner.Bytes()
			if len(bytes) == 0 {
				break
			}

			ch <- bytes
		}
	}()

	return ch
}

// dropCR drops a terminal \r from the data (copy/paste from bufio stdlib)
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
