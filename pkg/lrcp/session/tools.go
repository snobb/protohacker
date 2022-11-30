package session

import (
	"fmt"
	"strconv"
)

// Session constants
const (
	MinSessionID = 0
	MaxSessionID = 2147483648
)

// Escape escapes the special character in the given string
func Escape(buf []byte) []byte {
	// naive and inefficient implementation.
	// buf = bytes.ReplaceAll(buf, []byte{'\\/'}, []byte{'\\', '\\'})
	// return bytes.ReplaceAll(buf, []byte{'/'}, []byte{'\\', '/'})

	// potentially more efficient solution than above
	for i := 0; i < len(buf); i++ {
		if buf[i] == '\\' || buf[i] == '/' {
			buf = append(buf[:i+1], buf[i:]...)
			buf[i] = '\\'
			i++
		}
	}

	return buf
}

// Unescape escapes the special character in the given string
func Unescape(buf []byte) []byte {
	lo, hi := 0, 0
	for ; hi < len(buf); hi++ {
		if buf[hi] == '\\' && (buf[hi+1] == '/' || buf[hi+1] == '\\') {
			hi++
		}

		buf[lo] = buf[hi]
		lo++
	}

	if hi != lo {
		return buf[:lo]
	}

	return buf
}

// ParseID parses the session ID
func ParseID(buf []byte) (int, error) {
	sid, err := strconv.Atoi(string(buf))
	if err != nil {
		return -1, fmt.Errorf("cannot parse session ID: %w", err)
	}

	if int64(sid) < MinSessionID || int64(sid) > MaxSessionID {
		return -1, fmt.Errorf("the session ID is out of range: %s", string(buf))
	}

	return sid, nil
}
