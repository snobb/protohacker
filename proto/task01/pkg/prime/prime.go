package prime

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
)

const errorMethod = "error"

type request struct {
	Method string   `json:"method"`
	Number *float64 `json:"number"`
}

type response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func getResponse(req request) response {
	if !validate(req) {
		return response{Method: errorMethod}
	}

	num := *req.Number

	// check if the number is integer
	if num != float64(int64(num)) {
		return response{Method: "isPrime", Prime: false}
	}

	return response{Method: "isPrime", Prime: isPrime(int64(num))}
}

// Handle handles a new tcp connection
func Handle(ctx context.Context, conn net.Conn) {
	enc := json.NewEncoder(conn)
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			log.Println("PrimeHandle: got canceled")
			return

		default:
		}

		log.Println("handling input:", scanner.Text())

		var req request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			log.Printf("Invalid request - %s\n", err.Error())
			req.Method = errorMethod
		}

		resp := getResponse(req)

		if err := enc.Encode(resp); err != nil {
			err = fmt.Errorf("failed to encode response - %w: %v", err, resp)
			_ = enc.Encode(response{Method: errorMethod})
			log.Printf("Error encoding: %s", err.Error())
			return
		}

		log.Println("Sent response:", resp)
	}

	if err := scanner.Err(); err != nil {
		_ = enc.Encode(response{Method: errorMethod})
		log.Printf("Scanner: error: %s", err.Error())
	}
}

// validate the request
// A request is malformed if it is not a well-formed JSON object, if any required field is
// missing, if the method name is not "isPrime", or if the number value is not a number.
func validate(req request) bool {
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
