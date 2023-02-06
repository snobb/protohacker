package speed

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"sort"
	"sync"
	"time"
)

const (
	typeError         uint8 = 0x10
	typePlate         uint8 = 0x20
	typeTicket        uint8 = 0x21
	typeWantHeartbeat uint8 = 0x40
	typeHeartbeat     uint8 = 0x41
	typeIAMCamera     uint8 = 0x80
	typeIAMDispatcher uint8 = 0x81
)

type readings []PlateReading

// Speed is a speed camera managing solution
type Speed struct {
	mu         sync.Mutex
	limits     map[uint16]uint16              // speed limits per road
	plates     map[string]map[uint16]readings // plate,road -> mile, time record.
	ticketDays map[string]map[uint32]struct{}

	issuedTicketsCh map[uint16]chan *Ticket // tickets per road
	trackTicketsCh  chan *Ticket            // tickets to create
}

type clientState struct {
	addr       net.Addr
	haInterval time.Duration
	camera     *Camera
	dispatcher *Dispatcher
}

// New creates a new Speed instance
func New(ctx context.Context) *Speed {
	speed := &Speed{
		limits:          make(map[uint16]uint16),
		plates:          make(map[string]map[uint16]readings),
		ticketDays:      make(map[string]map[uint32]struct{}),
		issuedTicketsCh: make(map[uint16]chan *Ticket),
		trackTicketsCh:  make(chan *Ticket),
	}

	go func(ctx context.Context) {
		for ticket := range speed.trackTicketsCh {
			speed.trackTicket(ticket)
		}
	}(ctx)

	return speed
}

// Handle handles the tcp connection
func (s *Speed) Handle(ctx context.Context, rw io.ReadWriter, addr net.Addr) {
	state := &clientState{addr: addr}

	for {
		var kind [1]byte
		_, err := rw.Read(kind[:])
		if errors.Is(err, io.EOF) {
			return
		} else if err != nil {
			log.Printf("failed to read msg kind: %s", err.Error())
			writeError(rw, fmt.Errorf("Failed to read msg kind: %w", err))
			return
		}

		switch kind[0] {
		case typePlate:
			err = s.handlePlate(ctx, rw, state)

		case typeWantHeartbeat:
			err = s.handleHeartbeat(ctx, rw, state)

		case typeIAMCamera:
			err = s.handleCamera(ctx, rw, state)

		case typeIAMDispatcher:
			err = s.handleDispatcher(ctx, rw, state)

		default:
			writeError(rw, errors.New("Unexpected message type"))
		}

		if err != nil {
			writeError(rw, err)
		}
	}
}

// ==== Message handlers ==========================================================

func (s *Speed) handlePlate(ctx context.Context, rw io.ReadWriter, state *clientState) error {
	log.Printf("handling plate message from %s", state.addr.String())
	if state.camera == nil {
		return errors.New("The client has not identified itself as camera yet")
	}

	plate := &Plate{}
	if _, err := plate.ReadFrom(rw); err != nil {
		return fmt.Errorf("Failed to parse plate message: %w", err)
	}

	s.registerPlate(plate, state)
	s.issueTickets(plate, state)

	return nil
}

func (s *Speed) handleHeartbeat(ctx context.Context, rw io.ReadWriter, state *clientState) error {
	if state.haInterval != 0 {
		return errors.New("heartbeat has already been activated for the client.")
	}

	log.Printf("handling wantHeartBeat message from %s", state.addr.String())
	var interval uint32
	if err := binary.Read(rw, binary.BigEndian, &interval); err != nil {
		return fmt.Errorf("Failed to parse WantHeartBeat message")
	}

	if interval == 0 {
		return nil // no heartbeat
	}

	state.haInterval = time.Duration(interval) * 100 * time.Millisecond
	log.Printf("starting heartbeat with interval %s\n", state.haInterval)

	go func() {
		// interval in deciseconds. eg. 25 means 2.5seconds.
		// Converting to milliseconds for convenience by multiplying by 100.
		tick := time.NewTicker(state.haInterval)
		defer tick.Stop()

		for {
			select {

			case <-ctx.Done():
				log.Print("heartbeat cancelled")
				return

			case <-tick.C:
				_, _ = rw.Write([]byte{typeHeartbeat})
			}
		}
	}()

	return nil
}

func (s *Speed) handleCamera(ctx context.Context, rw io.ReadWriter, state *clientState) error {
	log.Printf("handling IAMCamera message from %s", state.addr.String())
	if state.camera != nil {
		return errors.New("The camera has already been identified")
	}

	state.camera = &Camera{}
	if _, err := state.camera.ReadFrom(rw); err != nil {
		return fmt.Errorf("Failed to parse IAMCamera message: %w", err)
	}

	log.Printf("registering camera at road %d [mile: %d, limit: %d]",
		state.camera.Road, state.camera.Mile, state.camera.Limit)

	s.limits[state.camera.Road] = state.camera.Limit

	return nil
}

