package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"proto/common/pkg/udpserver"
	"proto/task07/pkg/lrcp"
)

const udpPort = 5000

// Task07 - Line reversal - https://protohackers.com/problem/7
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lrcp := lrcp.New(ctx)

	listen := os.Getenv("ADDRESS")
	if listen == "" {
		listen = fmt.Sprintf(":%d", udpPort)
	}

	err := udpserver.Listen(ctx, listen, func(ctx context.Context, w io.Writer, buf []byte) {
		lrcp.Handle(ctx, w, buf)
	})
	if err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}
