package tcpserver

import (
	"bytes"
	"io"
	"net"
	"time"
)

// Recorder is a recording buffer for mock connection.
type Recorder struct {
	In  io.ReadWriter
	Out bytes.Buffer
}

// Read implements io.Reader for Recorder
func (r *Recorder) Read(p []byte) (n int, err error) {
	return r.In.Read(p)
}

// Write implements io.Writer for Recorder
func (r *Recorder) Write(p []byte) (n int, err error) {
	return r.Out.Write(p)
}

// TestConn is a mock connection for tests.
type TestConn struct {
	Recorder
}

// Read reads data from the connection.
// Read can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetReadDeadline.
func (t *TestConn) Read(b []byte) (n int, err error) {
	return t.Recorder.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetWriteDeadline.
func (t *TestConn) Write(b []byte) (n int, err error) {
	return t.Recorder.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (t *TestConn) Close() error {
	panic("not implemented") // TODO: Implement
}

// LocalAddr returns the local network address, if known.
func (t *TestConn) LocalAddr() net.Addr {
	panic("not implemented") // TODO: Implement
}

// RemoteAddr returns the remote network address, if known.
func (t *TestConn) RemoteAddr() net.Addr {
	panic("not implemented") // TODO: Implement
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail instead of blocking. The deadline applies to all future
// and pending I/O, not just the immediately following call to
// Read or Write. After a deadline has been exceeded, the
// connection can be refreshed by setting a deadline in the future.
//
// If the deadline is exceeded a call to Read or Write or to other
// I/O methods will return an error that wraps os.ErrDeadlineExceeded.
// This can be tested using errors.Is(err, os.ErrDeadlineExceeded).
// The error's Timeout method will return true, but note that there
// are other possible errors for which the Timeout method will
// return true even if the deadline has not been exceeded.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (t *TestConn) SetDeadline(tm time.Time) error {
	panic("not implemented") // TODO: Implement
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (t *TestConn) SetReadDeadline(tm time.Time) error {
	panic("not implemented") // TODO: Implement
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (t *TestConn) SetWriteDeadline(tm time.Time) error {
	panic("not implemented") // TODO: Implement
}
