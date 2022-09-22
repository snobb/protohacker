package echo

import (
	"context"
	"io"
	"log"
)

func Handle(ctx context.Context, ioh io.ReadWriteCloser) {
	defer ioh.Close()

	buf := make([]byte, 2048)

	size, err := io.CopyBuffer(ioh, ioh, buf)
	if err != nil {
		log.Println("Failed to echo:", err.Error())
	}

	log.Printf("Echoed %d bytes\n", size)
}
