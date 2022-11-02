package speed

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/matryer/is"
)

func TestSpeed_logic(t *testing.T) {
	is := is.New(t)
	s := New()

	plate1 := &Plate{Plate: "FOO", Timestamp: 716847}
	plate2 := &Plate{Plate: "FOO", Timestamp: 717147}

	camState1 := &clientState{
		camera: &Camera{
			Road:  42,
			Mile:  1587,
			Limit: 100,
		},
	}

	camState2 := &clientState{
		camera: &Camera{
			Road:  42,
			Mile:  1597,
			Limit: 100,
		},
	}

	buf := &bytes.Buffer{}
	s.dispatchers[42] = buf
	s.limits[42] = 100

	s.registerPlate(plate1, camState1)
	s.registerPlate(plate2, camState2)

	plates, ok := s.plates["FOO"][42]
	is.Equal(ok, true)
	is.Equal(len(plates), 2)

	fmt.Printf("plates records: %v\n", s.plates["FOO"][42])

	s.issueTickets(plate1, camState1)
	fmt.Printf("tickets: %v\n", s.issuedTickets[42][0])

	is.Equal(s.issuedTickets[42][0].Info.Speed, uint16(12000))

	s.dispatchTicket(camState1.camera.Road)

	fmt.Printf("%#v\n", buf.String())

	plate3 := &Plate{Plate: "FOO", Timestamp: 717247}
	camState3 := &clientState{
		camera: &Camera{
			Road:  42,
			Mile:  1607,
			Limit: 100,
		},
	}
	s.registerPlate(plate3, camState3)
	s.issueTickets(plate3, camState3)
}

func TestSpeed_Wierd_Readings(t *testing.T) {
	is := is.New(t)
	s := New()

	plate1 := &Plate{Plate: "FOO", Timestamp: 82724824}
	plate2 := &Plate{Plate: "FOO", Timestamp: 82765354}

	camState1 := &clientState{
		camera: &Camera{
			Road:  42,
			Mile:  677,
			Limit: 60,
		},
	}

	camState2 := &clientState{
		camera: &Camera{
			Road:  42,
			Mile:  10,
			Limit: 60,
		},
	}

	buf := &bytes.Buffer{}
	s.dispatchers[42] = buf
	s.limits[42] = 60

	s.registerPlate(plate1, camState1)
	s.registerPlate(plate2, camState2)

	plates, ok := s.plates["FOO"][42]
	is.Equal(ok, true)
	is.Equal(len(plates), 2)

	fmt.Printf("plates records: %v\n", s.plates["FOO"][42])

	s.issueTickets(plate1, camState1)
	// fmt.Printf("tickets: %v\n", s.issuedTickets[42][0])
	// is.Equal(s.issuedTickets[42][0].Info.Speed, uint16(5924))

	s.dispatchTicket(camState1.camera.Road)
}
