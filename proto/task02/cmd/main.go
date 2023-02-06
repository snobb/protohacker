package main

import (
	"context"
	"log"
	"net"

	"proto/common/pkg/tcpserver"
	"proto/task02/pkg/price"
)

const tcpPort = 8080

// Task02 - Means to an End - https://protohackers.com/problem/2
func Task02(ctx context.Context) {
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := tcpserver.Listen(ctx, tcpPort, func(ctx context.Context, conn net.Conn) {
		(&price.Handler{}).Handle(ctx, conn)
	})
	if err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}
