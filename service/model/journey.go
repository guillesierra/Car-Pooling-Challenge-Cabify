package model

type Journey struct {
	ID         uint `json:"id"`     // Journey identifier.
	People     uint `json:"people"` // Group size.
	AssignedTo *Car `json:"-"`      // Assigned car, nil when waiting.
}
