package unit

//go:generate enumer -type=Type -transform=lower -transform=snake -output=type_string.go

type Type int

const (
	Spirit Type = iota
	Spirit2
	Flam
	Flam2
	Octopus
	Octopus2
	Raccon
	GoldRacoon
	Cyclope
	Cyclope2
)
