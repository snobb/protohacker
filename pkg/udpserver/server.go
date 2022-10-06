package udpserver

import (
	"context"
	"errors"
	"log"
	"net"
)

// HandlerFunc is a tcp listener callback.
type HandlerFunc func(ctx context.Context, conn net.PacketConn, addr net.Addr, buf []byte)

// Listen listens for an UDP connection
func Listen(ctx context.Context, addr string, handler HandlerFunc) error {
	log.Printf("Listening on: %s\n", addr)

	pc, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Println("Failed to create a listener:", err.Error())
		return err
	}
	defer func() {
		_ = pc.Close()
	}()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
		}

		buf := make([]byte, 2048)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return err
			}

			log.Println("Failed to read buffer:", err.Error())
			continue
		}

		if n > 1000 {
			log.Println("Message is too long:", err.Error())
			continue
		}

		log.Printf("Accepted connection from %s - payload: %v", addr.String(), string(buf[:n]))

		go func() {
			handler(ctx, pc, addr, buf[:n])
		}()
	}

	return nil
}
