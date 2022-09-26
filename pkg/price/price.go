package price

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"proto/pkg/price/message"
)

type Handler struct {
	IO       io.ReadWriter
	payloads []message.Payload
}

func (m *Handler) Handle(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("PriceHandle cancelled")
			break

		default:
		}

		msg := &message.Msg{}
		if _, err := msg.ReadFrom(m.IO); err != nil {
			if !errors.Is(err, io.EOF) {
				log.Println("Error readfrom:", err)
			}

			return
		}

		switch msg.Type {
		case message.TypeInsert:
			if err := m.handleInsert(ctx, msg); err != nil {
				log.Println("Error:", err)
				return
			}

		case message.TypeQuery:
			if err := m.handleQuery(ctx, msg); err != nil {
				log.Println("Error:", err)
				return
			}

		default:
			log.Printf("Corrupt message: %#v", msg)
		}
	}
}

func (m *Handler) handleInsert(ctx context.Context, msg *message.Msg) error {
	log.Printf("handling insert msg: %d %d", msg.Payload.Time, msg.Payload.Data)
	m.payloads = append(m.payloads, msg.Payload)
	return nil
}

func (m *Handler) handleQuery(ctx context.Context, msg *message.Msg) error {
	log.Printf("handling query msg: %d %d", msg.Payload.Time, msg.Payload.Data)

	if msg.Payload.Time > msg.Payload.Data {
		log.Printf("Invalid time range [%d-%d]", msg.Payload.Time, msg.Payload.Data)
		return m.sendResponse(0) // return 0 on invalid time range.
	}

	prices := make([]int32, 0, len(m.payloads))
	for _, pl := range m.payloads {
		if pl.Time >= msg.Payload.Time && pl.Time <= msg.Payload.Data {
			prices = append(prices, pl.Data)
		}
	}

	return m.sendResponse(m.meanPrice(prices))
}

func (m *Handler) sendResponse(data int32) error {
	log.Print("Sending response: ", data)
	return binary.Write(m.IO, binary.BigEndian, data)
}

func (m *Handler) meanPrice(data []int32) int32 {
	if len(data) == 0 {
		return 0
	}

	var sum int64
	for _, v := range data {
		sum += int64(v)
	}

	return int32(sum / int64(len(data)))
}
