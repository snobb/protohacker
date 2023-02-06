package main

import (
	"context"
	"log"
	"net"

	"proto/common/pkg/tcpserver"
	"proto/task09/pkg/jobcentre"
)

const tcpPort = 8080

// Task09 - Job Centre - https://protohackers.com/problem/9
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := tcpserver.Listen(ctx, tcpPort, func(ctx context.Context, conn net.Conn) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		jobcentre.NewSession(ctx, conn).Handle(ctx)
	})
	if err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}
