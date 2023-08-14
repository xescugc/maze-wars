package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
)

type PlayersStore struct {
	*flux.ReduceStore
	CurrentPlayerID int
}

type PlayersState struct {
	Players map[int]*PlayerState
}

type PlayerState struct {
	ID     int
	Lives  int
	LineID int
}

func NewPlayersStore(d *flux.Dispatcher) *PlayersStore {
	// TODO: We fake 2 users and we are the 0
	ps := &PlayersStore{
		CurrentPlayerID: 0,
	}
	ps.ReduceStore = flux.NewReduceStore(d, ps.Reduce, PlayersState{
		Players: map[int]*PlayerState{
			0: &PlayerState{
				ID:     0,
				Lives:  20,
				LineID: 0,
			},
			1: &PlayerState{
				ID:     1,
				Lives:  20,
				LineID: 1,
			},
		},
	})

	return ps
}

func (ps *PlayersStore) GetPlayerByID(id int) PlayerState {
	p, _ := ps.GetState().(PlayersState).Players[id]
	return *p
}

func (ps *PlayersStore) GetByLineID(lid int) PlayerState {
	for _, p := range ps.GetState().(PlayersState).Players {
		if p.LineID == lid {
			return *p
		}
	}
	return PlayerState{}
}

func (ps *PlayersStore) GetCurrentPlayer() PlayerState {
	return ps.GetPlayerByID(ps.CurrentPlayerID)
}

func (ps *PlayersStore) Update() error {
	return nil
}

func (ps *PlayersStore) Draw(screen *ebiten.Image) {
}

func (ps *PlayersStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	pstate, ok := state.(PlayersState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.StealLive:
		fp := pstate.Players[act.StealLive.FromPlayerID]
		fp.Lives -= 1

		tp := pstate.Players[act.StealLive.ToPlayerID]
		tp.Lives += 1
	default:
	}
	return pstate
}
