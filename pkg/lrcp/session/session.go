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
	chunkSize          = 400
)

// Payload stores a storable payload.
type Payload struct {
	Pos  int
	Data []byte
}

// Session is a session managing unit
type Session struct {
	ID      int
	w       io.Writer
	closed  bool
	closeFn func()

	// RECEIVE
	rcvAcked int       // consequtive data we acked so far
	rcvLast  time.Time // time of the last received ack

	// SEND
	sendBytes bytes.Buffer
	sendAcked int // sent data acked
	sendCh    chan Payload

	// application layer
	app *app.App
}

// New creates a new Session instance
func New(ctx context.Context, w io.Writer, id int) *Session {
	ctx, cancel := context.WithCancel(ctx)

	s := &Session{
		w:       w,
		ID:      id,
		rcvLast: time.Now(),
		sendCh:  make(chan Payload, 1000),
		app:     &app.App{},
	}

	s.closeFn = func() {
		cancel()
		close(s.sendCh)

		// set last ack time to the past to sweep the connection.
		s.rcvLast = time.Now().Add(-1 * time.Hour)
		s.closed = true
	}

	go func() {
		for payload := range s.sendCh {
			select {
			case <-ctx.Done():
				return

			default:
			}

			s.sendData(ctx, payload.Pos, payload.Data)
			s.retransmit(ctx)
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

	if length < s.sendAcked {
		return nil

	} else if length > s.sendBytes.Len() {
		s.SendClose(ctx)

	} else {
		// if not all data received - send the remainder or if done
		s.sendAcked = length
	}

	return nil
}

// HandleData handles 'data' message
func (s *Session) HandleData(ctx context.Context, pos int, data []byte) error {
	log.Printf("DATA: sid:%d pos:%d data:%s", s.ID, pos, string(data))
	s.notify()

	if pos > s.rcvAcked {
		// do not have all the data - send duplicate ack for data we have.
		s.SendAck(ctx, s.rcvAcked)
		s.SendAck(ctx, s.rcvAcked)

	} else if pos < s.rcvAcked {
		s.SendAck(ctx, s.rcvAcked)

		if pos < s.sendBytes.Len() && !s.closed {
			// resend chunk
			s.sendCh <- Payload{
				Pos:  pos,
				Data: s.sendBytes.Bytes()[pos:],
			}
		}

	} else {
		// have all data up to pos + the current buffer
		buf := Unescape(data)
		s.rcvAcked += int(len(buf))
		s.SendAck(ctx, s.rcvAcked)
		s.processAppData(ctx, buf)
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
	if !s.closed {
		s.closeFn()
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

func (s *Session) processAppData(ctx context.Context, buf []byte) {
	if s.closed {
		return
	}

	_, _ = s.app.Write(buf) // pass data to the application layer.
	buf, err := io.ReadAll(s.app)
	if err != nil {
		log.Printf("[%d] Cannot read from the app layer: %s", s.ID, err.Error())
		return
	}

	s.sendCh <- Payload{
		Pos:  s.sendBytes.Len(),
		Data: buf,
	}
	s.sendBytes.Write(buf)
}

func (s *Session) sendData(ctx context.Context, pos int, data []byte) {
	rdata := bytes.NewReader(data)
	buf := make([]byte, chunkSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := rdata.Read(buf)
		if n == 0 {
			break
		}

		if err != nil {
			log.Printf("[%d] Cannot read from the send buffer: %s", s.ID, err.Error())
		}

		fmt.Fprintf(s.w, "/data/%d/%d/%s/", s.ID, pos, string(Escape(buf[:n])))

		pos += n
	}
}

func (s *Session) retransmit(ctx context.Context) {
	timer := time.NewTimer(retransmitInterval)
	defer func() {
		timer.Stop()
	}()

	select {
	case <-ctx.Done():
		return

	case <-timer.C:
		if s.sendAcked >= s.sendBytes.Len() || s.closed {
			return
		}

		s.sendData(ctx, s.sendAcked, s.sendBytes.Bytes()[s.sendAcked:])
		timer.Reset(retransmitInterval)
	}
}
