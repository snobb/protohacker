package lrcp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
)

const (
	MinSessionID = 0
	MaxSessionID = 2147483648
	MinLength    = 0
	MaxLength    = 1000

	TypeConnect = "connect"
	TypeAck     = "ack"
	TypeData    = "data"
	TypeClose   = "close"
)

type Session struct {
	data []byte
}

type LRCP struct {
	sessions map[uint32]Session
}

func New() *LRCP {
	return &LRCP{
		sessions: make(map[uint32]Session),
	}
}

func (l *LRCP) Handle(ctx context.Context, w io.Writer, buf []byte) {
	if buf[0] != '/' || buf[len(buf)-1] != '/' {
		log.Printf("invalid message: %s", string(buf))
		return
	}

	tokens := bytes.Split(buf[1:len(buf)-1], []byte{'/'})

	mtype := string(tokens[0])
	sid, err := parseSessionID(tokens[1])
	if err != nil {
		log.Printf("invalid session ID: %s", string(tokens[1]))
		return
	}

	switch mtype {
	case TypeConnect:
		if len(tokens) != 2 {
			log.Printf("CONNECT: invalid message: %s, %d tokens expected", string(buf), 2)
			return
		}

		if err := l.handleConnect(ctx, sid); err != nil {
			log.Printf("CONNECT: could not handle message: %s", err.Error())
			return
		}

	case TypeClose:
		if len(tokens) != 2 {
			log.Printf("CLOSE: invalid message: %s, %d tokens expected", string(buf), 2)
			return
		}

		if err := l.handleClose(ctx, sid); err != nil {
			log.Printf("CLOSE: could not handle message: %s", err.Error())
			return
		}

	case TypeAck:
		if len(tokens) != 3 {
			log.Printf("ACK: invalid message: %s, %d tokens expected", string(buf), 3)
			return
		}

		length, err := parseInt(tokens[2])
		if err != nil {
			log.Printf("ACK: could not parse length: %s", err.Error())
			return
		}

		if err := l.handleAck(ctx, sid, length); err != nil {
			log.Printf("ACK: could not handle message: %s", err.Error())
			return
		}

	case TypeData:
		if len(tokens) != 4 {
			log.Printf("DATA: invalid message: %s, %d tokens expected", string(buf), 4)
			return
		}

		pos, err := parseInt(tokens[2])
		if err != nil {
			log.Printf("DATA: could not parse pos: %s", err.Error())
			return
		}

		if err := l.handleData(ctx, sid, pos, tokens[3]); err != nil {
			log.Printf("DATA: could not handle message: %s", err.Error())
			return
		}

	default:
		log.Printf("invalid message type: %s", mtype)
		return
	}
}

func (l *LRCP) handleConnect(ctx context.Context, sid int) error {
	log.Printf("CONNECT: sid:%d", sid)
	return nil
}

func (l *LRCP) handleClose(ctx context.Context, sid int) error {
	log.Printf("CLOSE: sid:%d", sid)

	return nil
}

func (l *LRCP) handleAck(ctx context.Context, sid int, length int) error {
	log.Printf("ACK: sid:%d length:%d", sid, length)
	return nil
}

func (l *LRCP) handleData(ctx context.Context, sid int, length int, data []byte) error {
	log.Printf("DATA: sid:%d length:%d data:%s", sid, length, string(data))
	return nil
}

// ================================================================================
func parseSessionID(buf []byte) (int, error) {
	sid, err := parseInt(buf)
	if err != nil {
		return -1, err
	}

	if int64(sid) < MinSessionID || int64(sid) > MaxSessionID {
		return -1, fmt.Errorf("the value is out of range: %s", string(buf))
	}

	return sid, nil
}

func parseInt(buf []byte) (int, error) {
	val, err := strconv.Atoi(string(buf))
	if err != nil {
		return -1, fmt.Errorf("cannot parse session id: %s", err.Error())
	}

	return val, nil
}
