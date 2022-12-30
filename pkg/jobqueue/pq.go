package jobqueue

import (
	"sort"
	"sync"
)

// PriQueue represents a priority queue storage.
type PriQueue struct {
	mu         sync.Mutex
	Name       string
	priorities []int
	jobs       map[int][]*Job
}

// NewPriorityQueue creates a new PriQueue instance.
func NewPriorityQueue(name string) *PriQueue {
	return &PriQueue{
		Name: name,
		jobs: make(map[int][]*Job),
	}
}

// Enque adds a new job to the queue.
func (p *PriQueue) Enque(job *Job) {
	p.mu.Lock()
	defer p.mu.Unlock()

	pri := job.Priority
	// has priority
	_, ok := p.jobs[job.Priority]
	if !ok {
		// no such priority yet
		p.priorities = append(p.priorities, pri)
		sort.Ints(p.priorities)
		p.jobs[pri] = []*Job{job}
		return
	}

	p.jobs[pri] = append(p.jobs[pri], job)
}

// Peek return the next job with the given priority.
func (p *PriQueue) Peek(pri int) (*Job, bool) {
	_, ok := p.jobs[pri]
	if !ok {
		return nil, false
	}

	if len(p.jobs[pri]) == 0 {
		return nil, false
	}

	item := p.jobs[pri][0]

	return item, true
}

// PeekMax returns the next job with the highest proirity but does not remove it from the queue.
func (p *PriQueue) PeekMax() (*Job, bool) {
	if len(p.priorities) == 0 {
		// empty queue
		return nil, false
	}

	return p.Peek(p.priorities[len(p.priorities)-1])
}

// Deque pull a job from the queue with the given priority
func (p *PriQueue) Deque(pri int) (*Job, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	job, ok := p.Peek(pri)
	if ok {
		p.jobs[pri] = p.jobs[pri][1:]
		if len(p.jobs[job.Priority]) == 0 {
			p.priorities = p.priorities[:len(p.priorities)-1]
			delete(p.jobs, job.Priority)
		}
	}

	return job, ok
}

// DequeMax pull the next item from the queue with the highest priority.
func (p *PriQueue) DequeMax() (*Job, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	job, ok := p.PeekMax()
	if ok {
		p.jobs[job.Priority] = p.jobs[job.Priority][1:]
		if len(p.jobs[job.Priority]) == 0 {
			p.priorities = p.priorities[:len(p.priorities)-1]
			delete(p.jobs, job.Priority)
		}
	}

	return job, ok
}

// Delete deletes all instances of the job with given ID everywhere in the queue regardless of the
// priorities.
func (p *PriQueue) Delete(id uint64) bool {
	for pri, queue := range p.jobs {
		for i, job := range queue {
			if job.ID == id {
				p.mu.Lock()
				p.jobs[pri] = append(queue[:i], queue[i+1:]...)

				if len(p.jobs[pri]) == 0 {
					delete(p.jobs, pri)

					for i, pr := range p.priorities {
						if pr == pri {
							p.priorities = append(p.priorities[:i], p.priorities[i+1:]...)
							break
						}
					}
				}

				p.mu.Unlock()
				return true
			}
		}
	}

	return false
}

// HighestPriority returns the highest priority number stored in this priority-queue instance.
func (p *PriQueue) HighestPriority() int {
	job, ok := p.PeekMax()
	if !ok {
		return -1
	}

	return job.Priority
}
