package tower

//go:generate enumer -type=Type -transform=lower -output=type_string.go

type Type int

const (
	Range1 Type = iota
	Range2
	RangeSingel1
	RangeSingel2
	RangeAoE1
	RangeAoE2

	Melee1
	Melee2
	MeleeSingle1
	MeleeSingle2
	MeleeAoE1
	MeleeAoE2
)
