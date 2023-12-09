package store

import (
	"image"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
	"github.com/xescugc/ltw/tower"
	"github.com/xescugc/ltw/unit"
	"github.com/xescugc/ltw/utils"
)

type Units struct {
	*flux.ReduceStore

	store *Store

	mxUnits sync.RWMutex
}

type UnitsState struct {
	Units map[string]*Unit
}

type Unit struct {
	utils.MovingObject

	ID string

	Type          string
	PlayerID      string
	PlayerLineID  int
	CurrentLineID int

	Health float64

	Path     []utils.Step
	HashPath string
}

func (u *Unit) Faceset() image.Image {
	return unit.Units[u.Type].Faceset
}

func (u *Unit) Sprite() image.Image {
	return unit.Units[u.Type].Sprite
}

func NewUnits(d *flux.Dispatcher, s *Store) *Units {
	u := &Units{
		store: s,
	}

	u.ReduceStore = flux.NewReduceStore(d, u.Reduce, UnitsState{
		Units: make(map[string]*Unit),
	})

	return u
}

// List returns the units list and it's meant for reading only purposes
func (us *Units) List() []*Unit {
	us.mxUnits.RLock()
	defer us.mxUnits.RUnlock()
	munits := us.GetState().(UnitsState)
	units := make([]*Unit, 0, len(munits.Units))
	for _, u := range munits.Units {
		units = append(units, u)
	}
	return units
}

func (us *Units) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	ustate, ok := state.(UnitsState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.SummonUnit:
		us.GetDispatcher().WaitFor(
			us.store.Towers.GetDispatcherToken(),
		)
		us.mxUnits.Lock()
		defer us.mxUnits.Unlock()

		p := us.store.Players.FindByID(act.SummonUnit.PlayerID)
		if !p.CanSummonUnit(act.SummonUnit.Type) {
			break
		}
		// We wait for the towers store as we need to interact with it
		var w, h float64 = 16, 16
		var x, y float64 = us.store.Map.GetRandomSpawnCoordinatesForLineID(act.SummonUnit.CurrentLineID)
		uid := uuid.Must(uuid.NewV4())
		u := &Unit{
			MovingObject: utils.MovingObject{
				Object: utils.Object{
					X: x, Y: y,
					W: w, H: h,
				},
				Facing: utils.Down,
			},
			ID:            uid.String(),
			Type:          act.SummonUnit.Type,
			PlayerID:      act.SummonUnit.PlayerID,
			PlayerLineID:  act.SummonUnit.PlayerLineID,
			CurrentLineID: act.SummonUnit.CurrentLineID,
			Health:        unit.Units[act.SummonUnit.Type].Health,
		}
		ts := us.store.Towers.List()
		tws := make([]utils.Object, 0, 0)
		for _, t := range ts {
			if t.LineID == u.CurrentLineID {
				tws = append(tws, t.Object)
			}
		}
		u.Path = us.Astar(us.store.Map, u.CurrentLineID, u.MovingObject, tws)
		u.HashPath = utils.HashSteps(u.Path)
		ustate.Units[uid.String()] = u
	case action.TPS:
		us.mxUnits.Lock()
		defer us.mxUnits.Unlock()

		for _, u := range ustate.Units {
			if len(u.Path) > 0 {
				nextStep := u.Path[0]
				u.Path = u.Path[1:]
				u.MovingCount += 1
				u.Y = nextStep.Y
				u.X = nextStep.X
				u.Facing = nextStep.Facing
			}
		}
	case action.PlaceTower:
		// We wait for the towers store as we need to interact with it
		us.GetDispatcher().WaitFor(us.store.Towers.GetDispatcherToken())

		us.mxUnits.Lock()
		defer us.mxUnits.Unlock()

		ts := us.store.Towers.GetState().(TowersState)
		p := us.store.Players.FindByID(act.PlaceTower.PlayerID)

		if !p.CanPlaceTower(act.PlaceTower.Type) {
			break
		}
		for _, u := range ustate.Units {
			// Only need to recalculate path for each unit when the placed tower
			// is on the same LineID as the unit
			if u.CurrentLineID == p.LineID {
				tws := make([]utils.Object, 0, 0)
				for _, t := range ts.Towers {
					if t.LineID == u.CurrentLineID {
						tws = append(tws, t.Object)
					}
				}

				u.Path = us.Astar(us.store.Map, u.CurrentLineID, u.MovingObject, tws)
			}
		}
	case action.RemoveTower:
		us.mxUnits.Lock()
		defer us.mxUnits.Unlock()

		// We wait for the towers store as we need to interact with it
		us.GetDispatcher().WaitFor(us.store.Towers.GetDispatcherToken())
		ts := us.store.Towers.List()
		p := us.store.Players.FindByID(act.RemoveTower.PlayerID)
		for _, u := range ustate.Units {
			// Only need to recalculate path for each unit when the placed tower
			// is on the same LineID as the unit
			if u.CurrentLineID == p.LineID {
				tws := make([]utils.Object, 0, 0)
				for _, t := range ts {
					if t.LineID == u.CurrentLineID {
						tws = append(tws, t.Object)
					}
				}

				u.Path = us.Astar(us.store.Map, u.CurrentLineID, u.MovingObject, tws)
			}
		}
	case action.RemoveUnit:
		us.mxUnits.Lock()
		defer us.mxUnits.Unlock()

		delete(ustate.Units, act.RemoveUnit.UnitID)
	case action.TowerAttack:
		us.mxUnits.Lock()
		defer us.mxUnits.Unlock()

		u, ok := ustate.Units[act.TowerAttack.UnitID]
		if !ok {
			break
		}
		u.Health -= tower.Towers[act.TowerAttack.TowerType].Damage
		if u.Health <= 0 {
			u.Health = 0
		}
	case action.RemovePlayer:
		us.mxUnits.Lock()
		defer us.mxUnits.Unlock()

		for id, u := range ustate.Units {
			if u.PlayerID == act.RemovePlayer.ID {
				delete(ustate.Units, id)
			}
		}
	case action.UpdateState:
		us.mxUnits.Lock()
		defer us.mxUnits.Unlock()

		uids := make(map[string]struct{})
		for id := range ustate.Units {
			uids[id] = struct{}{}
		}
		for id, u := range act.UpdateState.Units.Units {
			delete(uids, id)
			nu := Unit(*u)
			ou, ok := ustate.Units[id]

			if ok {
				if ou.HashPath == nu.HashPath {
					nu.Path = ou.Path
				}
			}

			// If the path is the same we set it. This cannot be done directly
			// as the Unit on the client may be faster than the Unit on the server
			ustate.Units[id] = &nu
		}
		for id := range uids {
			delete(ustate.Units, id)
		}
	}
	return ustate
}
