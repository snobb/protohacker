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
	payloads []message.Payload
}

func (m *Handler) Handle(ctx context.Context, rw io.ReadWriter) {
	for {
		select {
		case <-ctx.Done():
			log.Println("PriceHandle cancelled")
			break

		default:
		}

		msg := &message.Msg{}
		if _, err := msg.ReadFrom(rw); err != nil {
			if !errors.Is(err, io.EOF) {
				log.Println("Error readfrom:", err.Error())
			}
			return
		}

		switch msg.Type {
		case message.TypeInsert:
			log.Printf("handling insert msg: %d %d", msg.Payload.Time, msg.Payload.Data)
			m.payloads = append(m.payloads, msg.Payload)

		case message.TypeQuery:
			mean := m.handleQuery(ctx, msg)
			if err := sendResponse(mean, rw); err != nil {
				log.Println("Error sendResponse:", err.Error())
				return
			}

		default:
			log.Printf("Corrupt message: %#v", msg)
		}
	}
}

func (m *Handler) handleQuery(ctx context.Context, msg *message.Msg) int32 {
	log.Printf("handling query msg: %d %d", msg.Payload.Time, msg.Payload.Data)

	if msg.Payload.Time > msg.Payload.Data {
		log.Printf("Invalid time range [%d-%d]", msg.Payload.Time, msg.Payload.Data)
		return 0
	}

	prices := make([]int32, 0, len(m.payloads))
	for _, pl := range m.payloads {
		if pl.Time >= msg.Payload.Time && pl.Time <= msg.Payload.Data {
			prices = append(prices, pl.Data)
		}
	}

	return meanPrice(prices)
}

func meanPrice(data []int32) int32 {
	if len(data) == 0 {
		return 0
	}

	var sum int64
	for _, v := range data {
		sum += int64(v)
	}

	return int32(sum / int64(len(data)))
}

func sendResponse(data int32, rw io.ReadWriter) error {
	log.Print("Sending response: ", data)
	return binary.Write(rw, binary.BigEndian, data)
}
