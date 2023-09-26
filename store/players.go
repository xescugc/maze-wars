package store

import (
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
)

var (
	unitIncome = map[string]int{
		"cyclope": 1,
	}
	incomeTimer = 15
	unitGold    = 10
	towerGold   = 10
)

type Players struct {
	*flux.ReduceStore
}

type PlayersState struct {
	Players map[string]*Player

	// IncomeTimer is the internal counter that goes from 15 to 0
	IncomeTimer int
}

type Player struct {
	ID      string
	Name    string
	Lives   int
	LineID  int
	Income  int
	Gold    int
	Current bool
}

func NewPlayers(d *flux.Dispatcher) *Players {
	p := &Players{}
	p.ReduceStore = flux.NewReduceStore(d, p.Reduce, PlayersState{
		IncomeTimer: incomeTimer,
		Players:     make(map[string]*Player),
	})

	return p
}

func (ps *Players) GetPlayerByID(id string) Player {
	p, _ := ps.GetState().(PlayersState).Players[id]
	return *p
}

func (ps *Players) GetByLineID(lid int) Player {
	for _, p := range ps.GetState().(PlayersState).Players {
		if p.LineID == lid {
			return *p
		}
	}
	return Player{}
}

func (ps *Players) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	pstate, ok := state.(PlayersState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.AddPlayer:
		pstate.Players[act.AddPlayer.ID] = &Player{
			ID:     act.AddPlayer.ID,
			Name:   act.AddPlayer.Name,
			Lives:  20,
			LineID: act.AddPlayer.LineID,
			Income: 25,
			Gold:   40,
		}
	case action.StealLive:
		fp := pstate.Players[act.StealLive.FromPlayerID]
		fp.Lives -= 1

		tp := pstate.Players[act.StealLive.ToPlayerID]
		tp.Lives += 1
	case action.SummonUnit:
		pstate.Players[act.SummonUnit.PlayerID].Income += unitIncome[act.SummonUnit.Type]
		pstate.Players[act.SummonUnit.PlayerID].Gold -= unitGold
	case action.IncomeTick:
		pstate.IncomeTimer -= 1
		if pstate.IncomeTimer == 0 {
			pstate.IncomeTimer = incomeTimer
			for _, p := range pstate.Players {
				p.Gold += p.Income
			}
		}
	case action.PlaceTower:
		pstate.Players[act.PlaceTower.PlayerID].Gold -= towerGold
	case action.UnitKilled:
		pstate.Players[act.UnitKilled.PlayerID].Gold += unitIncome[act.UnitKilled.UnitType]
	case action.UpdateState:
		for _, p := range act.UpdateState.Players.Players {
			np := Player(*p)
			pstate.Players[p.ID] = &np
		}
		pstate.IncomeTimer = act.UpdateState.Players.IncomeTimer
	}

	return pstate
}
