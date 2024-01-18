package store

import (
	"image"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/utils"
)

type Towers struct {
	*flux.ReduceStore

	store *Store

	mxTowers sync.RWMutex
}

type TowersState struct {
	Towers map[string]*Tower
}

type Tower struct {
	utils.Object

	ID string

	Type     string
	LineID   int
	PlayerID string
}

func (t *Tower) Faceset() image.Image {
	return tower.Towers[t.Type].Faceset
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

// List returns the towers list and it's meant for reading only purposes
func (ts *Towers) List() []*Tower {
	ts.mxTowers.RLock()
	defer ts.mxTowers.RUnlock()
	mtowers := ts.GetState().(TowersState)
	towers := make([]*Tower, 0, len(mtowers.Towers))
	for _, t := range mtowers.Towers {
		towers = append(towers, t)
	}
	return towers
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
		ts.mxTowers.Lock()
		defer ts.mxTowers.Unlock()

		p := ts.store.Players.FindByID(act.PlaceTower.PlayerID)

		if !p.CanPlaceTower(act.PlaceTower.Type) {
			break
		}

		var w, h float64 = 16 * 2, 16 * 2
		tid := uuid.Must(uuid.NewV4())
		tstate.Towers[tid.String()] = &Tower{
			ID: tid.String(),
			Object: utils.Object{
				X: float64(act.PlaceTower.X), Y: float64(act.PlaceTower.Y),
				W: w, H: h,
			},
			Type:     act.PlaceTower.Type,
			LineID:   p.LineID,
			PlayerID: p.ID,
		}
	case action.SyncState:
		ts.mxTowers.Lock()
		defer ts.mxTowers.Unlock()

		tids := make(map[string]struct{})
		for id := range tstate.Towers {
			tids[id] = struct{}{}
		}
		for id, t := range act.SyncState.Towers.Towers {
			delete(tids, id)
			nt := Tower(*t)
			tstate.Towers[id] = &nt
		}
		for id := range tids {
			delete(tstate.Towers, id)
		}
	case action.RemovePlayer:
		ts.mxTowers.Lock()
		defer ts.mxTowers.Unlock()

		for id, t := range tstate.Towers {
			if t.PlayerID == act.RemovePlayer.ID {
				delete(tstate.Towers, id)
			}
		}
	case action.RemoveTower:
		ts.mxTowers.Lock()
		defer ts.mxTowers.Unlock()

		delete(tstate.Towers, act.RemoveTower.TowerID)
	}
	return tstate
}
