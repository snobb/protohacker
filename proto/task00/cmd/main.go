package main

import (
	"context"
	"log"

	"proto/common/pkg/tcpserver"
	"proto/task00/pkg/echo"
)

const tcpPort = 8080

// Task00 - Smoke test - https://protohackers.com/problem/0
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := tcpserver.Listen(ctx, tcpPort, echo.Handle); err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}
