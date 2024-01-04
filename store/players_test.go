package store_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/store"
)

func TestNewPlayers(t *testing.T) {
	d := flux.NewDispatcher()
	st := store.NewStore(d)
	ps := store.NewPlayers(d, st)
	pstate := ps.GetState().(store.PlayersState)
	epstate := store.PlayersState{
		Players:     make(map[string]*store.Player),
		IncomeTimer: 15,
	}
	assert.Equal(t, epstate, pstate)
}

func TestPlayers_List(t *testing.T) {
	d := flux.NewDispatcher()
	st := store.NewStore(d)
	ps := store.NewPlayers(d, st)

	id := "id"
	name := "name"
	lid := 2
	// To have any player we have to add it first
	d.Dispatch(action.NewAddPlayer(id, name, lid))

	playes := ps.List()
	eplayers := []*store.Player{
		&store.Player{
			ID:     id,
			Name:   name,
			LineID: lid,
			Lives:  20,
			Income: 25,
			Gold:   40,
		},
	}

	assert.Equal(t, eplayers, playes)
}

func TestPlayers_FindCurrent(t *testing.T) {
	d := flux.NewDispatcher()
	st := store.NewStore(d)
	ps := store.NewPlayers(d, st)
	id := "id"
	name := "name"
	lid := 2
	// To have any player we have to add it first
	d.Dispatch(action.NewAddPlayer(id, name, lid))

	cp := ps.FindCurrent()

	assert.Empty(t, cp)

	pstate := ps.GetState().(store.PlayersState)
	// NOTE: There is no way to set the current value,
	// it's set directly from the server when sending
	// the state back
	pstate.Players[id].Current = true

	cp = ps.FindCurrent()
	ecp := store.Player{
		ID:      id,
		Name:    name,
		LineID:  lid,
		Lives:   20,
		Income:  25,
		Gold:    40,
		Current: true,
	}

	assert.Equal(t, ecp, cp)
}

func TestPlayers_FindByID(t *testing.T) {
	d := flux.NewDispatcher()
	st := store.NewStore(d)
	ps := store.NewPlayers(d, st)
	id := "id"
	name := "name"
	lid := 2
	// To have any player we have to add it first
	d.Dispatch(action.NewAddPlayer(id, name, lid))

	cp := ps.FindByID("none")

	assert.Empty(t, cp)

	cp = ps.FindByID(id)
	ecp := store.Player{
		ID:     id,
		Name:   name,
		LineID: lid,
		Lives:  20,
		Income: 25,
		Gold:   40,
	}

	assert.Equal(t, ecp, cp)
}

func TestPlayers_FindByLineID(t *testing.T) {
	d := flux.NewDispatcher()
	st := store.NewStore(d)
	ps := store.NewPlayers(d, st)
	id := "id"
	name := "name"
	lid := 2
	// To have any player we have to add it first
	d.Dispatch(action.NewAddPlayer(id, name, lid))

	cp := ps.FindByLineID(99)

	assert.Empty(t, cp)

	cp = ps.FindByLineID(lid)
	ecp := store.Player{
		ID:     id,
		Name:   name,
		LineID: lid,
		Lives:  20,
		Income: 25,
		Gold:   40,
	}

	assert.Equal(t, ecp, cp)
}
