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
	JoinVs6WaitingRoom
	ExitVs6WaitingRoom
	JoinVs1WaitingRoom
	ExitVs1WaitingRoom
	StartRoom
	ToggleStats
	VersionError
	SetupGame
	FindGame
	ExitSearchingGame
	AcceptWaitingGame
	CancelWaitingGame
	ShowScoreboard

	CreateLobby
	DeleteLobby
	JoinLobby
	AddLobbies
	SelectLobby
	LeaveLobby
	UpdateLobby
	StartLobby
	SeenLobbies

	// Specific to WS
	AddPlayer
	RemovePlayer
	SyncState
	WaitRoomCountdownTick
	SyncVs6WaitingRoom
	SyncVs1WaitingRoom
	SyncSearchingRoom
	SyncWaitingRoom
)
