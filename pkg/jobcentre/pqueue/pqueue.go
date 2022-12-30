package pqueue

import (
	"sort"
	"sync"
)

// PQueue represents a priority queue storage.
type PQueue struct {
	mu         sync.Mutex
	Name       string
	priorities []int
	items      map[int][]*Job
}

// NewPriorityQueue creates a new PriQueue instance.
func NewPriorityQueue(name string) *PQueue {
	return &PQueue{
		Name:  name,
		items: make(map[int][]*Job),
	}
}

// Enque adds a new job to the queue.
func (p *PQueue) Enque(job *Job) {
	p.mu.Lock()
	defer p.mu.Unlock()

	pri := job.Priority
	// has priority
	_, ok := p.items[job.Priority]
	if !ok {
		// no such priority yet
		p.priorities = append(p.priorities, pri)
		sort.Ints(p.priorities)
		p.items[pri] = []*Job{job}
		return
	}

	p.items[pri] = append(p.items[pri], job)
}

// Peek return the next job with the given priority.
func (p *PQueue) Peek(pri int) *Job {
	_, ok := p.items[pri]
	if !ok {
		return nil
	}

	if len(p.items[pri]) == 0 {
		return nil
	}

	return p.items[pri][0]
}

// PeekMax returns the next job with the highest proirity but does not remove it from the queue.
func (p *PQueue) PeekMax() *Job {
	if len(p.priorities) == 0 {
		// empty queue
		return nil
	}

	return p.Peek(p.priorities[len(p.priorities)-1])
}

// Deque pull a job from the queue with the given priority
func (p *PQueue) Deque(pri int) *Job {
	p.mu.Lock()
	defer p.mu.Unlock()

	job := p.Peek(pri)

	if job != nil {
		p.items[pri] = p.items[pri][1:]
		if len(p.items[job.Priority]) == 0 {
			p.priorities = p.priorities[:len(p.priorities)-1]
			delete(p.items, job.Priority)
		}
	}

	return job
}

// DequeMax pull the next item from the queue with the highest priority.
func (p *PQueue) DequeMax() *Job {
	p.mu.Lock()
	defer p.mu.Unlock()

	job := p.PeekMax()
	if job != nil {
		p.items[job.Priority] = p.items[job.Priority][1:]
		if len(p.items[job.Priority]) == 0 {
			p.priorities = p.priorities[:len(p.priorities)-1]
			delete(p.items, job.Priority)
		}
	}

	return job
}

// Delete deletes all instances of the job with given ID everywhere in the queue regardless of the
// priorities.
func (p *PQueue) Delete(id uint64) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for pri, queue := range p.items {
		for i, job := range queue {
			if job.ID == id {
				p.items[pri] = append(queue[:i], queue[i+1:]...)

				if len(p.items[pri]) == 0 {
					delete(p.items, pri)

					for i, pr := range p.priorities {
						if pr == pri {
							p.priorities = append(p.priorities[:i], p.priorities[i+1:]...)
							break
						}
					}
				}

				return true
			}
		}
	}

	return false
}

// HighestPriority returns the highest priority number stored in this priority-queue instance.
func (p *PQueue) HighestPriority() int {
	job := p.PeekMax()
	if job == nil {
		return -1
	}

	return job.Priority
}
