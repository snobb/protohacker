package iotools

import (
	"context"
	"io"
	"time"
)

// RateLimit implements io.ReadWriter and introduces rate limiting on write.
type RateLimit struct {
	Delay time.Duration
	rw    io.ReadWriter
}

// NewRateLimit creates a new RateLimit instance
func NewRateLimit(ctx context.Context, rw io.ReadWriter) *RateLimit {
	return &RateLimit{
		rw:    rw,
		Delay: 100 * time.Millisecond,
	}
}

// Read implements io.Reader for RateLimit - pass down
func (r *RateLimit) Read(p []byte) (n int, err error) {
	return r.rw.Read(p)
}

// Write writes with the rate limiting
func (r *RateLimit) Write(p []byte) (n int, err error) {
	var data = [1]byte{0x00}

	for _, b := range p {
		data[0] = b
		_, err = r.rw.Write(data[:])
		if err != nil {
			return
		}
		time.Sleep(r.Delay)
		n++
	}

	return
}
