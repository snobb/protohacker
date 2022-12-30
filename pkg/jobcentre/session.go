package jobcentre

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"sync/atomic"

	"proto/pkg/jobcentre/pqueue"
)

var debug = os.Getenv("DEBUG") != ""

// global state of the jobcentre server
var (
	idCounter uint64
	store     = pqueue.New()
	running   = make(map[uint64]bool)
	waiting   = make(map[string][]io.Writer) // queueName->conn
	mur       sync.Mutex                     // running
	muw       sync.Mutex                     // waiting
)

// Session represents a client connection context
type Session struct {
	rw      io.ReadWriter
	working map[uint64]*pqueue.Job
}

// request represents a server request
type request struct {
	ID       *uint64         `json:"id,omitempty"`
	Request  string          `json:"request"`
	Queues   []string        `json:"queues,omitempty"`
	Queue    string          `json:"queue,omitempty"`
	Wait     bool            `json:"wait,omitempty"`
	Priority *int            `json:"pri"`
	Job      json.RawMessage `json:"job,omitempty"`
}

// response represents a server response
type response struct {
	ID       uint64          `json:"id,omitempty"`
	Status   string          `json:"status"`
	Error    string          `json:"error,omitempty"`
	Job      json.RawMessage `json:"job,omitempty"`
	Priority int             `json:"pri,omitempty"`
	Queue    string          `json:"queue,omitempty"`
}

// NewSession creates a new session
func NewSession(ctx context.Context, rw io.ReadWriter) *Session {
	return &Session{
		rw:      rw,
		working: make(map[uint64]*pqueue.Job),
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
			log.Printf("%p request: %s", s.rw, scanner.Text())
		}

		var req request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			sendError(s.rw, err)
			continue
		}

		if req.Request == "" {
			sendError(s.rw, errors.New("invalid request - request field missing"))
			continue
		}

		switch req.Request {
		case "put":
			if req.Job == nil {
				sendError(s.rw, errors.New("invalid put request - job field missing"))
				continue
			}

			if req.Priority == nil {
				sendError(s.rw, errors.New("invalid put request - pri field missing"))
				continue
			}

			if req.Queue == "" {
				sendError(s.rw, errors.New("invalid put request - queue field missing"))
				continue
			}

			job := &pqueue.Job{
				ID:       atomic.AddUint64(&idCounter, 1),
				Priority: *req.Priority,
				Queue:    req.Queue,
				Body:     req.Job,
			}

			if !s.notify(job) {
				store.Enque(job)
			}

			send(s.rw, &response{Status: "ok", ID: job.ID})

		case "get":
			if len(req.Queues) == 0 {
				sendError(s.rw, errors.New("invalid get request - queues field missing"))
				continue
			}

			job := store.Deque(req.Queues)
			if job == nil {
				if req.Wait {
					s.subscribe(req.Queues)
				} else {
					send(s.rw, &response{Status: "no-job"})
				}
				continue
			}

			startJob(job.ID)
			s.working[job.ID] = job

			send(s.rw, &response{
				Status:   "ok",
				ID:       job.ID,
				Priority: job.Priority,
				Queue:    job.Queue,
				Job:      job.Body,
			})

		case "delete":
			if req.ID == nil {
				sendError(s.rw, errors.New("invalid request - id field missing"))
				continue
			}

			stopJob(*req.ID)

			if _, ok := s.working[*req.ID]; ok {
				delete(s.working, *req.ID)
				send(s.rw, &response{Status: "ok", ID: *req.ID})
				continue // stopped running job - it's not in the queue - we're done here.
			}

			if store.Delete(*req.ID) {
				send(s.rw, &response{Status: "ok", ID: *req.ID})
			} else {
				send(s.rw, &response{Status: "no-job", ID: *req.ID})
			}

		case "abort":
			if req.ID == nil {
				sendError(s.rw, errors.New("invalid request - id field missing"))
				continue
			}

			if ok := s.abortJob(*req.ID); ok {
				send(s.rw, &response{Status: "ok", ID: *req.ID})
			} else {
				send(s.rw, &response{Status: "no-job", ID: *req.ID})
			}
		}
	}
}

func (s *Session) abortJob(id uint64) bool {
	if job, ok := s.working[id]; ok {
		if ok := s.notify(job); !ok {
			store.Enque(job)
		}

		stopJob(id)
		delete(s.working, id)
		return true
	}

	return false
}

// subscribe socket for a new job in the given queue.
func (s *Session) subscribe(queues []string) {
	muw.Lock()
	defer muw.Unlock()

	for _, queue := range queues {
		if _, ok := waiting[queue]; !ok {
			waiting[queue] = []io.Writer{}
		}

		waiting[queue] = append(waiting[queue], s.rw)
	}
}

// notify assigns the job to the next waiting client
func (s *Session) notify(job *pqueue.Job) bool {
	muw.Lock()
	defer muw.Unlock()

	que, ok := waiting[job.Queue]
	if !ok || len(que) == 0 {
		return false
	}

	var w io.Writer
	w, waiting[job.Queue] = que[0], que[1:]

	startJob(job.ID)
	s.working[job.ID] = job

	send(w, &response{
		Status:   "ok",
		ID:       job.ID,
		Priority: job.Priority,
		Queue:    job.Queue,
		Job:      job.Body,
	})

	return true
}

// startJob marks a job as in progress by one of the workers.
func startJob(id uint64) {
	mur.Lock()
	defer mur.Unlock()
	running[id] = true
}

// stopJob marks a job as not in progress by one of the workers.
func stopJob(id uint64) {
	mur.Lock()
	defer mur.Unlock()
	delete(running, id)
}

func sendError(w io.Writer, err error) {
	res := response{
		Status: "error",
		Error:  err.Error(),
	}

	jsonb, _ := json.Marshal(res)
	log.Printf("%p response: %s", w, string(jsonb))

	if _, err := fmt.Fprintln(w, string(jsonb)); err != nil {
		log.Printf("%s error: could not send response: %s", w, string(jsonb))
	}
}

func send(w io.Writer, res *response) {
	if debug {
		jsonb, _ := json.Marshal(res)
		log.Printf("%p response: %s", w, string(jsonb))
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		jsonb, _ := json.Marshal(res)
		log.Printf("%s error: could not send response: %s", w, string(jsonb))
	}
}
