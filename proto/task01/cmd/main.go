package main

import (
	"context"
	"log"

	"proto/common/pkg/tcpserver"
	"proto/task01/pkg/prime"
)

const tcpPort = 8080

// Task01 - Prime Time - https://protohackers.com/problem/1
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := tcpserver.Listen(ctx, tcpPort, prime.Handle); err != nil {
		log.Println("Error: [Listen]:", err.Error())
	}
}
