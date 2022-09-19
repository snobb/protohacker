package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	port := os.Getenv("PORT")

	lst, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Printf("Failed to create a listener: %s", err.Error())
		os.Exit(1)
	}
	defer lst.Close()

	log.Printf("Listening on %s", port)

	for {
		conn, err := lst.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %s", err.Error())
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 2048)

	size, err := io.CopyBuffer(conn, conn, buf)
	if err != nil {
		log.Printf("Failed to echo: %s", err.Error())
	}

	log.Printf("Echoed %d bytes", size)
}
