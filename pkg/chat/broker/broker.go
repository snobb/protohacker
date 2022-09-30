package broker

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
)

// Broker is a message broker for the chat
type Broker struct {
	sync.Mutex
	clients map[string]io.Writer
}

// New creates a new Broker instance.
func New() *Broker {
	return &Broker{clients: make(map[string]io.Writer)}
}

// Register a new connection to the broker.
func (b *Broker) Register(id string, w io.Writer) error {
	if _, ok := b.clients[id]; ok {
		return errors.New("client with the same name is already registered")
	}

	b.Lock()
	b.clients[id] = w
	b.Unlock()

	b.broadcast(id, fmt.Sprintf("* %s has entered the room\n", id))
	fmt.Fprintf(w, "* the room contains: %s\n", strings.Join(b.getNames(id), ", "))

	return nil
}

// Sends broadcasts the message to all attached connections.
func (b *Broker) Send(id string, message string) {
	b.broadcast(id, fmt.Sprintf("[%s] %s\n", id, message))
}

// Unregister the connection from the broker.
func (b *Broker) Unregister(id string) {
	b.Lock()
	delete(b.clients, id)
	b.Unlock()

	b.broadcast(id, fmt.Sprintf("* %s has left the room\n", id))
}

func (b *Broker) broadcast(id string, message string) {
	for k, w := range b.clients {
		if k == id {
			continue
		}

		if _, err := fmt.Fprint(w, message); err != nil {
			log.Printf("Could not notify %s session", k)
		}
	}
}

func (b *Broker) getNames(exclude string) []string {
	var names []string
	for k := range b.clients {
		if k == exclude {
			continue
		}

		names = append(names, k)
	}

	return names
}
