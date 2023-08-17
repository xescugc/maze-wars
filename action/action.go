package action

type Action struct {
	Type Type `json:"type"`

	CursorMove           *CursorMovePayload           `json:"cursor_move,omitempty"`
	SummonUnit           *SummonUnitPayload           `json:"summon_unit,omitempty"`
	RemoveUnit           *RemoveUnitPayload           `json:"remove_unit,omitempty"`
	StealLive            *StealLivePayload            `json:"steal_live,omitempty"`
	CameraZoom           *CameraZoomPayload           `json:"camera_zoom,omitempty"`
	SelectTower          *SelectTowerPayload          `json:"select_tower,omitempty"`
	PlaceTower           *PlaceTowerPayload           `json:"place_tower,omitempty"`
	SelectedTowerInvalid *SelectedTowerInvalidPayload `json:"selected_tower_invalid",omitempty"`
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
	PlayerID      int
	PlayerLineID  int
	CurrentLineID int
}

func NewSummonUnit(t string, pid, plid, clid int) *Action {
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
	UnitID int
}

func NewRemoveUnit(uid int) *Action {
	return &Action{
		Type: RemoveUnit,
		RemoveUnit: &RemoveUnitPayload{
			UnitID: uid,
		},
	}
}

type StealLivePayload struct {
	FromPlayerID int
	ToPlayerID   int
}

func NewStealLive(fpid, tpid int) *Action {
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
	Type   string
	LineID int
	X      int
	Y      int
}

func NewPlaceTower(t string, x, y, lid int) *Action {
	return &Action{
		Type: PlaceTower,
		PlaceTower: &PlaceTowerPayload{
			Type:   t,
			LineID: lid,
			X:      x,
			Y:      y,
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

func NewDeselectTower(t string) *Action {
	return &Action{
		Type: DeselectTower,
	}
}
