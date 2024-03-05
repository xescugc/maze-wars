package action

import (
	"github.com/xescugc/maze-wars/utils"
	"github.com/xescugc/maze-wars/utils/graph"
	"nhooyr.io/websocket"
)

type Action struct {
	Type Type   `json:"type"`
	Room string `json:"room"`

	CursorMove           *CursorMovePayload           `json:"cursor_move,omitempty"`
	SummonUnit           *SummonUnitPayload           `json:"summon_unit,omitempty"`
	RemoveUnit           *RemoveUnitPayload           `json:"remove_unit,omitempty"`
	ChangeUnitLine       *ChangeUnitLinePayload       `json:"change_unit_line,omitempty"`
	StealLive            *StealLivePayload            `json:"steal_live,omitempty"`
	CameraZoom           *CameraZoomPayload           `json:"camera_zoom,omitempty"`
	SelectTower          *SelectTowerPayload          `json:"select_tower,omitempty"`
	PlaceTower           *PlaceTowerPayload           `json:"place_tower,omitempty"`
	RemoveTower          *RemoveTowerPayload          `json:"remove_tower,omitempty"`
	SelectedTowerInvalid *SelectedTowerInvalidPayload `json:"selected_tower_invalid,omitempty"`
	TowerAttack          *TowerAttackPayload          `json:"tower_attack,omitempty"`
	UnitKilled           *UnitKilledPayload           `json:"unit_killed,omitempty"`
	WindowResizing       *WindowResizingPayload       `json:"window_resizing,omitempty"`
	NavigateTo           *NavigateToPayload           `json:"navigate_to, omitempty"`
	StartGame            *StartGamePayload            `json:"start_game, omitempty"`
	GoHome               *GoHomePayload               `json:"go_home, omitempty"`
	ToggleStats          *ToggleStatsPayload          `json:"toggle_stats,omitempty"`

	OpenTowerMenu  *OpenTowerMenuPayload  `json:"open_tower_menu, omitempty"`
	CloseTowerMenu *CloseTowerMenuPayload `json:"close_tower_menu, omitempty"`

	UserSignUp  *UserSignUpPayload  `json:"user_sign_up, omitempty"`
	SignUpError *SignUpErrorPayload `json:"sign_in_error, omitempty"`
	UserSignIn  *UserSignInPayload  `json:"user_sign_in, omitempty"`
	UserSignOut *UserSignOutPayload `json:"user_sign_out, omitempty"`

	AddPlayer       *AddPlayerPayload       `json:"add_player, omitempty"`
	RemovePlayer    *RemovePlayerPayload    `json:"remove_player, omitempty"`
	JoinWaitingRoom *JoinWaitingRoomPayload `json:"join_waiting_room, omitempty"`
	ExitWaitingRoom *ExitWaitingRoomPayload `json:"exit_waiting_room, omitempty"`
	SyncState       *SyncStatePayload       `json:"sync_state, omitempty"`
	SyncUsers       *SyncUsersPayload       `json:"sync_users, omitempty"`
	SyncWaitingRoom *SyncWaitingRoomPayload `json:"sync_waiting_room, omitempty"`
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

func NewTPS() *Action {
	return &Action{
		Type: TPS,
	}
}

type RemoveUnitPayload struct{ UnitID string }

func NewRemoveUnit(uid string) *Action {
	return &Action{
		Type: RemoveUnit,
		RemoveUnit: &RemoveUnitPayload{
			UnitID: uid,
		},
	}
}

type ChangeUnitLinePayload struct {
	UnitID string
}

func NewChangeUnitLine(uid string) *Action {
	return &Action{
		Type: ChangeUnitLine,
		ChangeUnitLine: &ChangeUnitLinePayload{
			UnitID: uid,
		},
	}
}

type StealLivePayload struct {
	FromPlayerID string
	ToPlayerID   string
}

func NewStealLive(fpid, tpid string) *Action {
	return &Action{
		Type: StealLive,
		StealLive: &StealLivePayload{
			FromPlayerID: fpid,
			ToPlayerID:   tpid,
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
	PlayerID  string
	TowerID   string
	TowerType string
}

func NewRemoveTower(pid, tid, tt string) *Action {
	return &Action{
		Type: RemoveTower,
		RemoveTower: &RemoveTowerPayload{
			PlayerID:  pid,
			TowerID:   tid,
			TowerType: tt,
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

type TowerAttackPayload struct {
	TowerType string
	UnitID    string
}

func NewTowerAttack(uid, tt string) *Action {
	return &Action{
		Type: TowerAttack,
		TowerAttack: &TowerAttackPayload{
			UnitID:    uid,
			TowerType: tt,
		},
	}
}

type UnitKilledPayload struct {
	PlayerID string
	UnitType string
}

func NewUnitKilled(pid, ut string) *Action {
	return &Action{
		Type: UnitKilled,
		UnitKilled: &UnitKilledPayload{
			PlayerID: pid,
			UnitType: ut,
		},
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

type JoinWaitingRoomPayload struct {
	Username string
}

func NewJoinWaitingRoom(un string) *Action {
	return &Action{
		Type: JoinWaitingRoom,
		JoinWaitingRoom: &JoinWaitingRoomPayload{
			Username: un,
		},
	}
}

type ExitWaitingRoomPayload struct {
	Username string
}

func NewExitWaitingRoom(un string) *Action {
	return &Action{
		Type: ExitWaitingRoom,
		ExitWaitingRoom: &ExitWaitingRoomPayload{
			Username: un,
		},
	}
}

type SyncWaitingRoomPayload struct {
	TotalPlayers int
	Size         int
	Countdown    int
}

func NewSyncWaitingRoom(tp, s, cd int) *Action {
	return &Action{
		Type: SyncWaitingRoom,
		SyncWaitingRoom: &SyncWaitingRoomPayload{
			TotalPlayers: tp,
			Size:         s,
			Countdown:    cd,
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
}

type SyncStateTowerPayload struct {
	utils.Object

	ID       string
	Type     string
	LineID   int
	PlayerID string
}

type SyncStateUnitPayload struct {
	utils.MovingObject

	ID            string
	Type          string
	PlayerID      string
	PlayerLineID  int
	CurrentLineID int

	Health float64

	Path     []graph.Step
	HashPath string
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
