package action

type Type int

//go:generate enumer -type=Type -transform=lower -output=type_string.go -json

const (
	CameraMove Type = iota
)
