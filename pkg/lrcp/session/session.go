package session

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"proto/pkg/lrcp/app"
)

const (
	retransmitInterval = 3 * time.Second
	sessionTimeout     = 60 * time.Second
	pollInterval       = 50 * time.Millisecond
)

// Payload stores a storable payload.
type Payload struct {
	Pos    int
	Data   []byte
	SentAt time.Time
}

// Session is a session managing unit
type Session struct {
	ID     int
	w      io.Writer
	closed bool

	// RECEIVE
	rcvBytes int       // consequtive data we acked so far
	rcvLast  time.Time // time of the last received ack

	// SEND
	sendAck    int // sent data acked
	sendRtxCh  chan *Payload
	sendBytes  bytes.Buffer
	sendRtxPos int // pos of retransmitted chunk

	// application layer
	app *app.App
}

// New creates a new Session instance
func New(ctx context.Context, w io.Writer, id int) *Session {
	s := &Session{
		w:         w,
		ID:        id,
		rcvLast:   time.Now(),
		sendRtxCh: make(chan *Payload, 10),

		app: app.New(ctx),
	}

	ctx, cancel := context.WithCancel(ctx)

	// send data back from a single location
	go func() {
		// defer log.Printf("closing data sender go routine for %d", id)
		defer close(s.sendRtxCh)
		defer cancel()

		plCh := s.Payloads(ctx, s.sendRtxCh)

		for {
			select {
			case <-ctx.Done():
				return

			case pl, ok := <-plCh:
				if !ok {
					return
				}

				// send the payload and make sure it's acked in full (retransmit if necessary)
				s.sendRetransmit(ctx, pl)
			}
		}
	}()

	return s
}

// HandleConnect handles 'connect' message
func (s *Session) HandleConnect(ctx context.Context) error {
	log.Printf("CONNECT: sid:%d", s.ID)
	s.SendAck(ctx, 0)

	return nil
}

// HandleClose handles 'close' message
func (s *Session) HandleClose(ctx context.Context) error {
	s.Close()
	s.SendClose(ctx)

	return nil
}

// HandleAck handles 'ack' message
func (s *Session) HandleAck(ctx context.Context, length int) error {
	log.Printf("ACK: sid:%d length:%d", s.ID, length)
	s.notify()

	if length < s.sendAck {
		return nil

	} else if length > s.sendBytes.Len() {
		s.Close()
		s.SendClose(ctx)

	} else {
		// if not all data received - send the remainder
		// or if done
		s.sendAck = length
	}

	return nil
}

// HandleData handles 'data' message
func (s *Session) HandleData(ctx context.Context, pos int, data []byte) error {
	log.Printf("DATA: sid:%d pos:%d data:%s", s.ID, pos, string(data))
	s.notify()

	buf := Unescape(data)

	if pos > s.rcvBytes {
		// do not have all the data - send duplicate ack for data we have.
		s.SendAck(ctx, s.rcvBytes)
		s.SendAck(ctx, s.rcvBytes)

	} else if pos < s.rcvBytes {
		s.SendAck(ctx, s.rcvBytes)
		// resent chunk
		if pos != s.sendRtxPos && pos < s.sendBytes.Len() {
			s.sendRtxCh <- &Payload{
				Pos:  pos,
				Data: s.sendBytes.Bytes()[pos:],
			}
		}

	} else {
		// have all data up to pos + the current buffer
		s.rcvBytes += len(data)
		s.SendAck(ctx, s.rcvBytes)
		_, _ = s.app.Write(buf) // pass data to the application layer.
	}

	return nil
}

// SendAck sends an 'ack' message on wire.
func (s *Session) SendAck(ctx context.Context, length int) {
	// log.Printf("/ack/%d/%d/", s.ID, length)
	if _, err := fmt.Fprintf(s.w, "/ack/%d/%d/", s.ID, length); err != nil {
		log.Printf("Could not send an ack: %s", err.Error())
	}
}

// SendClose sends a 'close' message on wire.
func (s *Session) SendClose(ctx context.Context) {
	defer log.Printf("CLOSE: /close/%d/", s.ID)
	if _, err := fmt.Fprintf(s.w, "/close/%d/", s.ID); err != nil {
		log.Printf("Could not send a close: %s", err.Error())
	}
}

// Payloads listens to the application layer responses and sends the output into a designated
// channel + the possition of that payload.
func (s *Session) Payloads(ctx context.Context, in <-chan *Payload) <-chan *Payload {
	ch := make(chan *Payload, 1000)

	go func() {
		// defer log.Printf("closing Payloads go routine for %d", s.ID)
		defer close(ch)

		var pos int
		for {
			select {
			case <-ctx.Done():
				return

			case line, ok := <-s.app.OutCh():
				if !ok {
					log.Printf("[%d] close from app channel", s.ID)
					return
				}

				s.sendBytes.Write(line) // keeping all sent bytes in a single buf chunk.
				ch <- &Payload{
					Pos:  pos,
					Data: line,
				}
				pos += len(line) // advance pos

			case payload, ok := <-in:
				if !ok {
					log.Printf("[%d] close from out of band channel", s.ID)
					return
				}

				ch <- payload
			}
		}
	}()

	return ch
}

// Close closes the session
func (s *Session) Close() {
	s.closed = true

	// set last ack time to the past to sweep the connection.
	s.rcvLast = time.Now().Add(-1 * time.Hour)

	if err := s.app.Close(); err != nil {
		log.Printf("Could not close the app: %s", err.Error())
	}
}

// Expired returns true if session is expired
func (s *Session) Expired() bool {
	if s.rcvLast.IsZero() {
		return false
	}

	return time.Since(s.rcvLast) > sessionTimeout
}

// Closed returns true if session is closed or false otherwise.
func (s *Session) Closed() bool {
	return s.closed
}

func (s *Session) notify() {
	s.rcvLast = time.Now()
}

func (s *Session) sendRetransmit(ctx context.Context, pl *Payload) {
	// defer log.Printf("closing retransmit go routine for %d", s.ID)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	log.Printf("/data/%d/%d/%s/", s.ID, pl.Pos, string(Escape(pl.Data)))
	fmt.Fprintf(s.w, "/data/%d/%d/%s/", s.ID, pl.Pos, string(Escape(pl.Data)))
	s.sendRtxPos = pl.Pos
	sendTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if s.sendAck >= pl.Pos+len(pl.Data) {
				s.sendRtxPos = -1
				return
			}

			if time.Since(sendTime) >= retransmitInterval {
				if pl.Pos < s.sendAck {
					pl.Pos = s.sendAck
					pl.Data = s.sendBytes.Bytes()[s.sendAck:]
				}

				// log.Printf("retransmitting /data/%d/%d/%s/", s.ID, pl.Pos, string(Escape(pl.Data)))
				fmt.Fprintf(s.w, "/data/%d/%d/%s/", s.ID, pl.Pos, string(Escape(pl.Data)))
				sendTime = time.Now()
			}
		}
	}
}
