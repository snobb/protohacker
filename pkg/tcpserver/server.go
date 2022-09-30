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
	lst, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Println("Failed to create a listener:", err.Error())
		return err
	}
	defer lst.Close()

	for {
		select {
		case <-ctx.Done():
			break
		default:
		}

		conn, err := lst.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err.Error())
		}

		log.Printf("Accepted connection from %s", conn.RemoteAddr().String())

		go func() {
			defer conn.Close()
			handler(ctx, conn)
		}()
	}
}
