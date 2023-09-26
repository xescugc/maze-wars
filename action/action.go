package action

import (
	"github.com/gorilla/websocket"
	"github.com/xescugc/ltw/utils"
)

type Action struct {
	Type Type   `json:"type"`
	Room string `json:"room"`

	CursorMove           *CursorMovePayload           `json:"cursor_move,omitempty"`
	SummonUnit           *SummonUnitPayload           `json:"summon_unit,omitempty"`
	RemoveUnit           *RemoveUnitPayload           `json:"remove_unit,omitempty"`
	StealLive            *StealLivePayload            `json:"steal_live,omitempty"`
	CameraZoom           *CameraZoomPayload           `json:"camera_zoom,omitempty"`
	SelectTower          *SelectTowerPayload          `json:"select_tower,omitempty"`
	PlaceTower           *PlaceTowerPayload           `json:"place_tower,omitempty"`
	SelectedTowerInvalid *SelectedTowerInvalidPayload `json:"selected_tower_invalid,omitempty"`
	TowerAttack          *TowerAttackPayload          `json:"tower_attack,omitempty"`
	UnitKilled           *UnitKilledPayload           `json:"unit_killed,omitempty"`
	WindowResizing       *WindowResizingPayload       `json:"window_resizing,omitempty"`

	AddPlayer   *AddPlayerPayload   `json:"add_player, omitempty"`
	JoinRoom    *JoinRoomPayload    `json:"join_room, omitempty"`
	UpdateState *UpdateStatePayload `json:"update_state, omitempty"`
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

func NewMoveUnit() *Action {
	return &Action{
		Type: MoveUnit,
	}
}

type RemoveUnitPayload struct {
	UnitID string
}

func NewRemoveUnit(uid string) *Action {
	return &Action{
		Type: RemoveUnit,
		RemoveUnit: &RemoveUnitPayload{
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
	ID        string
	Name      string
	LineID    int
	Websocket *websocket.Conn
	Room      string
}

func NewAddPlayer(r, id, name string, lid int, ws *websocket.Conn) *Action {
	return &Action{
		Type: AddPlayer,
		AddPlayer: &AddPlayerPayload{
			ID:        id,
			Name:      name,
			LineID:    lid,
			Websocket: ws,
			Room:      r,
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
}

type UpdateStateTowersPayload struct {
	Towers map[string]*UpdateStateTowerPayload
}

type UpdateStateTowerPayload struct {
	utils.Object

	Type   string
	LineID int
}

type UpdateStateUnitsPayload struct {
	Units map[string]*UpdateStateUnitPayload
}

type UpdateStateUnitPayload struct {
	utils.MovingObject

	Type          string
	PlayerID      string
	PlayerLineID  int
	CurrentLineID int

	Health float64

	Path []utils.Step
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
