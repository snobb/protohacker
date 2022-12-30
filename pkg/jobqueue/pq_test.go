package jobqueue_test

import (
	"testing"

	"proto/pkg/jobqueue"

	"github.com/matryer/is"
)

func TestPriQueue_Enque_DequeMax(t *testing.T) {
	tests := []struct {
		name    string
		jobs    []*jobqueue.Job
		wantPri []int
	}{
		{
			name: "should enque several jobs in then deque them in correct order",
			jobs: []*jobqueue.Job{
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
			p := jobqueue.NewPriorityQueue("foobar")
			for _, job := range tt.jobs {
				go p.Enque(job)
			}

			for i := 0; i < len(tt.jobs); i++ {
				job, ok := p.DequeMax()
				t.Log(job)
				is.True(ok)
				is.Equal(job.Priority, tt.wantPri[i])
			}
		})
	}
}
