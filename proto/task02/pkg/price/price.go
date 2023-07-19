package price

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"log"

	"proto/task02/pkg/price/message"
)

// Handler handles price server
type Handler struct {
	payloads []message.Payload
}

// Handle handles a single connection
func (m *Handler) Handle(ctx context.Context, rw io.ReadWriter) {
	for {
		select {
		case <-ctx.Done():
			log.Println("PriceHandle cancelled")
			return

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
			log.Printf(">>> INS %d %d", msg.Payload.Time, msg.Payload.Data)
			m.payloads = append(m.payloads, msg.Payload)

		case message.TypeQuery:
			log.Printf(">>> QRY %d %d", msg.Payload.Time, msg.Payload.Data)
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
	if msg.Payload.Time > msg.Payload.Data {
		log.Printf("Invalid time range [%d-%d]", msg.Payload.Time, msg.Payload.Data)
		return 0
	}

	prices := make([]float64, 0, len(m.payloads))
	for _, pl := range m.payloads {
		if pl.Time >= msg.Payload.Time && pl.Time <= msg.Payload.Data {
			prices = append(prices, float64(pl.Data))
		}
	}

	return int32(average(prices))
}

func average(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	var avg float64 = 0
	for t, v := range data {
		avg += (v - avg) / float64(t+1)
	}

	return avg
}

func sendResponse(data int32, rw io.ReadWriter) error {
	log.Printf("<<< RES %d", data)
	return binary.Write(rw, binary.BigEndian, data)
}
