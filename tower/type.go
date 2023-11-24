package tower

//go:generate enumer -type=Type -transform=lower -output=type_string.go

type Type int

const (
	Soldier Type = iota
	Monk
)
