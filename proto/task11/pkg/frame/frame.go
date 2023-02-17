package frame

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"
)

// Kinds of message
const (
	KindHello            uint8 = 0x50
	KindError            uint8 = 0x51
	KindOK               uint8 = 0x52
	KindDialAuthority    uint8 = 0x53
	KindTargetPopulation uint8 = 0x54
	KindCreatePolicy     uint8 = 0x55
	KindDeletePolicy     uint8 = 0x56
	KindPolicyResult     uint8 = 0x57
	KindSiteVisit        uint8 = 0x58
)

// Kind to string mapping
var Kind2Name = map[uint8]string{
	KindHello:            "Hello",
	KindError:            "Error",
	KindOK:               "Ok",
	KindDialAuthority:    "DialAuthority",
	KindTargetPopulation: "TargetPopulation",
	KindCreatePolicy:     "CreatePolicy",
	KindDeletePolicy:     "DeletePolicy",
	KindPolicyResult:     "PolicyResult",
	KindSiteVisit:        "SiteVisit",
}

const unreasonablyLarge = 1024 * 1024

// Reader is an interface of a message that can be read
type Reader interface {
	Kind() uint8
	Read([]byte) error
}

// Writer is an interface of a message that can be written
type Writer interface {
	Kind() uint8
	Write() ([]byte, error)
}

// Frame is an abstraction of a message underlying layer
type Frame struct {
	Kind uint8

	// Not including kind, size and checksum
	Payload []byte
}

// New creates a new Frame of given kind and with given payload
func New(kind uint8, payload []byte) *Frame {
	return &Frame{
		Kind:    kind,
		Payload: payload,
	}
}

// ReadFrom implements io.ReaderFrom interface
func (f *Frame) ReadFrom(r io.Reader) (n int64, err error) {
	if f.Kind, err = ReadU8(r); err != nil {
		return -1, err
	}

	// if frame is an error - return the error right away.
	if f.Kind == KindError {
		var pcErr Error
		if err := pcErr.Read(f.Payload); err != nil {
			return -1, err
		}

		return -1, pcErr
	}

	sz, err := ReadU32(r)
	if err != nil {
		return -1, err
	}

	if sz > unreasonablyLarge {
		return -1, fmt.Errorf("Frame.ReadFrom: size is unreasonably large")
	}

	// do not read the kind, size and checksum (kind 1 + size 4 + chksum 1 == 6)
	if f.Payload, err = ReadBytes(r, sz-6); err != nil {
		return -1, err
	}

	if len(f.Payload) != int(sz-6) {
		return -1, fmt.Errorf("Frame.ReadFrom: incomplete payload")
	}

	chksum, err := ReadU8(r)
	if err != nil {
		return -1, err
	}

	if err := f.validateChecksum(chksum); err != nil {
		return -1, err
	}

	log.Printf("Frame:%s", f)

	return int64(sz), nil
}

// WriteTo implements io.WriterTo interface
func (f *Frame) WriteTo(w io.Writer) (n int64, err error) {
	if err := WriteU8(w, f.Kind); err != nil {
		return -1, err
	}

	// (kind 1 + size 4 + chksum 1 == 6)
	sz := len(f.Payload) + 6

	if err := WriteU32(w, uint32(sz)); err != nil {
		return -1, err
	}

	if err := WriteBytes(w, f.Payload); err != nil {
		return -1, err
	}

	if err := WriteU8(w, f.checksum()); err != nil {
		return -1, err
	}

	return int64(sz), nil
}

// UnloadInto unloads the frames payload into the privided segment (Reader)
func (f *Frame) UnloadInto(msg Reader) error {
	if f.Kind != msg.Kind() {
		return fmt.Errorf("ReadFrame: expected:%s, got:%s",
			Kind2Name[msg.Kind()], Kind2Name[f.Kind])
	}

	log.Printf("Segment: %v", msg)

	return msg.Read(f.Payload)
}

// ReadFrame is a convenience abstraction to read a single frame from io.Reader
func ReadFrame(r io.Reader) (*Frame, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	frm := &Frame{}
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("ReadFrame: timed out")

	default:
		if _, err := frm.ReadFrom(r); err != nil {
			return nil, err
		}
	}

	return frm, nil
}

// WriteFrame is a convenience abstraction to write a single frame to io.Writer
func WriteFrame(w io.Writer, msg Writer) error {
	log.Printf("Send: %v", msg)

	data, err := msg.Write()
	if err != nil {
		return err
	}

	frm := New(msg.Kind(), data)
	if _, err := frm.WriteTo(w); err != nil {
		return err
	}

	return nil
}

// WriteError writes Error to the provided io.Writer
func WriteError(w io.Writer, err error) {
	if !errors.Is(err, io.EOF) {
		WriteFrame(w, NewError(err))
	}
}

// Handshake exchanges Hello messages.
func Handshake(rw io.ReadWriter) error {
	if err := WriteFrame(rw, NewHello()); err != nil {
		return err
	}

	frm, err := ReadFrame(rw)
	if err != nil {
		return err
	}

	return frm.UnloadInto(NewHello())
}

// Len shows the full size of the frame.
func (f *Frame) Len() int {
	return len(f.Payload) + 6
}

// String implements fmt.Stringer for Frame.
func (f *Frame) String() string {
	return fmt.Sprintf("[kind:%s size:%d] %v", Kind2Name[f.Kind], f.Len(), f.Payload)
}

func (f *Frame) validateChecksum(chksum uint8) error {
	sum := int(f.Kind)

	sz := len(f.Payload) + 6
	for i := 0; i <= 24; i += 8 {
		sum += int((sz >> i) & 0xff)
	}

	for _, b := range f.Payload {
		sum += int(b)
	}

	sum += int(chksum)

	if sum%256 != 0 {
		return fmt.Errorf("Invalid checksum: %x", chksum)
	}

	return nil
}

func (f *Frame) checksum() uint8 {
	sum := int(f.Kind)

	sz := len(f.Payload) + 6
	for i := 0; i <= 24; i += 8 {
		sum += int((sz >> i) & 0xff)
	}

	for _, b := range f.Payload {
		sum += int(b)
	}

	return uint8(256 - sum%256)
}
