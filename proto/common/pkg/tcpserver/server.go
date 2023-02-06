package tcpserver

import (
	"context"
	"fmt"
	"log"
	"net"
)

// HandlerFunc is a tcp listener callback.
type HandlerFunc func(ctx context.Context, conn net.Conn)

func Listen(ctx context.Context, port int, handler HandlerFunc) error {
	log.Printf("Listening on: %d\n", port)

	lst, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Println("Failed to create a listener:", err.Error())
		return err
	}
	defer func() {
		_ = lst.Close()
	}()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
		}

		conn, err := lst.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err.Error())
		}

		log.Printf("Accepted connection from %s", conn.RemoteAddr().String())

		go func() {
			defer func() {
				_ = conn.Close()
			}()
			handler(ctx, conn)
		}()
	}

	return nil
}
