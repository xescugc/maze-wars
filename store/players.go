package store

import (
	"sync"

	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/unit"
)

const (
	incomeTimer = 15
)

type Players struct {
	*flux.ReduceStore

	store *Store

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
}

func (p Player) CanSummonUnit(ut string) bool {
	return (p.Gold - unit.Units[ut].Gold) >= 0
}
func (p Player) CanPlaceTower(tt string) bool {
	return (p.Gold - tower.Towers[tt].Gold) >= 0
}

func NewPlayers(d *flux.Dispatcher, s *Store) *Players {
	p := &Players{
		store: s,
	}
	p.ReduceStore = flux.NewReduceStore(d, p.Reduce, PlayersState{
		IncomeTimer: incomeTimer,
		Players:     make(map[string]*Player),
	})

	return p
}

// GetPlayers returns the players list and it's meant for reading only purposes
func (ps *Players) List() []*Player {
	ps.mxPlayers.RLock()
	defer ps.mxPlayers.RUnlock()

	mplayers := ps.GetState().(PlayersState)
	players := make([]*Player, 0, len(mplayers.Players))
	for _, p := range mplayers.Players {
		players = append(players, p)
	}
	return players
}

func (ps *Players) FindCurrent() Player {
	ps.mxPlayers.RLock()
	defer ps.mxPlayers.RUnlock()
	for _, p := range ps.GetState().(PlayersState).Players {
		if p.Current {
			return *p
		}
	}
	return Player{}
}

func (ps *Players) FindByID(id string) Player {
	ps.mxPlayers.RLock()
	defer ps.mxPlayers.RUnlock()
	p, ok := ps.GetState().(PlayersState).Players[id]
	if !ok {
		return Player{}
	}
	return *p
}

func (ps *Players) FindByLineID(lid int) Player {
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

		var found bool
		for _, p := range pstate.Players {
			if p.Name == act.AddPlayer.Name {
				found = true
				break
			}
		}

		if found {
			break
		}

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

		if len(pstate.Players) == 1 {
			for _, p := range pstate.Players {
				// As there is only 1 we can do it this way
				p.Winner = true
			}
		}

	case action.StealLive:
		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		fp := pstate.Players[act.StealLive.FromPlayerID]
		tp := pstate.Players[act.StealLive.ToPlayerID]

		fp.Lives -= 1
		if fp.Lives < 0 {
			fp.Lives = 0
		} else {
			tp.Lives += 1
		}

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
		// We need to wait for the units if not the units Store cannot check
		// if the unit can be summoned if the Gold has already been removed
		ps.GetDispatcher().WaitFor(ps.store.Lines.GetDispatcherToken())
		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		if !pstate.Players[act.SummonUnit.PlayerID].CanSummonUnit(act.SummonUnit.Type) {
			break
		}
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
		ps.GetDispatcher().WaitFor(
			ps.store.Lines.GetDispatcherToken(),
		)

		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		if !pstate.Players[act.PlaceTower.PlayerID].CanPlaceTower(act.PlaceTower.Type) {
			break
		}

		pstate.Players[act.PlaceTower.PlayerID].Gold -= tower.Towers[act.PlaceTower.Type].Gold
	case action.RemoveTower:
		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		pstate.Players[act.RemoveTower.PlayerID].Gold += tower.Towers[act.RemoveTower.TowerType].Gold / 2
	case action.UnitKilled:
		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		pstate.Players[act.UnitKilled.PlayerID].Gold += unit.Units[act.UnitKilled.UnitType].Income
	case action.SyncState:
		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		pids := make(map[string]struct{})
		for id := range pstate.Players {
			pids[id] = struct{}{}
		}
		for id, p := range act.SyncState.Players.Players {
			delete(pids, id)
			np := Player(*p)
			pstate.Players[id] = &np
		}
		for id := range pids {
			delete(pstate.Players, id)
		}
		pstate.IncomeTimer = act.SyncState.Players.IncomeTimer
	}

	return pstate
}
