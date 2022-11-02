package speed

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
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

type Reading struct {
	Mile      uint16
	Timestamp uint32
}

type Readings []Reading

// Speed is a speed camera managing solution
type Speed struct {
	mu         sync.Mutex
	limits     map[uint16]uint16              // speed limits per road
	plates     map[string]map[uint16]Readings // plate,road -> mile, time record.
	ticketedOn map[string]map[uint32]struct{}

	dispatchers   map[uint16]io.ReadWriter
	issuedTickets map[uint16][]*Ticket // tickets per road
}

type clientState struct {
	haInterval time.Duration
	camera     *Camera
	dispatcher *Dispatcher
}

// New creates a new Speed instance
func New() *Speed {
	return &Speed{
		limits:        make(map[uint16]uint16),
		plates:        make(map[string]map[uint16]Readings),
		ticketedOn:    make(map[string]map[uint32]struct{}),
		dispatchers:   make(map[uint16]io.ReadWriter),
		issuedTickets: make(map[uint16][]*Ticket),
	}
}

// Handle handles the tcp connection
func (s *Speed) Handle(ctx context.Context, rw io.ReadWriter) {
	state := &clientState{}

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
	log.Println("handling plate message")
	if state.camera == nil {
		return errors.New("The client has not identified itself as camera yet")
	}

	plate := &Plate{}
	if _, err := plate.ReadFrom(rw); err != nil {
		return fmt.Errorf("Failed to parse plate message: %w", err)
	}

	s.registerPlate(plate, state)
	s.issueTickets(plate, state)
	s.dispatchTicket(state.camera.Road)

	return nil
}

func (s *Speed) handleHeartbeat(ctx context.Context, rw io.ReadWriter, state *clientState) error {
	if state.haInterval != 0 {
		return errors.New("heartbeat has already been activated for the client.")
	}

	log.Println("handling wantHeartBeat message")
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
	log.Println("handling IAMCamera message")
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
	log.Println("handling IAMDispatcher message")

	if state.camera != nil {
		return errors.New("The client has already been identified as camera")
	}

	if state.dispatcher != nil {
		return errors.New("The client has already been identified as a dispatcher")
	}

	dispatcher := &Dispatcher{}
	if _, err := dispatcher.ReadFrom(rw); err != nil {
		return fmt.Errorf("Failed to parse IAMDispatcher message")
	}

	log.Printf("New dispatcher: %v", dispatcher)

	for _, road := range dispatcher.Roads {
		// if _, ok := s.dispatchers[road]; !ok {
		// 	// Add dispatchers - even though there can be many dispatchers responsible for the
		// 	// same road, given a ticket cannot be issued more then once - meaning we can only
		// 	// keep the most recent dispatcher socket per road.
		// 	s.dispatchers[road] = rw
		// }
		s.dispatchers[road] = rw
		s.dispatchTicket(road)
	}

	return nil
}

// ================================================================================
func (s *Speed) registerPlate(plate *Plate, state *clientState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("registerPlate start: %s %v", plate.Plate, state)
	defer log.Printf("registerPlate stop: %s %v", plate.Plate, state)

	roads, ok := s.plates[plate.Plate]
	if !ok {
		roads = make(map[uint16]Readings)
		s.plates[plate.Plate] = roads
	}

	records, ok := roads[state.camera.Road]
	if !ok {
		records = make([]Reading, 0)
	}

	roads[state.camera.Road] = append(records, Reading{
		Mile:      state.camera.Mile,
		Timestamp: plate.Timestamp,
	})

	log.Printf("Registering plate %s for road %d: [mile: %d, time: %d]",
		plate.Plate, state.camera.Road, state.camera.Mile, plate.Timestamp)
}

func (s *Speed) issueTickets(plate *Plate, state *clientState) {
	s.mu.Lock()

	log.Printf("issueTickets start: %s", plate.Plate)
	defer log.Printf("issueTickets stop: %s", plate.Plate)

	roads, ok := s.plates[plate.Plate]
	if !ok {
		roads = make(map[uint16]Readings)
		s.plates[plate.Plate] = roads
	}

	records, ok := roads[state.camera.Road]
	if !ok {
		records = make([]Reading, 0)
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
		r1 := records[i-1]
		r2 := records[i]
		log.Printf("r1: %v, r2: %v", r1, r2)

		distance := math.Abs(float64(r2.Mile) - float64(r1.Mile))
		delta := float64(r2.Timestamp - r1.Timestamp)

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
			s.createTicket(r1, r2, uint16(speed*100), plate.Plate, state)
		}
	}
}

func (s *Speed) createTicket(r1, r2 Reading, speed uint16, plate string, state *clientState) {
	ticketDates, ok := s.ticketedOn[plate]
	if !ok {
		ticketDates = make(map[uint32]struct{})
		s.mu.Lock()
		s.ticketedOn[plate] = ticketDates
		s.mu.Unlock()
	}

	day1 := r1.Timestamp / 86400
	day2 := r2.Timestamp / 86400

	for i := day1; i <= day2; i++ {
		if _, ok := ticketDates[i]; ok {
			log.Printf("%s has already been ticketed on the %d day", plate, i)
			return // already ticketed
		}
	}

	// register ticket
	for i := day1; i <= day2; i++ {
		ticketDates[i] = struct{}{}
	}

	log.Printf("createTicket start: %s -> %d: %v %v", plate, speed, r1, r2)
	defer log.Printf("createTicket start: %s -> %d: %v %v", plate, speed, r1, r2)

	ticket := &Ticket{
		Plate: plate,
		Info: TicketInfo{
			Road:       state.camera.Road,
			Mile1:      r1.Mile,
			TimeStamp1: r1.Timestamp,
			Mile2:      r2.Mile,
			TimeStamp2: r2.Timestamp,
			Speed:      speed,
		},
	}

	log.Printf("dispatching ticket: %v", ticket)
	s.mu.Lock()
	// append tickets to the pending tickets
	tickets, ok := s.issuedTickets[state.camera.Road]
	if !ok {
		tickets = make([]*Ticket, 0)
	}
	s.mu.Unlock()

	log.Printf("add pending ticket: %v", ticket)
	s.issuedTickets[state.camera.Road] = append(tickets, ticket)
}

func (s *Speed) dispatchTicket(road uint16) {
	log.Printf("dispatchTicket start: %d", road)
	defer log.Printf("dispatchTicket stop: %d", road)

	ww, ok := s.dispatchers[road]
	if !ok {
		log.Println("no dispatchers found so far")
		return // no dispatchers
	}

	tickets, ok := s.issuedTickets[road]
	if !ok {
		log.Println("no tickets found so far")
		return // no tickets
	}

	for _, ticket := range tickets {
		log.Printf("sending ticket %v to dispatcher", ticket)

		if _, err := ticket.WriteTo(ww); err != nil {
			log.Printf("error could not send ticket: %s", err.Error())
		}
	}

	s.mu.Lock()
	s.issuedTickets[road] = tickets[:0]
	s.mu.Unlock()
}
