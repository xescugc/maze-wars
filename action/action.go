package action

import (
	"github.com/xescugc/ltw/utils"
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
	PlayerReady          *PlayerReadyPayload          `json:"player_ready, omitempty"`
	NavigateTo           *NavigateToPayload           `json:"navigate_to, omitempty"`
	StartGame            *StartGamePayload            `json:"start_game, omitempty"`
	GoHome               *GoHomePayload               `json:"go_home, omitempty"`
	CheckedPath          *CheckedPathPayload          `json:"checked_path,omitempty"`

	OpenTowerMenu  *OpenTowerMenuPayload  `json:"open_tower_menu, omitempty"`
	CloseTowerMenu *CloseTowerMenuPayload `json:"close_tower_menu, omitempty"`

	AddPlayer    *AddPlayerPayload    `json:"add_player, omitempty"`
	RemovePlayer *RemovePlayerPayload `json:"remove_player, omitempty"`
	JoinRoom     *JoinRoomPayload     `json:"join_room, omitempty"`
	UpdateState  *UpdateStatePayload  `json:"update_state, omitempty"`
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

type JoinRoomPayload struct {
	Room string
	Name string
}

func NewJoinRoom(room, name string) *Action {
	return &Action{
		Type: JoinRoom,
		JoinRoom: &JoinRoomPayload{
			Room: room,
			Name: name,
		},
	}
}

type AddPlayerPayload struct {
	ID         string
	Name       string
	LineID     int
	Websocket  *websocket.Conn
	RemoteAddr string
	Room       string
}

func NewAddPlayer(r, id, name string, lid int, ws *websocket.Conn, ra string) *Action {
	return &Action{
		Type: AddPlayer,
		AddPlayer: &AddPlayerPayload{
			ID:         id,
			Name:       name,
			LineID:     lid,
			Websocket:  ws,
			RemoteAddr: ra,
			Room:       r,
		},
	}
}

type RemovePlayerPayload struct {
	ID   string
	Room string
}

func NewRemovePlayer(r, id string) *Action {
	return &Action{
		Type: RemovePlayer,
		RemovePlayer: &RemovePlayerPayload{
			ID:   id,
			Room: r,
		},
	}
}

type PlayerReadyPayload struct {
	ID string
}

func NewPlayerReady(id string) *Action {
	return &Action{
		Type: PlayerReady,
		PlayerReady: &PlayerReadyPayload{
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

type StartGamePayload struct{}

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

type CheckedPathPayload struct {
	Checked bool
}

func NewCheckedPath(cp bool) *Action {
	return &Action{
		Type: CheckedPath,
		CheckedPath: &CheckedPathPayload{
			Checked: cp,
		},
	}
}

type UpdateStatePayload struct {
	Players *UpdateStatePlayersPayload
	Towers  *UpdateStateTowersPayload
	Units   *UpdateStateUnitsPayload
}

type UpdateStatePlayersPayload struct {
	Players     map[string]*UpdateStatePlayerPayload
	IncomeTimer int
}

type UpdateStatePlayerPayload struct {
	ID      string
	Name    string
	Lives   int
	LineID  int
	Income  int
	Gold    int
	Current bool
	Winner  bool
	Ready   bool
}

type UpdateStateTowersPayload struct {
	Towers map[string]*UpdateStateTowerPayload
}

type UpdateStateTowerPayload struct {
	utils.Object

	ID       string
	Type     string
	LineID   int
	PlayerID string
}

type UpdateStateUnitsPayload struct {
	Units map[string]*UpdateStateUnitPayload
}

type UpdateStateUnitPayload struct {
	utils.MovingObject

	ID            string
	Type          string
	PlayerID      string
	PlayerLineID  int
	CurrentLineID int

	Health float64

	Path     []utils.Step
	HashPath string
}

// TODO: or make the action.Action separated or make the store.Player separated
func NewUpdateState(players *UpdateStatePlayersPayload, towers *UpdateStateTowersPayload, units *UpdateStateUnitsPayload) *Action {
	return &Action{
		Type: UpdateState,
		UpdateState: &UpdateStatePayload{
			Players: players,
			Towers:  towers,
			Units:   units,
		},
	}
}
