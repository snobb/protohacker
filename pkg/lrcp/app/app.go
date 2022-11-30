package app

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"log"
)

// App is the application layer of the line processor.
type App struct {
	in     chan []byte
	out    chan []byte
	closed bool
}

// New creates a new instance of App
// It splits the buffer contests by new line and returns a slice of the split lines, when each is
// reversed. The function assumes there is a trailing CR for every given line.
func New(ctx context.Context) *App {
	app := &App{
		in:  make(chan []byte, 100),
		out: make(chan []byte, 100),
	}

	go func() {
		defer app.Close()
		defer close(app.out)
		defer log.Print("Closing out channel")

		var buf bytes.Buffer
		br := bufio.NewReadWriter(bufio.NewReader(&buf), bufio.NewWriter(&buf))

		for chunk := range app.in {
			select {
			case <-ctx.Done():
				return
			default:
			}

			buf.Write(chunk)

			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				line, err := br.ReadBytes('\n')
				if err != nil {
					// return the remainder back to the buf
					buf.Write(line)
					break
				}

				sz := len(line) - 2
				for i, j := 0, sz; i < (len(line)-1)/2; i, j = i+1, j-1 {
					line[i], line[j] = line[j], line[i]
				}

				app.out <- line
			}
		}
	}()

	return app
}

// Write is the io.Writer implemetnation for App
func (a *App) Write(p []byte) (n int, err error) {
	if a.closed {
		return -1, errors.New("writing to a close app")
	}

	a.in <- p
	return len(p), nil
}

// Close implements io.Closer interface and closes the output channel.
func (a *App) Close() error {
	if a.in != nil {
		close(a.in)
		a.in = nil
		a.closed = true
	}

	// if a.out != nil {
	// 	close(a.out)
	// 	a.out = nil
	// }

	return nil
}

// OutCh return the output channel
func (a *App) OutCh() <-chan []byte {
	return a.out
}
