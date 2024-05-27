package store

type AbilitySplit struct {
	// UnitID  is the other unit
	// that belong to this Split.
	// The bounty must only be given
	// when all are dead
	UnitID string
}
