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
	chunkSize          = 300
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
	rcvAcked int       // consequtive data we acked so far
	rcvLast  time.Time // time of the last received ack

	// SEND
	sendBuffer    bytes.Buffer
	sendAcked     int // sent data acked
	sendTotal     int // total bytes sent so far
	sendRtxCancel func()

	// application layer
	app *app.App
}

// New creates a new Session instance
func New(ctx context.Context, w io.Writer, id int) *Session {
	return &Session{
		w:       w,
		ID:      id,
		rcvLast: time.Now(),
		app:     &app.App{},
	}
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

	if length < s.sendAcked {
		return nil

	} else if length > s.sendTotal {
		s.SendClose(ctx)

	} else {
		// if not all data received - send the remainder
		// or if done
		s.sendAcked = length
	}

	return nil
}

// HandleData handles 'data' message
func (s *Session) HandleData(ctx context.Context, pos int, data []byte) error {
	log.Printf("DATA: sid:%d pos:%d data:%s", s.ID, pos, string(data))
	s.notify()

	buf := Unescape(data)

	if pos > s.rcvAcked {
		// do not have all the data - send duplicate ack for data we have.
		s.SendAck(ctx, s.rcvAcked)
		s.SendAck(ctx, s.rcvAcked)

	} else if pos < s.rcvAcked {
		s.SendAck(ctx, s.rcvAcked)
		// resent chunk
		if pos < s.sendBuffer.Len() && s.sendRtxCancel == nil {
			ctx, s.sendRtxCancel = context.WithCancel(ctx)
			go s.retransmit(ctx)
		}

	} else {
		// have all data up to pos + the current buffer
		s.rcvAcked += int(len(data))
		s.SendAck(ctx, s.rcvAcked)
		_, _ = s.app.Write(buf) // pass data to the application layer.
		s.sendAppData(ctx)
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

// Close closes the session
func (s *Session) Close() {
	s.closed = true

	// set last ack time to the past to sweep the connection.
	s.rcvLast = time.Now().Add(-1 * time.Hour)
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

func (s *Session) sendAppData(ctx context.Context) {
	buf := make([]byte, chunkSize)

	for {
		n, err := s.app.Read(buf)
		if n == 0 {
			break
		}

		if err != nil {
			log.Printf("[%d] Cannot read from the app layer: %s", s.ID, err.Error())
		}

		log.Printf("/data/%d/%d/%s/", s.ID, s.sendTotal, string(Escape(buf[:n])))
		fmt.Fprintf(s.w, "/data/%d/%d/%s/", s.ID, s.sendTotal, string(Escape(buf[:n])))

		s.sendTotal += int(n)
		s.sendBuffer.Write(buf[:n])
	}

	if s.sendRtxCancel == nil {
		ctx, s.sendRtxCancel = context.WithCancel(ctx)
		go s.retransmit(ctx)
	}
}

func (s *Session) retransmit(ctx context.Context) {
	if s.sendRtxCancel != nil {
		return // retransmit is already in progress
	}

	ticker := time.NewTicker(retransmitInterval)
	defer ticker.Stop()
	defer func() {
		s.sendRtxCancel = nil
	}()

	select {
	case <-ctx.Done():
		return

	case <-ticker.C:
		if s.sendAcked < s.sendTotal {
			data := bytes.NewReader(s.sendBuffer.Bytes()[s.sendAcked:])
			buf := make([]byte, chunkSize)
			pos := s.sendAcked

			for {
				n, err := data.Read(buf)
				if n == 0 {
					break
				}

				if err != nil {
					log.Printf("[%d] Cannot read from the send buffer: %s", s.ID, err.Error())
				}

				log.Printf("retransmit /data/%d/%d/%s/", s.ID, pos, string(Escape(buf[:n])))
				fmt.Fprintf(s.w, "/data/%d/%d/%s/", s.ID, pos, string(Escape(buf[:n])))

				pos += n
			}

		} else {
			return
		}
	}
}
