package main

import (
	"context"
	"log"
	"net"

	"proto/common/pkg/tcpserver"
	"proto/task05/pkg/proxy"
)

const tcpPort = 8080

// Task05 - Mob in the Middle - https://protohackers.com/problem/5
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := tcpserver.Listen(ctx, tcpPort, func(ctx context.Context, conn net.Conn) {
		proxy.New(conn).Handle(ctx, conn)
	})
	if err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}
