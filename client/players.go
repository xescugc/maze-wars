package main

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
)

var (
	unitIncome = map[string]int{
		"cyclope": 1,
	}
	incomeTimer = 15
)

type PlayersStore struct {
	*flux.ReduceStore
	CurrentPlayerID int
}

type PlayersState struct {
	Players map[int]*PlayerState

	// IncomeTimer is the internal counter that goes from 15 to 0
	IncomeTimer int
}

type PlayerState struct {
	ID     int
	Lives  int
	LineID int
	Income int
	Gold   int
}

func NewPlayersStore(d *flux.Dispatcher) *PlayersStore {
	// TODO: We fake 2 users and we are the 0
	ps := &PlayersStore{
		CurrentPlayerID: 0,
	}
	ps.ReduceStore = flux.NewReduceStore(d, ps.Reduce, PlayersState{
		IncomeTimer: incomeTimer,
		Players: map[int]*PlayerState{
			0: &PlayerState{
				ID:     0,
				Lives:  20,
				LineID: 0,
				Income: 25,
				Gold:   40,
			},
			1: &PlayerState{
				ID:     1,
				Lives:  20,
				LineID: 1,
				Income: 25,
				Gold:   40,
			},
		},
	})

	go func() {
		t := time.NewTicker(time.Second)
		for {
			select {
			case <-t.C:
				actionDispatcher.IncomeTick()
			}
		}
	}()

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
	case action.SummonUnit:
		pstate.Players[act.SummonUnit.PlayerID].Income += unitIncome[act.SummonUnit.Type]
	case action.IncomeTick:
		pstate.IncomeTimer -= 1
		if pstate.IncomeTimer == 0 {
			pstate.IncomeTimer = incomeTimer
			for _, p := range pstate.Players {
				p.Gold += p.Income
			}
		}
	case action.UnitKilled:
		pstate.Players[act.UnitKilled.PlayerID].Gold += unitIncome[act.UnitKilled.UnitType]
	default:
	}
	return pstate
}
