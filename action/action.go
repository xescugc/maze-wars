package action

import (
	"time"

	"github.com/xescugc/maze-wars/unit"
	"github.com/xescugc/maze-wars/utils"
	"github.com/xescugc/maze-wars/utils/graph"
	"nhooyr.io/websocket"
)

type Action struct {
	Type Type   `json:"type"`
	Room string `json:"room"`

	CursorMove           *CursorMovePayload           `json:"cursor_move,omitempty"`
	SummonUnit           *SummonUnitPayload           `json:"summon_unit,omitempty"`
	UpdateUnit           *UpdateUnitPayload           `json:"update_unit,omitempty"`
	UpdateTower          *UpdateTowerPayload          `json:"update_tower,omitempty"`
	CameraZoom           *CameraZoomPayload           `json:"camera_zoom,omitempty"`
	SelectTower          *SelectTowerPayload          `json:"select_tower,omitempty"`
	PlaceTower           *PlaceTowerPayload           `json:"place_tower,omitempty"`
	RemoveTower          *RemoveTowerPayload          `json:"remove_tower,omitempty"`
	SelectedTowerInvalid *SelectedTowerInvalidPayload `json:"selected_tower_invalid,omitempty"`
	WindowResizing       *WindowResizingPayload       `json:"window_resizing,omitempty"`
	NavigateTo           *NavigateToPayload           `json:"navigate_to,omitempty"`
	StartGame            *StartGamePayload            `json:"start_game,omitempty"`
	GoHome               *GoHomePayload               `json:"go_home,omitempty"`
	ToggleStats          *ToggleStatsPayload          `json:"toggle_stats,omitempty"`
	TPS                  *TPSPayload                  `json:"tps,omitempty"`
	VersionError         *VersionErrorPayload         `json:"version_error,omitempty"`

	OpenTowerMenu  *OpenTowerMenuPayload  `json:"open_tower_menu,omitempty"`
	CloseTowerMenu *CloseTowerMenuPayload `json:"close_tower_menu,omitempty"`

	CreateLobby *CreateLobbyPayload `json:"create_lobby,omitempty"`
	DeleteLobby *DeleteLobbyPayload `json:"delete_lobby,omitempty"`
	JoinLobby   *JoinLobbyPayload   `json:"join_lobby,omitempty"`
	AddLobbies  *AddLobbiesPayload  `json:"add_lobbies,omitempty"`
	SelectLobby *SelectLobbyPayload `json:"select_lobby,omitempty"`
	LeaveLobby  *LeaveLobbyPayload  `json:"leave_lobby,omitempty"`
	UpdateLobby *UpdateLobbyPayload `json:"update_lobby,omitempty"`
	StartLobby  *StartLobbyPayload  `json:"start_lobby,omitempty"`

	UserSignUp  *UserSignUpPayload  `json:"user_sign_up,omitempty"`
	SignUpError *SignUpErrorPayload `json:"sign_in_error,omitempty"`
	UserSignIn  *UserSignInPayload  `json:"user_sign_in,omitempty"`
	UserSignOut *UserSignOutPayload `json:"user_sign_out,omitempty"`

	AddPlayer          *AddPlayerPayload          `json:"add_player,omitempty"`
	RemovePlayer       *RemovePlayerPayload       `json:"remove_player,omitempty"`
	JoinVs6WaitingRoom *JoinVs6WaitingRoomPayload `json:"join_vs6_waiting_room,omitempty"`
	ExitVs6WaitingRoom *ExitVs6WaitingRoomPayload `json:"exit_vs6_waiting_room,omitempty"`
	JoinVs1WaitingRoom *JoinVs1WaitingRoomPayload `json:"join_vs1_waiting_room,omitempty"`
	ExitVs1WaitingRoom *ExitVs1WaitingRoomPayload `json:"exit_vs1_waiting_room,omitempty"`
	StartRoom          *StartRoomPayload          `json:"start_room,omitempty"`
	SyncState          *SyncStatePayload          `json:"sync_state,omitempty"`
	SyncUsers          *SyncUsersPayload          `json:"sync_users,omitempty"`
	SyncVs6WaitingRoom *SyncVs6WaitingRoomPayload `json:"sync_vs6_waiting_room,omitempty"`
	SyncVs1WaitingRoom *SyncVs1WaitingRoomPayload `json:"sync_vs1_waiting_room,omitempty"`
}

type CursorMovePayload struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func NewCursorMove(x, y int) *Action {
	return &Action{
		Type: CursorMove,
		CursorMove: &CursorMovePayload{
			X: x,
			Y: y,
		},
	}
}

