package main

import (
	"context"
	"log"
	"net"

	"proto/common/pkg/tcpserver"
	"proto/task06/pkg/speed"
)

const tcpPort = 8080

// Task06 - Speed Daemon - https://protohackers.com/problem/6
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sd := speed.New(ctx)
	err := tcpserver.Listen(ctx, tcpPort, func(ctx context.Context, conn net.Conn) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		sd.Handle(ctx, conn, conn.RemoteAddr())
	})
	if err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}
