package store_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/utils"
)

func TestNewTowers(t *testing.T) {
	d := flux.NewDispatcher()
	st := store.NewStore(d, newEmptyLogger())
	ts := store.NewTowers(d, st)
	tstate := ts.GetState().(store.TowersState)
	etstate := store.TowersState{
		Towers: make(map[string]*store.Tower),
	}
	assert.Equal(t, etstate, tstate)
}

func TestTowers_List(t *testing.T) {
	d := flux.NewDispatcher()
	st := store.NewStore(d, newEmptyLogger())
	ts := store.NewTowers(d, st)

	player := addPlayer(st)

	d.Dispatch(action.NewPlaceTower(tower.Soldier.String(), player.ID, 10, 20))

	towers := ts.List()
	etowers := []*store.Tower{
		&store.Tower{
			// As the ID is a UUID we cannot guess it
			ID: towers[0].ID,
			Object: utils.Object{
				X: 10, Y: 20,
				W: 16 * 2, H: 16 * 2,
			},
			Type:     tower.Soldier.String(),
			LineID:   player.LineID,
			PlayerID: player.ID,
		},
	}
	assert.Equal(t, etowers, towers)
}
