package main

import (
	"context"
	"log"
	"net"

	"proto/common/pkg/tcpserver"
	"proto/task03/pkg/chat"
	"proto/task03/pkg/chat/broker"
)

const tcpPort = 8080

// Task03 - Budget Chat - https://protohackers.com/problem/3
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	broker := broker.New()
	err := tcpserver.Listen(ctx, tcpPort, func(ctx context.Context, conn net.Conn) {
		chat.NewSession(broker).Handle(ctx, conn)
	})
	if err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}
