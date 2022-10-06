package database

import (
	"context"
	"fmt"
	"log"
	"net"
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
func (d *DB) Handle(ctx context.Context, pc net.PacketConn, addr net.Addr, buf []byte) {
	key, value, insert := parseMessage(string(buf))
	if insert {
		d.store[key] = value
		return
	}

	// retrieve
	if key == "version" {
		send(fmt.Sprintf("version=%s", version), pc, addr)
		return
	}

	value = d.store[key]
	send(fmt.Sprintf("%s=%s", key, value), pc, addr)
}

func send(msg string, pc net.PacketConn, addr net.Addr) {
	if len(msg) > 1000 {
		return
	}

	log.Println("sending:", msg)

	if _, err := pc.WriteTo([]byte(msg), addr); err != nil {
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
