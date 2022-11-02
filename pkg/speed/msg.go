package speed

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
)

// ================================================================================
type Plate struct {
	Plate     string
	Timestamp uint32
}

// ReadFrom implements io.ReaderFrom interface
func (p *Plate) ReadFrom(r io.Reader) (int64, error) {
	var err error

	p.Plate, err = ReadString(r)
	if err != nil {
		return -1, err
	}

	if err := binary.Read(r, binary.BigEndian, &p.Timestamp); err != nil {
		return -1, err
	}

	return int64(len(p.Plate)) + 4, nil
}

// ================================================================================
type Camera struct {
	Road  uint16
	Mile  uint16
	Limit uint16
}

// ReadFrom implements io.ReaderFrom interface
func (imc *Camera) ReadFrom(r io.Reader) (int64, error) {
	if err := binary.Read(r, binary.BigEndian, imc); err != nil {
		return -1, err
	}

	return 6, nil
}

// ================================================================================
type Dispatcher struct {
	NumRoads uint8
	Roads    []uint16
}

// ReadFrom implements io.ReaderFrom interface
func (d *Dispatcher) ReadFrom(r io.Reader) (int64, error) {
	var err error
	var byteBuf [1]byte

	if _, err = r.Read(byteBuf[:]); err != nil {
		return -1, err
	}
	d.NumRoads = byteBuf[0]

	for i := 0; i < int(d.NumRoads); i++ {
		var value uint16
		if err = binary.Read(r, binary.BigEndian, &value); err != nil {
			return -1, err
		}

		d.Roads = append(d.Roads, value)
	}

	return int64(d.NumRoads) * 2, nil
}

// ================================================================================
type TicketInfo struct {
	Road       uint16
	Mile1      uint16
	TimeStamp1 uint32
	Mile2      uint16
	TimeStamp2 uint32
	Speed      uint16
}

type Ticket struct {
	Plate string
	Info  TicketInfo
}

// WriteTo implements io.WriterTo interface
func (t *Ticket) WriteTo(w io.Writer) (int64, error) {
	if _, err := w.Write([]byte{typeTicket}); err != nil {
		return -1, err
	}

	WriteString(w, t.Plate)

	if err := binary.Write(w, binary.BigEndian, &t.Info); err != nil {
		return -1, err
	}

	// type + string, length + info (uint16 * 4) + (uint32 * 2)
	return 1 + int64(len(t.Plate)) + 1 + 16, nil
}

// === Generic functions ==========================================================
func writeError(w io.Writer, err error) {
	log.Print(err.Error())

	if _, err := w.Write([]byte{typeError}); err != nil {
		log.Printf("failed to write error msg: %s", err.Error())
	}

	if err := WriteString(w, err.Error()); err != nil {
		log.Printf("failed to write error msg: %s", err.Error())
	}
}

func ReadString(r io.Reader) (string, error) {
	var sz [1]byte
	_, err := r.Read(sz[:])
	if errors.Is(err, io.EOF) {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("failed to read string size: %w", err)
	}

	buf := make([]byte, sz[0])
	_, err = io.ReadFull(r, buf)

	return string(buf), err
}

func WriteString(w io.Writer, str string) error {
	if len(str) > 255 {
		return fmt.Errorf("string is too long (%d): %s", len(str), str)
	}

	buf := append([]byte{byte(len(str))}, []byte(str)...)
	if _, err := w.Write(buf); err != nil {
		return err
	}

	return nil
}
