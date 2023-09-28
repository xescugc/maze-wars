package store

import (
	"image"

	"github.com/gofrs/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
	"github.com/xescugc/ltw/tower"
	"github.com/xescugc/ltw/unit"
	"github.com/xescugc/ltw/utils"
)

type Units struct {
	*flux.ReduceStore

	store *Store
}

type UnitsState struct {
	Units map[string]*Unit
}

type Unit struct {
	utils.MovingObject

	Type          string
	PlayerID      string
	PlayerLineID  int
	CurrentLineID int

	Health float64

	Path []utils.Step
}

func (u *Unit) Image() image.Image {
	return unit.Units[u.Type].Image
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
		// We wait for the towers store as we need to interact with it
		us.GetDispatcher().WaitFor(us.store.Towers.GetDispatcherToken())
		var w, h float64 = 16, 16
		var x, y float64 = us.store.Map.GetRandomSpawnCoordinatesForLineID(act.SummonUnit.CurrentLineID)
		uid := uuid.Must(uuid.NewV4())
		u := &Unit{
			MovingObject: utils.MovingObject{
				Object: utils.Object{
					X: x, Y: y,
					W: w, H: h,
				},
				Facing: ebiten.KeyS,
			},
			Type:          act.SummonUnit.Type,
			PlayerID:      act.SummonUnit.PlayerID,
			PlayerLineID:  act.SummonUnit.PlayerLineID,
			CurrentLineID: act.SummonUnit.CurrentLineID,
			Health:        10,
		}
		ts := us.store.Towers.GetState().(TowersState)
		tws := make([]utils.Object, 0, 0)
		for _, t := range ts.Towers {
			if t.LineID == u.CurrentLineID {
				tws = append(tws, t.Object)
			}
		}
		u.Path = us.astar(us.store.Map, u.CurrentLineID, u.MovingObject, tws)
		ustate.Units[uid.String()] = u
	case action.MoveUnit:
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
		ts := us.store.Towers.GetState().(TowersState)
		p := us.store.Players.GetPlayerByID(act.PlaceTower.PlayerID)
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

				u.Path = us.astar(us.store.Map, u.CurrentLineID, u.MovingObject, tws)
			}
		}
	case action.RemoveUnit:
		delete(ustate.Units, act.RemoveUnit.UnitID)
	case action.TowerAttack:
		u := ustate.Units[act.TowerAttack.UnitID]
		// For now the damage is just 1
		u.Health -= tower.Towers[act.TowerAttack.TowerType].Damage
		if u.Health <= 0 {
			u.Health = 0
		}
	case action.UpdateState:
		uids := make(map[string]struct{})
		for id := range ustate.Units {
			uids[id] = struct{}{}
		}
		for id, u := range act.UpdateState.Units.Units {
			delete(uids, id)
			nu := Unit(*u)
			ustate.Units[id] = &nu
		}
		for id := range uids {
			delete(ustate.Units, id)
		}
	case action.RemovePlayer:
		for id, u := range ustate.Units {
			if u.PlayerID == act.RemovePlayer.ID {
				delete(ustate.Units, id)
			}
		}
	default:
	}
	return ustate
}