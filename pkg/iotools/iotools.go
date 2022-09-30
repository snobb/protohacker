package iotools

import (
	"bufio"
	"context"
	"io"
	"log"
)

func GetLine(ctx context.Context, r io.Reader) <-chan []byte {
	ch := make(chan []byte)

	scanner := bufio.NewScanner(r)
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
