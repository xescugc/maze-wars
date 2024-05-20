package action

type Type int

//go:generate enumer -type=Type -transform=snake -output=type_string.go -json

const (
	CursorMove Type = iota
	CameraZoom
	SummonUnit
	UpdateUnit
	UpdateTower
	TPS
	PlaceTower
	RemoveTower
	SelectTower
	SelectedTower
	SelectedTowerInvalid
	DeselectTower
	IncomeTick
	WindowResizing
	NavigateTo
	StartGame
	OpenTowerMenu
	CloseTowerMenu
	GoHome
	SignUpError
	UserSignUp
	UserSignIn
	UserSignOut
	JoinWaitingRoom
	ExitWaitingRoom
	StartRoom
	ToggleStats
	VersionError

	CreateLobby
	DeleteLobby
	JoinLobby
	AddLobbies
	SelectLobby
	LeaveLobby
	UpdateLobby
	StartLobby

	// Specific to WS
	AddPlayer
	RemovePlayer
	SyncState
	SyncUsers
	WaitRoomCountdownTick
	SyncWaitingRoom
)