type SummonUnitPayload struct {
	Type          string
	PlayerID      string
	PlayerLineID  int
	CurrentLineID int
}

func NewSummonUnit(t, pid string, plid, clid int) *Action {
	return &Action{
		Type: SummonUnit,
		SummonUnit: &SummonUnitPayload{
			Type:          t,
			PlayerID:      pid,
			PlayerLineID:  plid,
			CurrentLineID: clid,
		},
	}
}

type TPSPayload struct {
	Time time.Time
}

func NewTPS(t time.Time) *Action {
	return &Action{
		Type: TPS,
		TPS: &TPSPayload{
			Time: t,
		},
	}
}

type CameraZoomPayload struct {
	Direction int
}

func NewCameraZoom(d int) *Action {
	return &Action{
		Type: CameraZoom,
		CameraZoom: &CameraZoomPayload{
			Direction: d,
		},
	}
}

type PlaceTowerPayload struct {
	Type     string
	PlayerID string
	X        int
	Y        int
}

func NewPlaceTower(t, pid string, x, y int) *Action {
	return &Action{
		Type: PlaceTower,
		PlaceTower: &PlaceTowerPayload{
			Type:     t,
			PlayerID: pid,
			X:        x,
			Y:        y,
		},
	}
}

type RemoveTowerPayload struct {
	PlayerID string
	TowerID  string
}

func NewRemoveTower(pid, tid string) *Action {
	return &Action{
		Type: RemoveTower,
		RemoveTower: &RemoveTowerPayload{
			PlayerID: pid,
			TowerID:  tid,
		},
	}
}

type SelectTowerPayload struct {
	Type string
	X    int
	Y    int
}

func NewSelectTower(t string, x, y int) *Action {
	return &Action{
		Type: SelectTower,
		SelectTower: &SelectTowerPayload{
			Type: t,
			X:    x,
			Y:    y,
		},
	}
}

type SelectedTowerInvalidPayload struct {
	Invalid bool
}

func NewSelectedTowerInvalid(i bool) *Action {
	return &Action{
		Type: SelectedTowerInvalid,
		SelectedTowerInvalid: &SelectedTowerInvalidPayload{
			Invalid: i,
		},
	}
}

func NewDeselectTower(_ string) *Action {
	return &Action{
		Type: DeselectTower,
	}
}

func NewIncomeTick() *Action {
	return &Action{
		Type: IncomeTick,
	}
}

func NewWaitRoomCountdownTick() *Action {
	return &Action{
		Type: WaitRoomCountdownTick,
	}
}

type WindowResizingPayload struct {
	Width  int
	Height int
}

func NewWindowResizing(w, h int) *Action {
	return &Action{
		Type: WindowResizing,
		WindowResizing: &WindowResizingPayload{
			Width:  w,
			Height: h,
		},
	}
}

type AddPlayerPayload struct {
	ID     string
	Name   string
	LineID int
}

func NewAddPlayer(id, name string, lid int) *Action {
	return &Action{
		Type: AddPlayer,
		AddPlayer: &AddPlayerPayload{
			ID:     id,
			Name:   name,
			LineID: lid,
		},
	}
}

type RemovePlayerPayload struct {
	ID string
}

func NewRemovePlayer(id string) *Action {
	return &Action{
		Type: RemovePlayer,
		RemovePlayer: &RemovePlayerPayload{
			ID: id,
		},
	}
}

type NavigateToPayload struct {
	Route string
}

func NewNavigateTo(route string) *Action {
	return &Action{
		Type: NavigateTo,
		NavigateTo: &NavigateToPayload{
			Route: route,
		},
	}
}

type StartGamePayload struct {
}

func NewStartGame() *Action {
	return &Action{
		Type:      StartGame,
		StartGame: &StartGamePayload{},
	}
}

type StartRoomPayload struct {
	RoomID string
}

func NewStartRoom(rid string) *Action {
	return &Action{
		Type: StartRoom,
		StartRoom: &StartRoomPayload{
			RoomID: rid,
		},
	}
}

type OpenTowerMenuPayload struct {
	TowerID string
}

func NewOpenTowerMenu(tid string) *Action {
	return &Action{
		Type: OpenTowerMenu,
		OpenTowerMenu: &OpenTowerMenuPayload{
			TowerID: tid,
		},
	}
}

type CloseTowerMenuPayload struct{}

func NewCloseTowerMenu() *Action {
	return &Action{
		Type:           CloseTowerMenu,
		CloseTowerMenu: &CloseTowerMenuPayload{},
	}
}

type GoHomePayload struct{}

func NewGoHome() *Action {
	return &Action{
		Type:   GoHome,
		GoHome: &GoHomePayload{},
	}
}

