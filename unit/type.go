package unit

//go:generate enumer -type=Type -transform=lower -transform=snake -output=type_string.go

type Type int

const (
	Ninja Type = iota
	Statue
	Hunter
	Slime
	Mole
	SkeletonDemon
	Butterfly
	BlendMaster
	Robot
	MonkeyBoxer
)
