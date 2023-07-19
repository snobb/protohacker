package udpserver

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
)

const bufsz = 16384

type PacketConn struct {
	pconn net.PacketConn
	addr  net.Addr
}

// HandlerFunc is a tcp listener callback.
type HandlerFunc func(ctx context.Context, w io.Writer, buf []byte)

// Write writes buf buffer into the UDP socket and returns number of bytes written or an error
func (u *PacketConn) Write(buf []byte) (n int, err error) {
	return u.pconn.WriteTo(buf, u.addr)
}

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

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		buf := make([]byte, bufsz)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return err
			}

			log.Println("Failed to read buffer:", err.Error())
			continue
		}

		log.Printf("Accepted connection from %s - payload: %v", addr.String(), string(buf[:n]))

		conn := &PacketConn{
			pconn: pc,
			addr:  addr,
		}

		go handler(ctx, conn, buf[:n])
	}
}
