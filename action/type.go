package action

type Type int

//go:generate enumer -type=Type -transform=snake -output=type_string.go -json

const (
	CursorMove Type = iota
	CameraZoom
	SummonUnit
	MoveUnit
	RemoveUnit
	StealLive
	PlaceTower
	SelectTower
	SelectedTower
	SelectedTowerInvalid
	DeselectTower
	IncomeTick
	TowerAttack
	UnitKilled
	WindowResizing

	// Specific to WS
	JoinRoom
	AddPlayer
	RemovePlayer
	UpdateState
)