func (s *Speed) handleDispatcher(ctx context.Context, rw io.ReadWriter, state *clientState) error {
	log.Printf("handling IAMDispatcher message from %s", state.addr.String())

	if state.camera != nil {
		return errors.New("The client has already been identified as camera")
	}

	if state.dispatcher != nil {
		return errors.New("The client has already been identified as a dispatcher")
	}

	state.dispatcher = &Dispatcher{}
	if _, err := state.dispatcher.ReadFrom(rw); err != nil {
		return fmt.Errorf("Failed to parse IAMDispatcher message")
	}

	log.Printf("New dispatcher: %v", state.dispatcher)

	for _, road := range state.dispatcher.Roads {
		go s.subscribeForRoad(ctx, rw, road)
	}

	return nil
}

// ================================================================================

func (s *Speed) subscribeForRoad(ctx context.Context, w io.Writer, road uint16) {
	var ticket *Ticket

	ch := s.issuedTicketsChannel(road)
	for {
		select {
		case <-ctx.Done():
			return

		case ticket = <-ch:
			log.Printf("Dispatching ticket %v for road: %d", ticket, road)

			if _, err := ticket.WriteTo(w); err != nil {
				log.Printf("error could not send ticket: %s", err.Error())
			}
		}
	}
}

func (s *Speed) registerPlate(plate *Plate, state *clientState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	roads, ok := s.plates[plate.Plate]
	if !ok {
		roads = make(map[uint16]readings)
		s.plates[plate.Plate] = roads
	}

	records, ok := roads[state.camera.Road]
	if !ok {
		records = make([]PlateReading, 0)
	}

	roads[state.camera.Road] = append(records, PlateReading{
		Mile:      state.camera.Mile,
		Timestamp: plate.Timestamp,
	})

	log.Printf("Registering plate %s for road %d: [mile: %d, time: %d]",
		plate.Plate, state.camera.Road, state.camera.Mile, plate.Timestamp)
}

func (s *Speed) issueTickets(plate *Plate, state *clientState) {
	s.mu.Lock()

	roads, ok := s.plates[plate.Plate]
	if !ok {
		roads = make(map[uint16]readings)
		s.plates[plate.Plate] = roads
	}

	records, ok := roads[state.camera.Road]
	if !ok {
		records = make([]PlateReading, 0)
	} else {
		// sort records
		sort.Slice(records, func(i, j int) bool {
			return records[i].Timestamp < records[j].Timestamp
		})
	}
	s.mu.Unlock()

	limit, ok := s.limits[state.camera.Road]
	if !ok {
		// paranoia - no limits for the road yet
		return
	}

	for i := 1; i < len(records); i++ {
		r1, r2 := records[i-1], records[i]
		log.Printf("r1: %v, r2: %v", r1, r2)

		distance := math.Abs(float64(r2.Mile) - float64(r1.Mile))
		delta := float64(r2.Timestamp) - float64(r1.Timestamp)

		if delta == 0 {
			continue
		}
		log.Printf("distance: %v, delta: %v", distance, delta)

		speed := distance / delta * 3600
		log.Printf("calculated speed: %d", uint16(speed))
		log.Printf("limit: %d", limit)

		// the error is 0.5 but to avoid corner cases we can half the error since it's acceptable
		// by the spec.
		if speed > float64(limit)+0.3 {
			log.Printf("speeding detected: %s - speed: %d", plate.Plate, uint16(speed))
			s.trackTicket(&Ticket{
				Plate: plate.Plate,
				Info: TicketInfo{
					Reading1: r1,
					Reading2: r2,
					Road:     state.camera.Road,
					Speed:    uint16(speed * 100),
				},
			})
		}
	}
}

func (s *Speed) trackTicket(ticket *Ticket) {
	ticketDates, ok := s.ticketDays[ticket.Plate]
	if !ok {
		ticketDates = make(map[uint32]struct{})
		s.ticketDays[ticket.Plate] = ticketDates
	}

	day1 := ticket.Info.Reading1.Timestamp / 86400
	day2 := ticket.Info.Reading2.Timestamp / 86400

	for i := day1; i <= day2; i++ {
		if _, ok := ticketDates[i]; ok {
			log.Printf("%s has already been ticketed on the %d day", ticket.Plate, i)
			return // already ticketed
		}
	}

	// register ticket
	for i := day1; i <= day2; i++ {
		ticketDates[i] = struct{}{}
	}

	log.Printf("add pending ticket: %v", ticket)
	ch := s.issuedTicketsChannel(ticket.Info.Road)
	ch <- ticket
}

func (s *Speed) issuedTicketsChannel(road uint16) chan *Ticket {
	s.mu.Lock()
	defer s.mu.Unlock()

	// append tickets to the pending tickets
	ch, ok := s.issuedTicketsCh[road]
	if !ok {
		ch = make(chan *Ticket, 1000)
		s.issuedTicketsCh[road] = ch
	}

	return ch
}
