package utils

//go:generate enumer -type=Direction -transform=lower -output=direction_string.go

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)
