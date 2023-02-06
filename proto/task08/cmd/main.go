package main

import (
	"context"
	"log"
	"net"

	"proto/common/pkg/tcpserver"
	"proto/task08/pkg/insecsock"
)

const tcpPort = 8080

// Task08 - Insecure Sockets Layer - https://protohackers.com/problem/8
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := tcpserver.Listen(ctx, tcpPort, func(ctx context.Context, conn net.Conn) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		sockLayer, err := insecsock.NewLayer(ctx, conn)
		if err != nil {
			log.Printf("failed to create an (in)secure layer: %s", err.Error())
			return
		}

		sockLayer.Handle(ctx)
	})
	if err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}
