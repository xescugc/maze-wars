package store

import (
	"image"

	"github.com/gofrs/uuid"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
	"github.com/xescugc/ltw/tower"
	"github.com/xescugc/ltw/utils"
)

type Towers struct {
	*flux.ReduceStore

	store *Store
}

type TowersState struct {
	Towers map[string]*Tower
}

type Tower struct {
	utils.Object

	Type     string
	LineID   int
	PlayerID string
}

func (t *Tower) Image() image.Image {
	return tower.Towers[t.Type].Image
}

func NewTowers(d *flux.Dispatcher, s *Store) *Towers {
	t := &Towers{
		store: s,
	}

	t.ReduceStore = flux.NewReduceStore(d, t.Reduce, TowersState{
		Towers: make(map[string]*Tower),
	})

	return t
}

func (ts *Towers) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	tstate, ok := state.(TowersState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.PlaceTower:
		p := ts.store.Players.GetPlayerByID(act.PlaceTower.PlayerID)

		var w, h float64 = 16 * 2, 16 * 2
		tid := uuid.Must(uuid.NewV4())
		tstate.Towers[tid.String()] = &Tower{
			Object: utils.Object{
				X: float64(act.PlaceTower.X), Y: float64(act.PlaceTower.Y),
				W: w, H: h,
			},
			Type:     act.PlaceTower.Type,
			LineID:   p.LineID,
			PlayerID: p.ID,
		}
	case action.UpdateState:
		tids := make(map[string]struct{})
		for id := range tstate.Towers {
			tids[id] = struct{}{}
		}
		for id, t := range act.UpdateState.Towers.Towers {
			delete(tids, id)
			nt := Tower(*t)
			tstate.Towers[id] = &nt
		}
		for id := range tids {
			delete(tstate.Towers, id)
		}
	case action.RemovePlayer:
		for id, t := range tstate.Towers {
			if t.PlayerID == act.RemovePlayer.ID {
				delete(tstate.Towers, id)
			}
		}
	default:
	}
	return tstate
}