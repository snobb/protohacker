package prime

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

const errorMethod = "error"

type Request struct {
	Method string  `json:"method"`
	Number float64 `json:"number"`
}

type Response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func getRequests(ctx context.Context, io io.ReadWriteCloser) <-chan Request {
	defer io.Close()
	ch := make(chan Request)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	scanner := bufio.NewScanner(io)
	go func() {
		defer close(ch)

		for idx := 0; scanner.Scan(); idx++ {
			select {
			case <-ctx.Done():
				log.Println("getLines: got canceled")
				return

			default:
			}

			bytes := scanner.Bytes()
			if len(bytes) == 0 {
				break
			}

			log.Println("handling input:", string(bytes))

			var req Request
			if err := json.Unmarshal(bytes, &req); err != nil || !validate(req) {
				log.Printf("Invalid request - %s: %#v\n", err.Error(), req)
				req.Method = errorMethod
				ch <- req
				return
			}

			ch <- req
		}
	}()

	return ch
}

func Handle(ctx context.Context, ioh io.ReadWriteCloser) {
	defer ioh.Close()

	concur := make(chan struct{}, 5)
	defer close(concur)

	var (
		cnt  int64
		wg   sync.WaitGroup
		resp []Response
		enc  = json.NewEncoder(ioh)
	)

	for req := range getRequests(ctx, ioh) {
		if req.Method == errorMethod {
			_ = enc.Encode(req)
			return
		}

		concur <- struct{}{}
		resp = append(resp, Response{Method: "isPrime"})

		log.Println("handling request:", req)

		wg.Add(1)
		go func(id int64, r Request) {
			defer func() {
				wg.Done()
				<-concur
			}()

			select {
			case <-ctx.Done():
				log.Println("PrimeHandle: got canceled")
				return

			default:
			}

			// check if the number is integer
			if r.Number != float64(int64(r.Number)) {
				resp[id].Prime = false
				return
			}

			resp[id].Prime = isPrime(int64(r.Number))
		}(cnt, req)

		atomic.AddInt64(&cnt, 1)
	}

	wg.Wait()

	for _, res := range resp {
		if err := enc.Encode(res); err != nil {
			log.Printf("Failed to encode response - %s: %v\n", err.Error(), res)
			_ = enc.Encode(Response{Method: "error"})
			return
		}

		log.Println("Sent response:", res)
	}
}

// validate the request
// A request is malformed if it is not a well-formed JSON object, if any
// required field is missing, if the method name is not "isPrime", or if the
// number value is not a number.
func validate(req Request) bool {
	return req.Method == "isPrime" // Unmarshal will check the number
}

func isPrime(num int64) bool {
	if num < 2 {
		return false
	}

	sqRoot := int64(math.Sqrt(float64(num)))
	for i := int64(2); i < sqRoot; i++ {
		if num%i == 0 {
			return false
		}
	}

	return true
}
