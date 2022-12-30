package jobqueue

import (
	"context"
	"io"
	"sort"
	"sync"
)

// JobQueue represents the JobCentre queue.
type JobQueue struct {
	idCounter  uint64
	queues     map[string]*PriQueue       // queueName->priorityQueue
	waiting    map[string][]io.ReadWriter // queueName->conn
	muq        sync.Mutex
	muw        sync.Mutex
	inProgress map[uint64]bool
}

// New creates a new instance of JobQueue
func New(ctx context.Context) *JobQueue {
	return &JobQueue{
		queues:     make(map[string]*PriQueue),
		waiting:    make(map[string][]io.ReadWriter),
		inProgress: make(map[uint64]bool),
	}
}

// GetQueue returns the instance of priority queue for the given name if exists or creates
// otherwise.
func (j *JobQueue) GetQueue(queue string) *PriQueue {
	j.muq.Lock()
	defer j.muq.Unlock()

	qq, ok := j.queues[queue]
	if !ok {
		qq = NewPriorityQueue(queue)
		j.queues[queue] = qq
	}

	return qq
}

// MaxQueue returns the queue in the provided list that has the job with the highest priority
func (j *JobQueue) MaxQueue(queues []string) (*PriQueue, bool) {
	j.muq.Lock()
	defer j.muq.Unlock()

	pqs := make([]*PriQueue, 0, len(queues))
	for _, q := range queues {
		if pq, ok := j.queues[q]; ok {
			pqs = append(pqs, pq)
		}
	}

	if len(pqs) == 0 {
		return nil, false
	}

	if len(pqs) < 2 {
		return pqs[0], true
	}

	sort.Slice(pqs, func(i, j int) bool {
		return pqs[i].HighestPriority() > pqs[j].HighestPriority()
	})

	return pqs[0], true
}

// AddWaiting subscribe socket for a new job in the given queue.
func (j *JobQueue) AddWaiting(queue string, rw io.ReadWriter) {
	j.muw.Lock()
	defer j.muw.Unlock()

	lst, ok := j.waiting[queue]
	if !ok {
		j.waiting[queue] = []io.ReadWriter{rw}
		return
	}

	j.waiting[queue] = append(lst, rw)
}

// Waiting returns the next waiting client for the given queue if exists.
func (j *JobQueue) Waiting(queue string) (io.ReadWriter, bool) {
	j.muw.Lock()
	defer j.muw.Unlock()

	_, ok := j.waiting[queue]
	if !ok {
		return nil, false
	}

	if len(j.waiting[queue]) == 0 {
		return nil, false
	}

	client := j.waiting[queue][0]
	j.waiting[queue] = j.waiting[queue][1:]

	return client, true
}

// StartJob marks a job as in progress by one of the workers.
func (j *JobQueue) StartJob(id uint64) {
	j.inProgress[id] = true
}

// StopJob marks a job as not in progress by one of the workers.
func (j *JobQueue) StopJob(id uint64) {
	delete(j.inProgress, id)
}

// IsRunning returns if the job with given ID is in progress by some workers
func (j *JobQueue) IsRunning(id uint64) bool {
	_, ok := j.inProgress[id]
	return ok
}
