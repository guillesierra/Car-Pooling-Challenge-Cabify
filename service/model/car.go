package model

type Car struct {
	ID             uint `json:"id"`    // Car identifier.
	Seats          uint `json:"seats"` // Total car capacity.
	AvailableSeats uint `json:"-"`     // Free seats at runtime.
}
