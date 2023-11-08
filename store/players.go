package store

import (
	"sync"

	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
	"github.com/xescugc/ltw/tower"
	"github.com/xescugc/ltw/unit"
)

const (
	incomeTimer = 15
)

type Players struct {
	*flux.ReduceStore

	mxPlayers sync.RWMutex
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
	Winner  bool
	Ready   bool
}

func NewPlayers(d *flux.Dispatcher) *Players {
	p := &Players{}
	p.ReduceStore = flux.NewReduceStore(d, p.Reduce, PlayersState{
		IncomeTimer: incomeTimer,
		Players:     make(map[string]*Player),
	})

	return p
}

// GetPlayers returns the players list and it's meant for reading only purposes
func (ps *Players) GetPlayers() []*Player {
	ps.mxPlayers.RLock()
	defer ps.mxPlayers.RUnlock()
	mplayers := ps.GetState().(PlayersState)
	players := make([]*Player, 0, len(mplayers.Players))
	for _, p := range mplayers.Players {
		players = append(players, p)
	}
	return players
}

func (ps *Players) GetCurrentPlayer() Player {
	ps.mxPlayers.RLock()
	defer ps.mxPlayers.RUnlock()
	for _, p := range ps.GetState().(PlayersState).Players {
		if p.Current {
			return *p
		}
	}
	return Player{}
}

func (ps *Players) GetPlayerByID(id string) Player {
	ps.mxPlayers.RLock()
	defer ps.mxPlayers.RUnlock()
	p, _ := ps.GetState().(PlayersState).Players[id]
	return *p
}

func (ps *Players) GetByLineID(lid int) Player {
	ps.mxPlayers.RLock()
	defer ps.mxPlayers.RUnlock()
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
		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		pstate.Players[act.AddPlayer.ID] = &Player{
			ID:     act.AddPlayer.ID,
			Name:   act.AddPlayer.Name,
			Lives:  20,
			LineID: act.AddPlayer.LineID,
			Income: 25,
			Gold:   40,
		}
	case action.RemovePlayer:
		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		delete(pstate.Players, act.RemovePlayer.ID)
	case action.StealLive:
		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		fp := pstate.Players[act.StealLive.FromPlayerID]
		fp.Lives -= 1
		if fp.Lives < 0 {
			fp.Lives = 0
		}

		tp := pstate.Players[act.StealLive.ToPlayerID]
		tp.Lives += 1

		var stillPlayersLeft bool
		for _, p := range pstate.Players {
			if stillPlayersLeft {
				continue
			}
			if p.Lives != 0 && p.ID != tp.ID {
				stillPlayersLeft = true
			}
		}

		if !stillPlayersLeft {
			tp.Winner = true
		}
	case action.SummonUnit:
		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		pstate.Players[act.SummonUnit.PlayerID].Income += unit.Units[act.SummonUnit.Type].Income
		pstate.Players[act.SummonUnit.PlayerID].Gold -= unit.Units[act.SummonUnit.Type].Gold
	case action.IncomeTick:
		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		pstate.IncomeTimer -= 1
		if pstate.IncomeTimer == 0 {
			pstate.IncomeTimer = incomeTimer
			for _, p := range pstate.Players {
				p.Gold += p.Income
			}
		}
	case action.PlaceTower:
		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		pstate.Players[act.PlaceTower.PlayerID].Gold -= tower.Towers[act.PlaceTower.Type].Gold
	case action.RemoveTower:
		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		pstate.Players[act.RemoveTower.PlayerID].Gold += tower.Towers[act.RemoveTower.TowerType].Gold / 2
	case action.PlayerReady:
		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		pstate.Players[act.PlayerReady.ID].Ready = true
	case action.UnitKilled:
		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		pstate.Players[act.UnitKilled.PlayerID].Gold += unit.Units[act.UnitKilled.UnitType].Income
	case action.UpdateState:
		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		pids := make(map[string]struct{})
		for id := range pstate.Players {
			pids[id] = struct{}{}
		}
		for id, p := range act.UpdateState.Players.Players {
			delete(pids, id)
			np := Player(*p)
			pstate.Players[id] = &np
		}
		for id := range pids {
			delete(pstate.Players, id)
		}
		pstate.IncomeTimer = act.UpdateState.Players.IncomeTimer
	}

	return pstate
}
