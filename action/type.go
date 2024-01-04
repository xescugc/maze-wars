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
	ChangeUnitLine
	SignUpError
	UserSignUp
	UserSignIn
	UserSignOut
	JoinWaitingRoom
	ExitWaitingRoom

	// Specific to WS
	AddPlayer
	RemovePlayer
	UpdateState
	UpdateUsers
	WaitRoomCountdownTick
	SyncWaitingRoom
)
