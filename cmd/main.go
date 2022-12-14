package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"proto/pkg/chat"
	"proto/pkg/chat/broker"
	"proto/pkg/database"
	"proto/pkg/echo"
	"proto/pkg/insecsock"
	"proto/pkg/jobcentre"
	"proto/pkg/lrcp"
	"proto/pkg/price"
	"proto/pkg/prime"
	"proto/pkg/proxy"
	"proto/pkg/speed"
	"proto/pkg/tcpserver"
	"proto/pkg/udpserver"
)

const (
	problem = 9
	port    = 8080
	udpPort = 5000
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	switch problem {
	case 0:
		// Task 00 - Smoke test - https://protohackers.com/problem/0
		ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
		if err := tcpserver.Listen(ctx, port, echo.Handle); err != nil {
			log.Println("Error: [Listen]:", err.Error())
		}
		cancel()

	case 1:
		// Task 01 - Prime Time - https://protohackers.com/problem/1
		ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
		if err := tcpserver.Listen(ctx, port, prime.Handle); err != nil {
			log.Println("Error: [Listen]:", err.Error())
		}
		cancel()

	case 2:
		// Task 02 - Means to an End - https://protohackers.com/problem/2
		ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
		err := tcpserver.Listen(ctx, port, func(ctx context.Context, conn net.Conn) {
			(&price.Handler{}).Handle(ctx, conn)
		})
		if err != nil {
			log.Println("Error: [Listen]:", err.Error())
		}
		cancel()

	case 3:
		// Task 03 - Budget Chat - https://protohackers.com/problem/3
		broker := broker.New()
		err := tcpserver.Listen(ctx, port, func(ctx context.Context, conn net.Conn) {
			chat.NewSession(broker).Handle(ctx, conn)
		})
		if err != nil {
			log.Println("Error: [Listen]:", err.Error())
		}

	case 4:
		// Task 04 - Unusual Database Program - https://protohackers.com/problem/4
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

		err := udpserver.Listen(ctx, listen, func(ctx context.Context, w io.Writer, buf []byte) {
			db.Handle(ctx, w, buf)
		})
		if err != nil {
			log.Println("Error: [Listen]:", err.Error())
		}

	case 5:
		// Task 05 - Mob in the Middle - https://protohackers.com/problem/5
		err := tcpserver.Listen(ctx, port, func(ctx context.Context, conn net.Conn) {
			proxy.New(conn).Handle(ctx, conn)
		})
		if err != nil {
			log.Println("Error: [Listen]:", err.Error())
		}

	case 6:
		// Task 06 - Speed Daemon - https://protohackers.com/problem/6
		sd := speed.New(ctx)
		err := tcpserver.Listen(ctx, port, func(ctx context.Context, conn net.Conn) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			sd.Handle(ctx, conn, conn.RemoteAddr())
		})
		if err != nil {
			log.Println("Error: [Listen]:", err.Error())
		}

	case 7:
		// Task 07 - Line reversal - https://protohackers.com/problem/7
		lrcp := lrcp.New(ctx)

		listen := os.Getenv("ADDRESS")
		if listen == "" {
			listen = fmt.Sprintf(":%d", udpPort)
		}

		err := udpserver.Listen(ctx, listen, func(ctx context.Context, w io.Writer, buf []byte) {
			lrcp.Handle(ctx, w, buf)
		})
		if err != nil {
			log.Println("Error: [Listen]:", err.Error())
		}

	case 8:
		// Task 08 - Insecure Sockets Layer - https://protohackers.com/problem/8

		err := tcpserver.Listen(ctx, port, func(ctx context.Context, conn net.Conn) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			sockLayer, err := insecsock.NewLayer(ctx, conn)
			if err != nil {
				log.Printf("failed to create a secure layer: %s", err.Error())
				return
			}

			sockLayer.Handle(ctx)
		})
		if err != nil {
			log.Println("Error: [Listen]:", err.Error())
		}

	case 9:
		// Task 09 - Job Centre - https://protohackers.com/problem/9
		err := tcpserver.Listen(ctx, port, func(ctx context.Context, conn net.Conn) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			jobcentre.NewSession(ctx, conn).Handle(ctx)
		})
		if err != nil {
			log.Println("Error: [Listen]:", err.Error())
		}
	}
}
