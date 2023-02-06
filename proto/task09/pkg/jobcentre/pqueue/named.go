package pqueue

import (
	"encoding/json"
	"sync"
)

// Job represents an internal job storage.
type Job struct {
	ID       uint64
	Priority int
	Queue    string
	Body     json.RawMessage
}

// Named is a named priority queues.
type Named struct {
	mu     sync.Mutex
	queues map[string]*PQueue // queueName->priorityQueue
}

// New creates a new Named Priority Queue.
func New() *Named {
	return &Named{queues: make(map[string]*PQueue)}
}

// Enque enqueus the job to the named queue.
func (n *Named) Enque(job *Job) {
	n.mu.Lock()
	defer n.mu.Unlock()

	queue := job.Queue
	qq, ok := n.queues[queue]
	if !ok {
		qq = NewPriorityQueue(queue)
		n.queues[queue] = qq
	}

	qq.Enque(job)
}

// Deque removes the highest priority job from any of the provided queues.
func (n *Named) Deque(queues []string) *Job {
	n.mu.Lock()
	defer n.mu.Unlock()

	queue := n.maxQueue(queues)
	if queue == nil {
		return nil
	}

	return queue.DequeMax()
}

// Delete deletes the job with id in every named queue.
func (n *Named) Delete(id uint64) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	var ok bool
	for _, pq := range n.queues {
		ok = pq.Delete(id) || ok
	}

	return ok
}

// maxQueue returns the queue in the provided list that has the job with the highest priority
func (n *Named) maxQueue(queues []string) *PQueue {
	maxPri := -1
	var maxQue *PQueue

	for _, q := range queues {
		if pq, ok := n.queues[q]; ok {
			if pq.HighestPriority() > maxPri {
				maxPri = pq.HighestPriority()
				maxQue = pq
			}
		}
	}

	return maxQue
}
