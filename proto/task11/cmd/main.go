package main

import (
	"context"
	"log"
	"net"

	"proto/common/pkg/tcpserver"
	"proto/task11/pkg/pestcontrol"
)

const tcpPort = 8080

// 11. Pest control - https://protohackers.com/problem/11

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := tcpserver.Listen(ctx, tcpPort, func(ctx context.Context, conn net.Conn) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		pestcontrol.New(conn, conn.RemoteAddr()).Handle(ctx)
	})
	if err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}
