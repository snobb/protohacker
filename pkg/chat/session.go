package chat

import (
	"context"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"

	"proto/pkg/chat/broker"
	"proto/pkg/iotools"
)

type Session struct {
	name   string
	broker *broker.Broker
}

func NewSession(broker *broker.Broker) *Session {
	return &Session{broker: broker}
}

func (s *Session) Handle(ctx context.Context, rw io.ReadWriter) {
	fmt.Fprintln(rw, "Welcome to budgetchat! What shall I call you?")

	lines := iotools.GetLine(ctx, rw)
	s.name = strings.TrimSpace(string(<-lines))
	if !validate(s.name) {
		log.Printf("Invalid name: %s.", s.name)
		return
	}

	if err := s.broker.Register(s.name, rw); err != nil {
		log.Println("Error:", err.Error())
		return
	}
	defer s.broker.Unregister(s.name)

	for buf := range lines {
		select {
		case <-ctx.Done():
			break

		default:
		}

		line := strings.TrimSpace(string(buf))
		s.broker.Send(s.name, line)
	}
}

func validate(name string) bool {
	if len(name) < 1 {
		return false
	}

	matched, err := regexp.Match("^[a-zA-Z0-9]*$", []byte(name))
	if err != nil {
		log.Println("validation:", err.Error())
		return false
	}

	return matched
}
