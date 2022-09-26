package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"proto/pkg/echo"
	"proto/pkg/price"
	"proto/pkg/prime"
)

var version string

const problem = 2 // price

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	lst, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Println("Failed to create a listener:", err.Error())
		os.Exit(1)
	}
	defer lst.Close()

	log.Printf("[version: %s] Listening on: %s\n", version, port)

	for {
		conn, err := lst.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err.Error())
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer func() {
				cancel()
				conn.Close()
			}()

			switch problem {
			case 0:
				// Task 00 - https://protohackers.com/problem/0
				echo.Handle(ctx, conn)

			case 1:
				// Task 01 - https://protohackers.com/problem/1
				prime.Handle(ctx, conn)

			case 2:
				// Task 02 - https://protohackers.com/problem/2
				(&price.Handler{IO: conn}).Handle(ctx)
			}
		}()
	}
}