type ToggleStatsPayload struct {
}

func NewToggleStats() *Action {
	return &Action{
		Type:        ToggleStats,
		ToggleStats: &ToggleStatsPayload{},
	}
}

type SignUpErrorPayload struct {
	Error string
}

func NewSignUpError(e string) *Action {
	return &Action{
		Type: SignUpError,
		SignUpError: &SignUpErrorPayload{
			Error: e,
		},
	}
}

type UserSignInPayload struct {
	Username   string
	Websocket  *websocket.Conn
	RemoteAddr string
}

// NewUserSignIn initializes the UserSignIn with just the username
// the rest of the data needs to be manually set by someone else
func NewUserSignIn(un string) *Action {
	return &Action{
		Type: UserSignIn,
		UserSignIn: &UserSignInPayload{
			Username: un,
		},
	}
}

type UserSignOutPayload struct {
	Username string
}

func NewUserSignOut(un string) *Action {
	return &Action{
		Type: UserSignOut,
		UserSignOut: &UserSignOutPayload{
			Username: un,
		},
	}
}

type UserSignUpPayload struct {
	Username string
}

func NewUserSignUp(un string) *Action {
	return &Action{
		Type: UserSignUp,
		UserSignUp: &UserSignUpPayload{
			Username: un,
		},
	}
}

type JoinVs6WaitingRoomPayload struct {
	Username string
}

func NewJoinVs6WaitingRoom(un string) *Action {
	return &Action{
		Type: JoinVs6WaitingRoom,
		JoinVs6WaitingRoom: &JoinVs6WaitingRoomPayload{
			Username: un,
		},
	}
}

type ExitVs6WaitingRoomPayload struct {
	Username string
}

func NewExitVs6WaitingRoom(un string) *Action {
	return &Action{
		Type: ExitVs6WaitingRoom,
		ExitVs6WaitingRoom: &ExitVs6WaitingRoomPayload{
			Username: un,
		},
	}
}

type SyncVs6WaitingRoomPayload struct {
	TotalPlayers int
	Size         int
	Countdown    int
}

func NewSyncVs6WaitingRoom(tp, s, cd int) *Action {
	return &Action{
		Type: SyncVs6WaitingRoom,
		SyncVs6WaitingRoom: &SyncVs6WaitingRoomPayload{
			TotalPlayers: tp,
			Size:         s,
			Countdown:    cd,
		},
	}
}

type JoinVs1WaitingRoomPayload struct {
	Username string
}

func NewJoinVs1WaitingRoom(un string) *Action {
	return &Action{
		Type: JoinVs1WaitingRoom,
		JoinVs1WaitingRoom: &JoinVs1WaitingRoomPayload{
			Username: un,
		},
	}
}

type ExitVs1WaitingRoomPayload struct {
	Username string
}

func NewExitVs1WaitingRoom(un string) *Action {
	return &Action{
		Type: ExitVs1WaitingRoom,
		ExitVs1WaitingRoom: &ExitVs1WaitingRoomPayload{
			Username: un,
		},
	}
}

type SyncVs1WaitingRoomPayload struct {
	TotalPlayers int
	Size         int
}

func NewSyncVs1WaitingRoom(tp, s int) *Action {
	return &Action{
		Type: SyncVs1WaitingRoom,
		SyncVs1WaitingRoom: &SyncVs1WaitingRoomPayload{
			TotalPlayers: tp,
			Size:         s,
		},
	}
}

type SyncStatePayload struct {
	Players *SyncStatePlayersPayload
	Lines   *SyncStateLinesPayload
}

type SyncStateLinesPayload struct {
	Lines map[int]*SyncStateLinePayload
}

type SyncStateLinePayload struct {
	ID     int
	Towers map[string]*SyncStateTowerPayload
	Units  map[string]*SyncStateUnitPayload
}

type SyncStatePlayersPayload struct {
	Players     map[string]*SyncStatePlayerPayload
	IncomeTimer int
}

type SyncStatePlayerPayload struct {
	ID      string
	Name    string
	Lives   int
	LineID  int
	Income  int
	Gold    int
	Current bool
	Winner  bool

	UnitUpdates map[string]SyncStatePlayerUnitUpdatePayload
}

type SyncStatePlayerUnitUpdatePayload struct {
	Current    unit.Stats
	Level      int
	UpdateCost int
	Next       unit.Stats
}

type SyncStateTowerPayload struct {
	utils.Object

	ID       string
	Type     string
	LineID   int
	PlayerID string

	Health float64

	TargetUnitID string
	LastAttack   time.Time
}

