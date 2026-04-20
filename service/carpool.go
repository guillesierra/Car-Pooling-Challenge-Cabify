package service

import (
	"errors"
	"sort"

	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/service/model"
)

const MaxSeats = 6

var (
	ErrNotFound      = errors.New("not found")
	ErrDuplicatedID  = errors.New("duplicated ID")
	ErrInvalidCars   = errors.New("invalid cars payload")
	ErrInvalidJourney = errors.New("invalid journey payload")
)

type JourneyStatus int

const (
	statusWaiting JourneyStatus = iota
	statusRiding
)

type journeyEntry struct {
	journey *model.Journey // Journey payload.
	status  JourneyStatus  // Waiting or riding.
}

type CarPool struct {
	cars     map[uint]*model.Car      // Fleet by car ID.
	journeys map[uint]*journeyEntry   // Journeys by journey ID.
	pending  []uint                   // FIFO waiting queue.
}

// New_CarPool creates an empty pool service.
func New_CarPool() *CarPool {
	return &CarPool{
		cars:     make(map[uint]*model.Car),
		journeys: make(map[uint]*journeyEntry),
		pending:  make([]uint, 0),
	}
}

// ResetCars loads cars and clears previous state.
func (cp *CarPool) ResetCars(cars []*model.Car) error {
	nextCars := make(map[uint]*model.Car, len(cars))
	for _, car := range cars {
		if car == nil || car.ID == 0 {
			return ErrInvalidCars
		}
		if car.Seats < 4 || car.Seats > MaxSeats {
			return ErrInvalidCars
		}
		if _, exists := nextCars[car.ID]; exists {
			return ErrDuplicatedID
		}

		copied := &model.Car{
			ID:             car.ID,
			Seats:          car.Seats,
			AvailableSeats: car.Seats,
		}
		nextCars[copied.ID] = copied
	}

	cp.cars = nextCars
	cp.journeys = make(map[uint]*journeyEntry)
	cp.pending = cp.pending[:0]
	return nil
}

// NewJourney registers and tries to assign a journey.
func (cp *CarPool) NewJourney(journey *model.Journey) (assigned bool, err error) {
	if journey == nil || journey.ID == 0 {
		return false, ErrInvalidJourney
	}
	if journey.People == 0 || journey.People > MaxSeats {
		return false, ErrInvalidJourney
	}
	if _, exists := cp.journeys[journey.ID]; exists {
		return false, ErrDuplicatedID
	}

	copied := &model.Journey{
		ID:     journey.ID,
		People: journey.People,
	}
	entry := &journeyEntry{
		journey: copied,
		status:  statusWaiting,
	}
	cp.journeys[copied.ID] = entry
	cp.pending = append(cp.pending, copied.ID)

	cp.reassignPending()
	return entry.status == statusRiding, nil
}

// Dropoff removes a journey and frees occupied seats.
func (cp *CarPool) Dropoff(journeyID uint) error {
	entry, exists := cp.journeys[journeyID]
	if !exists {
		return ErrNotFound
	}

	if entry.status == statusWaiting {
		cp.removePending(journeyID)
		delete(cp.journeys, journeyID)
		return nil
	}

	car := entry.journey.AssignedTo
	if car != nil {
		car.AvailableSeats += entry.journey.People
		if car.AvailableSeats > car.Seats {
			car.AvailableSeats = car.Seats
		}
	}

	delete(cp.journeys, journeyID)
	cp.reassignPending()
	return nil
}

// Locate returns whether a journey is waiting or riding.
func (cp *CarPool) Locate(journeyID uint) (car *model.Car, waiting bool, err error) {
	entry, exists := cp.journeys[journeyID]
	if !exists {
		return nil, false, ErrNotFound
	}
	if entry.status == statusWaiting {
		return nil, true, nil
	}
	return entry.journey.AssignedTo, false, nil
}

// reassignPending keeps assigning while possible.
func (cp *CarPool) reassignPending() {
	for {
		assignedID, car := cp.nextAssignment()
		if assignedID == 0 || car == nil {
			return
		}

		entry := cp.journeys[assignedID]
		entry.journey.AssignedTo = car
		entry.status = statusRiding
		car.AvailableSeats -= entry.journey.People
		cp.removePending(assignedID)
	}
}

// nextAssignment applies FIFO when possible.
func (cp *CarPool) nextAssignment() (journeyID uint, car *model.Car) {
	for _, id := range cp.pending {
		entry, exists := cp.journeys[id]
		if !exists || entry.status != statusWaiting {
			continue
		}

		selected := cp.bestFitCar(entry.journey.People)
		if selected == nil {
			continue
		}

		return id, selected
	}
	return 0, nil
}

// bestFitCar picks the tightest fitting available car.
func (cp *CarPool) bestFitCar(people uint) *model.Car {
	cars := make([]*model.Car, 0, len(cp.cars))
	for _, car := range cp.cars {
		if car.AvailableSeats >= people {
			cars = append(cars, car)
		}
	}
	if len(cars) == 0 {
		return nil
	}

	sort.Slice(cars, func(i, j int) bool {
		left, right := cars[i], cars[j]
		if left.AvailableSeats == right.AvailableSeats {
			return left.ID < right.ID
		}
		return left.AvailableSeats < right.AvailableSeats
	})

	return cars[0]
}

// removePending deletes a journey from waiting queue.
func (cp *CarPool) removePending(journeyID uint) {
	for i, id := range cp.pending {
		if id != journeyID {
			continue
		}
		cp.pending = append(cp.pending[:i], cp.pending[i+1:]...)
		return
	}
}
