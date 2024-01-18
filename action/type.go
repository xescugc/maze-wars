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
	NavigateTo
	StartGame
	OpenTowerMenu
	CloseTowerMenu
	GoHome
	CheckedPath
	ChangeUnitLine
	SignUpError
	UserSignUp
	UserSignIn
	UserSignOut
	JoinWaitingRoom
	ExitWaitingRoom
	ToggleStats

	// Specific to WS
	AddPlayer
	RemovePlayer
	SyncState
	SyncUsers
	WaitRoomCountdownTick
	SyncWaitingRoom
)
