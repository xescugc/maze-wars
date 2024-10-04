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
	OpenUnitMenu
	CloseTowerMenu
	CloseUnitMenu
	GoHome
	SignUpError
	UserSignUp
	UserSignUpChangeImage
	UserSignIn
	UserSignOut
	VersionError
	SetupGame
	FindGame
	ExitSearchingGame
	AcceptWaitingGame
	CancelWaitingGame
	ShowScoreboard
	AddError

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
	SyncSearchingRoom
	SyncWaitingRoom
)
