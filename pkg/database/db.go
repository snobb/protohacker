package database

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
)

const (
	version = "wierd database v1.0"
)

// DB is a wierd database implementation.
type DB struct {
	store map[string]string
}

// New creates a new DB instance
func New() *DB {
	return &DB{
		store: make(map[string]string),
	}
}

// Handle handles an UDP message
func (d *DB) Handle(ctx context.Context, w io.Writer, buf []byte) {
	key, value, insert := parseMessage(string(buf))
	if insert {
		d.store[key] = value
		return
	}

	// retrieve
	if key == "version" {
		send(fmt.Sprintf("version=%s", version), w)
		return
	}

	value = d.store[key]
	send(fmt.Sprintf("%s=%s", key, value), w)
}

func send(msg string, w io.Writer) {
	if len(msg) > 1000 {
		return
	}

	log.Println("sending:", msg)

	if _, err := w.Write([]byte(msg)); err != nil {
		log.Println("Error sending a response:", err.Error())
	}
}

// returns key, value, isInsert bool
func parseMessage(buf string) (string, string, bool) {
	idx := strings.Index(buf, "=")
	if idx == -1 {
		return buf, "", false
	}

	return buf[:idx], buf[idx+1:], true
}
