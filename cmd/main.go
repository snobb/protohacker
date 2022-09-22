package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"proto/pkg/prime"
)

var version string

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

			prime.Handle(ctx, conn)
		}()
	}
}
