package service

import (
	"errors"
	"testing"

	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/service/model"
)

// TestResetCarsValidations checks payload constraints for the fleet reset endpoint.
func TestResetCarsValidations(t *testing.T) {
	cp := New_CarPool()

	err := cp.ResetCars([]*model.Car{
		{ID: 1, Seats: 3},
	})
	if !errors.Is(err, ErrInvalidCars) {
		t.Fatalf("expected ErrInvalidCars, got %v", err)
	}

	err = cp.ResetCars([]*model.Car{
		{ID: 1, Seats: 4},
		{ID: 1, Seats: 6},
	})
	if !errors.Is(err, ErrDuplicatedID) {
		t.Fatalf("expected ErrDuplicatedID, got %v", err)
	}
}

// TestJourneyAssignmentBestFit validates best-fit selection and deterministic tie-breaking.
func TestJourneyAssignmentBestFit(t *testing.T) {
	cp := New_CarPool()
	err := cp.ResetCars([]*model.Car{
		{ID: 1, Seats: 6},
		{ID: 2, Seats: 4},
		{ID: 3, Seats: 5},
	})
	if err != nil {
		t.Fatalf("unexpected error resetting cars: %v", err)
	}

	assigned, err := cp.NewJourney(&model.Journey{ID: 100, People: 4})
	if err != nil {
		t.Fatalf("unexpected error creating journey 100: %v", err)
	}
	if !assigned {
		t.Fatalf("journey 100 should be assigned immediately")
	}

	car, waiting, err := cp.Locate(100)
	if err != nil {
		t.Fatalf("unexpected locate error: %v", err)
	}
	if waiting {
		t.Fatalf("journey 100 should be riding, not waiting")
	}
	if car.ID != 2 {
		t.Fatalf("expected best-fit car 2, got car %d", car.ID)
	}
}

// TestFairnessAllowsOvertakeWhenHeadCannotFit enforces "FIFO when possible".
func TestFairnessAllowsOvertakeWhenHeadCannotFit(t *testing.T) {
	cp := New_CarPool()
	err := cp.ResetCars([]*model.Car{
		{ID: 1, Seats: 4},
		{ID: 2, Seats: 6},
	})
	if err != nil {
		t.Fatalf("unexpected error resetting cars: %v", err)
	}

	// Occupy the 6-seat car completely.
	assigned, err := cp.NewJourney(&model.Journey{ID: 10, People: 6})
	if err != nil || !assigned {
		t.Fatalf("journey 10 must be assigned immediately, assigned=%v err=%v", assigned, err)
	}

	// Head of queue cannot fit in any remaining car.
	assigned, err = cp.NewJourney(&model.Journey{ID: 11, People: 5})
	if err != nil {
		t.Fatalf("unexpected error creating journey 11: %v", err)
	}
	if assigned {
		t.Fatalf("journey 11 should remain waiting")
	}

	// Later journey can fit the 4-seat car, so it may overtake journey 11.
	assigned, err = cp.NewJourney(&model.Journey{ID: 12, People: 4})
	if err != nil {
		t.Fatalf("unexpected error creating journey 12: %v", err)
	}
	if !assigned {
		t.Fatalf("journey 12 should be assigned by fairness exception")
	}

	_, waiting, err := cp.Locate(11)
	if err != nil {
		t.Fatalf("unexpected locate error for journey 11: %v", err)
	}
	if !waiting {
		t.Fatalf("journey 11 should still be waiting")
	}

	car, waiting, err := cp.Locate(12)
	if err != nil {
		t.Fatalf("unexpected locate error for journey 12: %v", err)
	}
	if waiting {
		t.Fatalf("journey 12 should be riding")
	}
	if car.ID != 1 {
		t.Fatalf("journey 12 expected car 1, got car %d", car.ID)
	}
}

// TestDropoffTriggersCascadeReassign ensures one dropoff can assign multiple waiting groups.
func TestDropoffTriggersCascadeReassign(t *testing.T) {
	cp := New_CarPool()
	err := cp.ResetCars([]*model.Car{
		{ID: 1, Seats: 6},
	})
	if err != nil {
		t.Fatalf("unexpected error resetting cars: %v", err)
	}

	assigned, err := cp.NewJourney(&model.Journey{ID: 1, People: 6})
	if err != nil || !assigned {
		t.Fatalf("journey 1 must be assigned, assigned=%v err=%v", assigned, err)
	}

	assigned, err = cp.NewJourney(&model.Journey{ID: 2, People: 4})
	if err != nil || assigned {
		t.Fatalf("journey 2 must wait, assigned=%v err=%v", assigned, err)
	}

	assigned, err = cp.NewJourney(&model.Journey{ID: 3, People: 2})
	if err != nil || assigned {
		t.Fatalf("journey 3 must wait, assigned=%v err=%v", assigned, err)
	}

	if err := cp.Dropoff(1); err != nil {
		t.Fatalf("unexpected dropoff error: %v", err)
	}

	_, waiting, err := cp.Locate(2)
	if err != nil || waiting {
		t.Fatalf("journey 2 should be riding after cascade, waiting=%v err=%v", waiting, err)
	}

	_, waiting, err = cp.Locate(3)
	if err != nil || waiting {
		t.Fatalf("journey 3 should be riding after cascade, waiting=%v err=%v", waiting, err)
	}
}

// TestDropoffWaitingJourneyRemovesIt verifies waiting journeys can be dropped cleanly.
func TestDropoffWaitingJourneyRemovesIt(t *testing.T) {
	cp := New_CarPool()
	err := cp.ResetCars([]*model.Car{
		{ID: 1, Seats: 4},
	})
	if err != nil {
		t.Fatalf("unexpected error resetting cars: %v", err)
	}

	_, err = cp.NewJourney(&model.Journey{ID: 1, People: 4})
	if err != nil {
		t.Fatalf("unexpected error creating journey 1: %v", err)
	}
	assigned, err := cp.NewJourney(&model.Journey{ID: 2, People: 6})
	if err != nil {
		t.Fatalf("unexpected error creating journey 2: %v", err)
	}
	if assigned {
		t.Fatalf("journey 2 should be waiting")
	}

	if err := cp.Dropoff(2); err != nil {
		t.Fatalf("unexpected dropoff error: %v", err)
	}

	_, _, err = cp.Locate(2)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after dropoff, got %v", err)
	}
}
