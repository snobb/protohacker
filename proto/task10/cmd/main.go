package main

import (
	"context"
	"log"
	"net"

	"proto/common/pkg/tcpserver"
	"proto/task10/pkg/codestore"
)

const tcpPort = 8080

// Task10 - Voracious Code Storage - https://protohackers.com/problem/10
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := tcpserver.Listen(ctx, tcpPort, func(ctx context.Context, conn net.Conn) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		codestore.New(conn, conn.RemoteAddr()).Handle(ctx)
	})
	if err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}
