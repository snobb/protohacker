package prime

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
)

const errorMethod = "error"

type Request struct {
	Method string   `json:"method"`
	Number *float64 `json:"number"`
}

type Response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func getLine(ctx context.Context, r io.Reader) <-chan []byte {
	ch := make(chan []byte)

	scanner := bufio.NewScanner(r)
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

			ch <- bytes
		}
	}()

	return ch
}

func getResponse(req Request) Response {
	if !validate(req) {
		return Response{Method: errorMethod}
	}

	resp := Response{Method: "isPrime"}
	num := *req.Number

	// check if the number is integer
	if num != float64(int64(num)) {
		resp.Prime = false
	}

	resp.Prime = isPrime(int64(num))

	return resp
}

// Handle handles a new tcp connection
func Handle(ctx context.Context, conn net.Conn) {
	enc := json.NewEncoder(conn)

	for line := range getLine(ctx, conn) {
		select {
		case <-ctx.Done():
			log.Println("PrimeHandle: got canceled")
			return

		default:
		}

		log.Println("handling input:", string(line))

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			log.Printf("Invalid request - %s\n", err.Error())
			req.Method = errorMethod
		}

		resp := getResponse(req)

		if err := enc.Encode(resp); err != nil {
			err = fmt.Errorf("failed to encode response - %w: %v", err, resp)
			_ = enc.Encode(Response{Method: errorMethod})
			log.Printf("Error encoding: %s", err.Error())
			return
		}

		log.Println("Sent response:", resp)
	}
}

// validate the request
// A request is malformed if it is not a well-formed JSON object, if any required field is
// missing, if the method name is not "isPrime", or if the number value is not a number.
func validate(req Request) bool {
	return req.Method == "isPrime" && req.Number != nil
}

func isPrime(num int64) bool {
	if num <= 1 {
		return false
	}

	sqRoot := int64(math.Sqrt(float64(num)))
	for i := int64(2); i <= sqRoot; i++ {
		if num%i == 0 {
			return false
		}
	}

	return true
}
