package buff

//go:generate enumer -type=Buff -transform=lower -json -transform=snake -output=buff_string.go

type Buff int

const (
	Burrowoed Buff = iota
)