type SyncStateUnitPayload struct {
	utils.MovingObject
	AnimationCount int

	ID            string
	Type          string
	PlayerID      string
	PlayerLineID  int
	CurrentLineID int

	MaxHealth float64
	Health    float64

	MaxShield float64
	Shield    float64

	MovementSpeed float64
	Bounty        int

	Level int

	Path      []graph.Step
	HashPath  string
	CreatedAt time.Time

	Abilities map[string]interface{}
	Buffs     map[string]interface{}

	TargetTowerID string
	LastAttack    time.Time
}

// TODO: or make the action.Action separated or make the store.Player separated
func NewSyncState(players *SyncStatePlayersPayload, lines *SyncStateLinesPayload) *Action {
	return &Action{
		Type: SyncState,
		SyncState: &SyncStatePayload{
			Players: players,
			Lines:   lines,
		},
	}
}

type SyncUsersPayload struct {
	TotalUsers int
}

func NewSyncUsers(totalUsers int) *Action {
	return &Action{
		Type: SyncUsers,
		SyncUsers: &SyncUsersPayload{
			TotalUsers: totalUsers,
		},
	}
}

type VersionErrorPayload struct {
	Error string
}

func NewVersionError(err string) *Action {
	return &Action{
		Type: VersionError,
		VersionError: &VersionErrorPayload{
			Error: err,
		},
	}
}

type UpdateUnitPayload struct {
	Type     string
	PlayerID string
}

func NewUpdateUnit(pid, t string) *Action {
	return &Action{
		Type: UpdateUnit,
		UpdateUnit: &UpdateUnitPayload{
			Type:     t,
			PlayerID: pid,
		},
	}
}

type UpdateTowerPayload struct {
	TowerID   string
	PlayerID  string
	TowerType string
}

func NewUpdateTower(pid, tid, tt string) *Action {
	return &Action{
		Type: UpdateTower,
		UpdateTower: &UpdateTowerPayload{
			TowerID:   tid,
			PlayerID:  pid,
			TowerType: tt,
		},
	}
}

type CreateLobbyPayload struct {
	LobbyID         string
	Owner           string
	LobbyName       string
	LobbyMaxPlayers int
}

func NewCreateLobby(lid, o, ln string, lmp int) *Action {
	return &Action{
		Type: CreateLobby,
		CreateLobby: &CreateLobbyPayload{
			LobbyID:         lid,
			Owner:           o,
			LobbyName:       ln,
			LobbyMaxPlayers: lmp,
		},
	}
}

type DeleteLobbyPayload struct {
	LobbyID string
}

func NewDeleteLobby(lid string) *Action {
	return &Action{
		Type: DeleteLobby,
		DeleteLobby: &DeleteLobbyPayload{
			LobbyID: lid,
		},
	}
}

type JoinLobbyPayload struct {
	LobbyID  string
	Username string
	IsBot    bool
}

func NewJoinLobby(lid, un string, ib bool) *Action {
	return &Action{
		Type: JoinLobby,
		JoinLobby: &JoinLobbyPayload{
			LobbyID:  lid,
			Username: un,
			IsBot:    ib,
		},
	}
}

type AddLobbiesPayload struct {
	Lobbies []*LobbyPayload
}

type LobbyPayload struct {
	ID         string
	Name       string
	MaxPlayers int

	Players map[string]bool

	Owner string
}

func NewAddLobbies(lbs *AddLobbiesPayload) *Action {
	return &Action{
		Type:       AddLobbies,
		AddLobbies: lbs,
	}
}

type SelectLobbyPayload struct {
	LobbyID string
}

func NewSelectLobby(lbi string) *Action {
	return &Action{
		Type: SelectLobby,
		SelectLobby: &SelectLobbyPayload{
			LobbyID: lbi,
		},
	}
}

type LeaveLobbyPayload struct {
	LobbyID  string
	Username string
}

func NewLeaveLobby(lbi, un string) *Action {
	return &Action{
		Type: LeaveLobby,
		LeaveLobby: &LeaveLobbyPayload{
			LobbyID:  lbi,
			Username: un,
		},
	}
}

type UpdateLobbyPayload struct {
	Lobby LobbyPayload
}

func NewUpdateLobby(l LobbyPayload) *Action {
	return &Action{
		Type: UpdateLobby,
		UpdateLobby: &UpdateLobbyPayload{
			Lobby: l,
		},
	}
}

type StartLobbyPayload struct {
	LobbyID string
}

func NewStartLobby(lid string) *Action {
	return &Action{
		Type: StartLobby,
		StartLobby: &StartLobbyPayload{
			LobbyID: lid,
		},
	}
}
