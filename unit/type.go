package unit

//go:generate enumer -type=Type -transform=lower -transform=snake -output=type_string.go

type Type int

const (
	Spirit Type = iota
	Flam
	Octopus
	Raccon
	Cyclope
	Eye
	Beast
	Butterfly
	Mole
	Skull
	Snake
)
