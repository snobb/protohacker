package heap_test

import (
	"proto/pkg/heap"
	"testing"

	"github.com/matryer/is"
)

func TestHeap_PushPop(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{
			name: "should push at random order and pop ordered",
			in:   []int{1, 15, 2, 7, 43, 81},
			want: []int{81, 43, 15, 7, 2, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)

			h := &heap.Heap[int]{}

			for _, v := range tt.in {
				h.Push(v)
			}

			var res []int
			for !h.Empty() {
				v := h.Pop()
				t.Logf("value: %d", v)
				res = append(res, v)
			}

			is.Equal(res, tt.want)
		})
	}
}
