package pqueue_test

import (
	"testing"

	"github.com/matryer/is"

	"proto/pkg/jobcentre/pqueue"
)

func TestPriQueue_Enque_DequeMax(t *testing.T) {
	tests := []struct {
		name    string
		jobs    []*pqueue.Job
		wantPri []int
	}{
		{
			name: "should enque several jobs in then deque them in correct order",
			jobs: []*pqueue.Job{
				{Priority: 5},
				{Priority: 7},
				{Priority: 4},
				{Priority: 8},
				{Priority: 2},
				{Priority: 10},
			},
			wantPri: []int{10, 8, 7, 5, 4, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			p := pqueue.NewPriorityQueue("foobar")
			for _, job := range tt.jobs {
				p.Enque(job)
			}

			for i := 0; i < len(tt.jobs); i++ {
				job := p.DequeMax()
				// t.Logf("%d %v", i, job)
				is.True(job != nil)
				is.Equal(job.Priority, tt.wantPri[i])
			}
		})
	}
}
