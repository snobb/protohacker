package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"proto/pkg/chat"
	"proto/pkg/chat/broker"
	"proto/pkg/codestore"
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
	tcpPort = 8080
	udpPort = 5000
)

// Task00 - Smoke test - https://protohackers.com/problem/0
func Task00(ctx context.Context) {
	if err := tcpserver.Listen(ctx, tcpPort, echo.Handle); err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}

// Task01 - Prime Time - https://protohackers.com/problem/1
func Task01(ctx context.Context) {
	if err := tcpserver.Listen(ctx, tcpPort, prime.Handle); err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}

// Task02 - Means to an End - https://protohackers.com/problem/2
func Task02(ctx context.Context) {
	err := tcpserver.Listen(ctx, tcpPort, func(ctx context.Context, conn net.Conn) {
		(&price.Handler{}).Handle(ctx, conn)
	})
	if err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}

// Task03 - Budget Chat - https://protohackers.com/problem/3
func Task03(ctx context.Context) {
	broker := broker.New()
	err := tcpserver.Listen(ctx, tcpPort, func(ctx context.Context, conn net.Conn) {
		chat.NewSession(broker).Handle(ctx, conn)
	})
	if err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}

// Task04 - Unusual Database Program - https://protohackers.com/problem/4
// Add the following to the fly.toml
// [env]
//
//	ADDRESS = "fly-global-services:5000"
//
// [[services]]
//
//	internal_port = 5000
//	protocol = "udp"
//	[[services.ports]]
//	   port = 5000
func Task04(ctx context.Context) {
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
}

// Task05 - Mob in the Middle - https://protohackers.com/problem/5
func Task05(ctx context.Context) {
	err := tcpserver.Listen(ctx, tcpPort, func(ctx context.Context, conn net.Conn) {
		proxy.New(conn).Handle(ctx, conn)
	})
	if err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}

// Task06 - Speed Daemon - https://protohackers.com/problem/6
func Task06(ctx context.Context) {
	sd := speed.New(ctx)
	err := tcpserver.Listen(ctx, tcpPort, func(ctx context.Context, conn net.Conn) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		sd.Handle(ctx, conn, conn.RemoteAddr())
	})
	if err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}

// Task07 - Line reversal - https://protohackers.com/problem/7
func Task07(ctx context.Context) {
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
}

// Task08 - Insecure Sockets Layer - https://protohackers.com/problem/8
func Task08(ctx context.Context) {
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

// Task09 - Job Centre - https://protohackers.com/problem/9
func Task09(ctx context.Context) {
	err := tcpserver.Listen(ctx, tcpPort, func(ctx context.Context, conn net.Conn) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		jobcentre.NewSession(ctx, conn).Handle(ctx)
	})
	if err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}

// Task10 - Voracious Code Storage - https://protohackers.com/problem/10
func Task10(ctx context.Context) {
	err := tcpserver.Listen(ctx, tcpPort, func(ctx context.Context, conn net.Conn) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		codestore.New(conn, conn.RemoteAddr()).Handle(ctx)
	})
	if err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	Task10(ctx)
}
