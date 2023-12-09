package store_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
	"github.com/xescugc/ltw/store"
	"github.com/xescugc/ltw/unit"
	"github.com/xescugc/ltw/utils"
)

func TestNewUnits(t *testing.T) {
	d := flux.NewDispatcher()
	st := store.NewStore(d)
	us := store.NewUnits(d, st)
	ustate := us.GetState().(store.UnitsState)
	eustate := store.UnitsState{
		Units: make(map[string]*store.Unit),
	}
	assert.Equal(t, eustate, ustate)
}

func TestUnits_List(t *testing.T) {
	d := flux.NewDispatcher()
	st := store.NewStore(d)
	us := store.NewUnits(d, st)

	player := addPlayer(st)
	clid := 2

	d.Dispatch(action.NewSummonUnit(unit.Spirit.String(), player.ID, player.LineID, clid))

	units := us.List()
	eunits := []*store.Unit{
		&store.Unit{
			// As the ID is a UUID we cannot guess it
			ID: units[0].ID,
			MovingObject: utils.MovingObject{
				Object: utils.Object{
					// This is also random
					X: units[0].X, Y: units[0].Y,
					W: 16, H: 16,
				},
				Facing: utils.Down,
			},
			Type:          unit.Spirit.String(),
			PlayerID:      player.ID,
			PlayerLineID:  player.LineID,
			CurrentLineID: clid,
			Health:        unit.Units[unit.Spirit.String()].Health,
		},
	}
	// We calculate the path also
	eunits[0].Path = us.Astar(st.Map, clid, eunits[0].MovingObject, nil)
	eunits[0].HashPath = utils.HashSteps(eunits[0].Path)
	assert.Equal(t, eunits, units)
}
