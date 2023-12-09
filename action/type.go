package action

type Type int

//go:generate enumer -type=Type -transform=snake -output=type_string.go -json

const (
	CursorMove Type = iota
	CameraZoom
	SummonUnit
	TPS
	RemoveUnit
	StealLive
	PlaceTower
	RemoveTower
	SelectTower
	SelectedTower
	SelectedTowerInvalid
	DeselectTower
	IncomeTick
	TowerAttack
	UnitKilled
	WindowResizing
	PlayerReady
	NavigateTo
	StartGame
	OpenTowerMenu
	CloseTowerMenu
	GoHome
	CheckedPath

	// Specific to WS
	JoinRoom
	AddPlayer
	RemovePlayer
	UpdateState
)
