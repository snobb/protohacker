package lrcp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	"proto/pkg/lrcp/session"
)

// LRCP constants
const (
	TypeConnect = "connect"
	TypeAck     = "ack"
	TypeData    = "data"
	TypeClose   = "close"
)

// LRCP is a lrcp protocol handler.
type LRCP struct {
	mu       sync.Mutex
	sessions map[int]*session.Session
}

// New creates a new LRCP instance.
func New(ctx context.Context) *LRCP {
	lrcp := &LRCP{
		sessions: make(map[int]*session.Session),
	}

	// Check expired sessions.
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Printf("cancelled...")
				return

			case <-ticker.C:
				for _, session := range lrcp.sessions {
					lrcp.SweepExpired(ctx, session)
				}
			}
		}
	}()

	return lrcp
}

// Handle handles a single incoming udp datagram.
func (l *LRCP) Handle(ctx context.Context, w io.Writer, buf []byte) {
	tokens, err := normalise(buf)
	if err != nil {
		log.Printf("Validation error: %s", err.Error())
		return
	}

	mtype := string(tokens[0])
	sid, err := session.ParseID(tokens[1])
	if err != nil {
		log.Printf("invalid session ID: %s", string(tokens[1]))
		return
	}

	sess, ok := l.sessions[sid]
	if !ok {
		sess = session.New(ctx, w, sid)

		if mtype == TypeConnect {
			l.AddSession(sid, sess)
		} else {
			sess.Close()
			sess.SendClose(ctx)
			return
		}
	}

	switch mtype {
	case TypeConnect:
		if len(tokens) != 2 {
			log.Printf("CONNECT: invalid message: %s, %d tokens expected", string(buf), 2)
			return
		}

		if err := sess.HandleConnect(ctx); err != nil {
			log.Printf("CONNECT: could not handle message: %s", err.Error())
			return
		}

	case TypeClose:
		if len(tokens) != 2 {
			log.Printf("CLOSE: invalid message: %s, %d tokens expected", string(buf), 2)
			return
		}

		if err := sess.HandleClose(ctx); err != nil {
			log.Printf("CLOSE: could not handle message: %s", err.Error())
			return
		}

		delete(l.sessions, sess.ID)

	case TypeAck:
		if len(tokens) != 3 {
			log.Printf("ACK: invalid message: %s, %d tokens expected", string(buf), 3)
			return
		}

		length, err := strconv.Atoi(string(tokens[2]))
		if err != nil {
			log.Printf("ACK: could not parse length: %s", err.Error())
			return
		}

		if length < 0 {
			log.Printf("ACK: invalid length: %d", length)
			return
		}

		if err := sess.HandleAck(ctx, length); err != nil {
			log.Printf("ACK: could not handle message: %s", err.Error())
			return
		}

	case TypeData:
		if len(tokens) != 4 {
			log.Printf("DATA: invalid message: %s, %d tokens expected", string(buf), 4)
			return
		}

		pos, err := strconv.Atoi(string(tokens[2]))
		if err != nil {
			log.Printf("DATA: could not parse pos: %s", err.Error())
			return
		}

		if pos < 0 {
			log.Printf("DATA: invalid pos: %d", pos)
			return
		}

		if err := sess.HandleData(ctx, pos, tokens[3]); err != nil {
			log.Printf("DATA: could not handle message: %s", err.Error())
			return
		}

	default:
		log.Printf("invalid message type: %s", mtype)
		return
	}
}

// AddSession adds session to the storage safely.
func (l *LRCP) AddSession(sid int, session *session.Session) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.sessions[sid] = session
}

// SweepExpired checks and clears if it's expired.
func (l *LRCP) SweepExpired(ctx context.Context, session *session.Session) {
	if session.Closed() {
		l.mu.Lock()
		delete(l.sessions, session.ID)
		l.mu.Unlock()
	}

	if session.Expired() {
		log.Printf("Session expired: %d", session.ID)
		session.Close()
	}
}

func normalise(buf []byte) ([][]byte, error) {
	if len(buf) < 2 || len(buf) > 1000 {
		return nil, errors.New("invalid size")
	}

	if buf[0] != '/' || buf[len(buf)-1] != '/' {
		return nil, fmt.Errorf("invalid message: %s", string(buf))
	}

	tokens := bytes.SplitN(buf[1:len(buf)-1], []byte{'/'}, 3)
	if len(tokens) < 2 {
		return nil, fmt.Errorf("invalid split: %#v", tokens)
	}

	if len(tokens) == 2 {
		return tokens, nil
	}

	tokens1 := bytes.SplitN(tokens[2], []byte{'/'}, 2)
	if len(tokens1) == 1 {
		return append(tokens[:2], tokens1[0]), nil
	}

	escaped := bytes.Count(tokens1[1], []byte("\\/"))
	wild := bytes.Count(tokens1[1], []byte("/"))

	if escaped != wild {
		return nil, fmt.Errorf("invalid message: %s", string(buf))
	}

	return append(tokens[:2], tokens1...), nil
}
