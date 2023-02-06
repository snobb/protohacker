package speed

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/matryer/is"
)

func TestSpeed_logic(t *testing.T) {
	is := is.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := New(ctx)

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

	s.limits[42] = 100
	var buf bytes.Buffer
	go s.subscribeForRoad(ctx, &buf, 42)
	go s.subscribeForRoad(ctx, &buf, 42)

	s.registerPlate(plate1, camState1)
	s.registerPlate(plate2, camState2)

	plates, ok := s.plates["FOO"][42]
	is.Equal(ok, true)
	is.Equal(len(plates), 2)

	fmt.Printf("plates records: %v\n", s.plates["FOO"][42])

	s.issueTickets(plate1, camState1)

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := New(ctx)

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

	s.limits[42] = 60

	s.registerPlate(plate1, camState1)
	s.registerPlate(plate2, camState2)

	plates, ok := s.plates["FOO"][42]
	is.Equal(ok, true)
	is.Equal(len(plates), 2)

	fmt.Printf("plates records: %v\n", s.plates["FOO"][42])

	s.issueTickets(plate1, camState1)
}
