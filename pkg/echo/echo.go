package echo

import (
	"context"
	"io"
	"log"
	"net"
)

func Handle(ctx context.Context, conn net.Conn) {
	buf := make([]byte, 2048)

	size, err := io.CopyBuffer(conn, conn, buf)
	if err != nil {
		log.Println("Failed to echo:", err.Error())
		return
	}

	log.Printf("Echoed %d bytes\n", size)
}
