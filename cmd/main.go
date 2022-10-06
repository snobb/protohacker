package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"proto/pkg/chat"
	"proto/pkg/chat/broker"
	"proto/pkg/database"
	"proto/pkg/echo"
	"proto/pkg/price"
	"proto/pkg/prime"
	"proto/pkg/tcpserver"
	"proto/pkg/udpserver"
)

const (
	problem = 4
	port    = 8080
	udpPort = 5000
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
		broker := broker.New()
		err := tcpserver.Listen(ctx, port, func(ctx context.Context, conn net.Conn) {
			chat.NewSession(broker).Handle(ctx, conn)
		})
		if err != nil {
			log.Println("Error: [Listen]:", err.Error())
		}

	case 4:
		// Task 04 - https://protohackers.com/problem/4
		//
		// Add the following to the fly.toml
		// [env]
		//   ADDRESS = "fly-global-services:5000"
		// [[services]]
		//   internal_port = 5000
		//   protocol = "udp"
		//   [[services.ports]]
		//      port = 5000
		db := database.New()

		listen := os.Getenv("ADDRESS")
		if listen == "" {
			listen = fmt.Sprintf(":%d", udpPort)
		}

		err := udpserver.Listen(ctx, listen, func(ctx context.Context, conn net.PacketConn, addr net.Addr, buf []byte) {
			db.Handle(ctx, conn, addr, buf)
		})
		if err != nil {
			log.Println("Error: [Listen]:", err.Error())
		}
	}
}
