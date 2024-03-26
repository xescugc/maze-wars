package environment

//go:generate enumer -type=Environment -transform=lower -json -transform=snake -output=environment_string.go

type Environment int

const (
	Terrestrial Environment = iota
	Aerial
)
