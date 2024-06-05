package ability

//go:generate enumer -type=Ability -transform=lower -json -transform=snake -output=ability_string.go

type Ability int

const (
	Split Ability = iota
	Burrow
	Resurrection
	Hybrid
	Camouflage
)
