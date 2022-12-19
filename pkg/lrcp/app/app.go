package app

import (
	"bufio"
	"bytes"
)

// App is the application layer of the line processor.
type App struct {
	in  bytes.Buffer
	out bytes.Buffer
}

// Read is the io.Reader implemetnation for App
func (a *App) Read(p []byte) (n int, err error) {
	return a.out.Read(p)
}

// Write is the io.Writer implemetnation for App
func (a *App) Write(p []byte) (n int, err error) {
	n, err = a.in.Write(p)
	a.process()
	return
}

// process splits the contents of the "in" buffer by the new line and reverses
// the contents (keeping newlines in place). Incomplete line is returned back to "in" buffer and
// will be reassessed on the next Write.
func (a *App) process() {
	br := bufio.NewReader(&a.in)

	for {
		line, err := br.ReadBytes('\n')
		if err != nil {
			// incomplete line - return the remainder back to the in buffer
			a.in.Write(line)
			break
		}

		sz := len(line) - 2
		for i, j := 0, sz; i < (len(line)-1)/2; i, j = i+1, j-1 {
			line[i], line[j] = line[j], line[i]
		}

		a.out.Write(line)
	}
}
