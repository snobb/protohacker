package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"proto/common/pkg/udpserver"
	"proto/task04/pkg/database"
)

const udpPort = 5000

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

// Task04 - Unusual Database Program - https://protohackers.com/problem/4
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
