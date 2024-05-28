package store

import "time"

type AbilitySplit struct {
	// UnitID  is the other unit
	// that belong to this Split.
	// The bounty must only be given
	// when all are dead
	UnitID string `mapstructure:"UnitID"`
}

const (
	burrowTime    = time.Second * 15
	maxBurrowTime = time.Second * 45
)

type AbilityBurrow struct {
	// BurrowAt is the time in which it was burrowed.
	// It'll stay there for 15s and then the next unit
	// that steps on it it'll pop up again.
	// If in 45s no unit stept on it it'll unburrow
	// itself up
	BurrowAt   time.Time `mapstructure:"BurrowAt"`
	Unburrowed bool      `mapstructure:"Unburrowed"`
}

func (ab AbilityBurrow) CanUnburrow(t time.Time) bool {
	return t.Sub(ab.BurrowAt) > burrowTime
}

func (ab AbilityBurrow) MustUnburrow(t time.Time) bool {
	return t.Sub(ab.BurrowAt) > maxBurrowTime
}
