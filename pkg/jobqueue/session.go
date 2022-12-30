package jobqueue

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"sync/atomic"
)

var debug = os.Getenv("DEBUG") != ""

// Session represents a client connection
type Session struct {
	jq      *JobQueue
	rw      io.ReadWriter
	working map[uint64]*Job
}

// Request represents a connection request
type Request struct {
	ID       *uint64         `json:"id,omitempty"`
	Request  string          `json:"request"`
	Queues   []string        `json:"queues,omitempty"`
	Queue    string          `json:"queue,omitempty"`
	Wait     bool            `json:"wait,omitempty"`
	Priority *int            `json:"pri"`
	Job      json.RawMessage `json:"job,omitempty"`
}

// Response represents a connection response
type Response struct {
	ID       uint64          `json:"id,omitempty"`
	Status   string          `json:"status"`
	Error    string          `json:"error,omitempty"`
	Job      json.RawMessage `json:"job,omitempty"`
	Priority int             `json:"pri,omitempty"`
	Queue    string          `json:"queue,omitempty"`
}

// Job represents an internal job storage.
type Job struct {
	ID       uint64
	Priority int
	Queue    string
	Body     json.RawMessage
}

// NewSession creates a new session
func NewSession(ctx context.Context, jq *JobQueue, rw io.ReadWriter) *Session {
	return &Session{
		jq:      jq,
		rw:      rw,
		working: make(map[uint64]*Job),
	}
}

// Handle handles the client connection from start to finish.
func (s *Session) Handle(ctx context.Context) {
	scanner := bufio.NewScanner(s.rw)
	defer func() {
		for id := range s.working {
			s.abortJob(id)
		}
	}()

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if debug {
			log.Printf("request: %s", scanner.Text())
		}

		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			s.sendError(err)
			continue
		}

		if req.Request == "" {
			s.sendError(errors.New("invalid request - request field missing"))
			continue
		}

		switch req.Request {
		case "put":
			if err := s.validatePut(&req); err != nil {
				s.sendError(err)
			}

			job := &Job{
				ID:       atomic.AddUint64(&s.jq.idCounter, 1),
				Priority: *req.Priority,
				Queue:    req.Queue,
				Body:     req.Job,
			}

			if ok := s.updateWaiting(req.Queue, job); !ok {
				s.jq.GetQueue(req.Queue).Enque(job)
			}

			s.send(&Response{Status: "ok", ID: job.ID})

		case "get":
			if err := s.validateGet(&req); err != nil {
				s.sendError(err)
			}

			pq, ok := s.jq.MaxQueue(req.Queues)
			if !ok {
				if req.Wait {
					for _, queue := range req.Queues {
						s.jq.AddWaiting(queue, s.rw)
					}
				} else {
					s.send(&Response{Status: "no-job"})
				}
				continue
			}

			var job *Job
			if req.Priority != nil {
				job, ok = pq.Deque(*req.Priority)
			} else {
				job, ok = pq.DequeMax()
			}

			if ok {
				s.jq.StartJob(job.ID)
				s.working[job.ID] = job

			} else {
				if req.Wait {
					s.jq.AddWaiting(pq.Name, s.rw)
				} else {
					s.send(&Response{Status: "no-job"})
				}
				continue
			}

			s.send(&Response{
				Status:   "ok",
				ID:       job.ID,
				Priority: job.Priority,
				Queue:    pq.Name,
				Job:      job.Body,
			})

		case "delete":
			if err := s.validateID(&req); err != nil {
				s.sendError(err)
			}

			if _, ok := s.working[*req.ID]; ok {
				// Abort current
				s.abortJob(*req.ID)
			}

			ok := false
			for _, pq := range s.jq.queues {
				ok = ok || pq.Delete(*req.ID)
			}

			ok = ok || s.jq.IsRunning(*req.ID)

			if ok {
				s.send(&Response{Status: "ok"})
			} else {
				s.send(&Response{Status: "no-job"})
			}

		case "abort":
			if err := s.validateID(&req); err != nil {
				s.sendError(err)
			}

			if ok := s.abortJob(*req.ID); ok {
				s.send(&Response{Status: "ok"})
			} else {
				s.send(&Response{Status: "no-job"})
			}
		}
	}
}

func (s *Session) abortJob(id uint64) bool {
	if job, ok := s.working[id]; ok {
		if ok := s.updateWaiting(job.Queue, job); !ok {
			s.jq.GetQueue(job.Queue).Enque(job)
		}
		delete(s.working, id)
		s.jq.StopJob(id)
		return true
	}

	return false
}

func (s *Session) sendError(err error) {
	enc := json.NewEncoder(s.rw)

	response := Response{
		Status: "error",
		Error:  err.Error(),
	}

	if err := enc.Encode(&response); err != nil {
		log.Printf("Could not send response: %v", response)
	}
}

func (s *Session) send(res *Response) {
	enc := json.NewEncoder(s.rw)

	if debug {
		jsonb, _ := json.Marshal(res)
		log.Printf("response: %s", string(jsonb))
	}

	if err := enc.Encode(res); err != nil {
		log.Printf("Could not send response: %v", res)
	}
}

func (s *Session) updateWaiting(queue string, job *Job) bool {
	client, ok := s.jq.Waiting(queue)
	if !ok {
		return false
	}

	enc := json.NewEncoder(client)

	res := &Response{
		Status:   "ok",
		ID:       job.ID,
		Priority: job.Priority,
		Queue:    queue,
		Job:      job.Body,
	}

	s.working[job.ID] = job
	s.jq.StartJob(job.ID)

	if err := enc.Encode(res); err != nil {
		log.Printf("Could not send response: %v", res)
		return s.updateWaiting(queue, job)
	}

	return true
}

func (s *Session) validatePut(req *Request) error {
	if req.Job == nil {
		return errors.New("invalid put request - job field missing")
	}

	if req.Priority == nil {
		return errors.New("invalid put request - pri field missing")
	}

	if req.Queue == "" {
		return errors.New("invalid put request - queue field missing")
	}

	return nil
}

func (s *Session) validateGet(req *Request) error {
	if len(req.Queues) == 0 {
		return errors.New("invalid get request - queues field missing")
	}

	return nil
}

func (s *Session) validateID(req *Request) error {
	if req.ID == nil {
		return errors.New("invalid request - id field missing")
	}

	return nil
}
