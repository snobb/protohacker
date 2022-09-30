package main

import (
	"context"
	"log"
	"net"
	"time"

	"proto/pkg/echo"
	"proto/pkg/price"
	"proto/pkg/prime"
	"proto/pkg/tcpserver"
)

var version = "devel"

const (
	problem = 2
	port    = 8080
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("[version: %s] Listening on: %d\n", version, port)

	switch problem {
	case 0:
		// Task 00 - https://protohackers.com/problem/0
		ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
		if err := tcpserver.Listen(ctx, port, echo.Handle); err != nil {
			log.Println("Error: [Listen]:", err.Error())
		}
		cancel()

	case 1:
		// Task 01 - https://protohackers.com/problem/1
		ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
		if err := tcpserver.Listen(ctx, port, prime.Handle); err != nil {
			log.Println("Error: [Listen]:", err.Error())
		}
		cancel()

	case 2:
		// Task 02 - https://protohackers.com/problem/2
		ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
		err := tcpserver.Listen(ctx, port, func(ctx context.Context, conn net.Conn) {
			(&price.Handler{}).Handle(ctx, conn)
		})
		if err != nil {
			log.Println("Error: [Listen]:", err.Error())
		}
		cancel()

	case 3:
		// Task 03 - https://protohackers.com/problem/3
	}
}
