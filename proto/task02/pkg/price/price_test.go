package price_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"testing"

	"github.com/matryer/is"

	"proto/common/pkg/tcpserver"
	"proto/task02/pkg/price"
	"proto/task02/pkg/price/message"
)

func TestHandle_Handle(t *testing.T) {
	tests := []struct {
		name    string
		inserts []message.Msg
		wantOut int32
	}{
		{
			name: "should return a correct query for the given dataset",
			inserts: []message.Msg{
				{
					Type:    message.TypeInsert,
					Payload: message.Payload{Time: 12345, Data: 101},
				},
				{
					Type:    message.TypeInsert,
					Payload: message.Payload{Time: 12346, Data: 102},
				},
				{
					Type:    message.TypeInsert,
					Payload: message.Payload{Time: 12347, Data: 100},
				},
				{
					Type:    message.TypeInsert,
					Payload: message.Payload{Time: 40960, Data: 5},
				},
				{
					Type:    message.TypeQuery,
					Payload: message.Payload{Time: 12345, Data: 16384},
				},
			},
			wantOut: 101,
		},
		{
			name: "should return a correct query for the given dataset",
			inserts: []message.Msg{
				{
					Type:    message.TypeInsert,
					Payload: message.Payload{Time: 355, Data: 1215},
				},
				{
					Type:    message.TypeInsert,
					Payload: message.Payload{Time: 977, Data: 674},
				},
				{
					Type:    message.TypeInsert,
					Payload: message.Payload{Time: 979, Data: 3128},
				},
				{
					Type:    message.TypeInsert,
					Payload: message.Payload{Time: 313, Data: 3178},
				},
				{
					Type:    message.TypeInsert,
					Payload: message.Payload{Time: 984, Data: 671},
				},
				{
					Type:    message.TypeInsert,
					Payload: message.Payload{Time: 987, Data: 672},
				},
				{
					Type:    message.TypeInsert,
					Payload: message.Payload{Time: 988, Data: 659},
				},
				{
					Type:    message.TypeQuery,
					Payload: message.Payload{Time: 971, Data: 988},
				},
			},
			// (674 + 671 + 672 + 659 + 3128) / 5
			wantOut: 1160,
		},
		{
			name: "should calculate correct mean for very large int values",
			inserts: []message.Msg{
				{
					Type:    message.TypeInsert,
					Payload: message.Payload{Time: 42, Data: math.MaxInt32},
				},
				{
					Type:    message.TypeInsert,
					Payload: message.Payload{Time: 43, Data: 42},
				},
				{
					Type:    message.TypeQuery,
					Payload: message.Payload{Time: 0, Data: 100},
				},
			},
			wantOut: int32(uint64(math.MaxInt32+42) / 2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			is := is.New(t)

			rec := &tcpserver.Recorder{In: &bytes.Buffer{}}
			for _, msg := range tt.inserts {
				err := binary.Write(rec.In, binary.BigEndian, msg)
				is.NoErr(err)
			}

			handler := price.Handler{}
			handler.Handle(ctx, rec)

			var res int32
			err := binary.Read(&rec.Out, binary.BigEndian, &res)
			if err != nil && !errors.Is(err, io.EOF) {
				t.Fatal(err)
			}

			is.Equal(tt.wantOut, res)
		})
	}
}
