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

	// It's 10% so we use 1.1 to get the increment
	updateFactor = 1.1

	// It's 100% more
	updateCostFactor = 100
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

	// UnitUpdates holds the current unit version
	UnitUpdates map[string]UnitUpdate
}

type UnitUpdate struct {
	// Current is the current unit
	Current unit.Stats

	// Level is the number of the unit level
	// which is basically the number of times
	// it has been updated
	Level int

	UpdateCost int

	// Is how the unit will look after the next update
	Next unit.Stats
}

func (p Player) CanSummonUnit(ut string) bool {
	return (p.Gold - unit.Units[ut].Gold) >= 0
}
func (p Player) CanUpdateUnit(ut string) bool {
	return (p.Gold - p.UnitUpdates[ut].UpdateCost) >= 0
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

		p := &Player{
			ID:     act.AddPlayer.ID,
			Name:   act.AddPlayer.Name,
			Lives:  20,
			LineID: act.AddPlayer.LineID,
			Income: 25,
			Gold:   40,

			UnitUpdates: make(map[string]UnitUpdate),
		}
		for _, u := range unit.Units {
			p.UnitUpdates[u.Type.String()] = UnitUpdate{
				Current:    u.Stats,
				Level:      1,
				UpdateCost: updateCostFactor * u.Gold,
				Next:       unitUpdate(2, u.Type.String(), u.Stats),
			}
		}

		pstate.Players[act.AddPlayer.ID] = p
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

		cp := pstate.Players[act.SummonUnit.PlayerID]
		if !cp.CanSummonUnit(act.SummonUnit.Type) {
			break
		}
		pstate.Players[act.SummonUnit.PlayerID].Income += cp.UnitUpdates[act.SummonUnit.Type].Current.Income
		pstate.Players[act.SummonUnit.PlayerID].Gold -= cp.UnitUpdates[act.SummonUnit.Type].Current.Gold
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

		cp := pstate.Players[act.UnitKilled.PlayerID]
		u := ps.store.Lines.FindByID(cp.LineID).Units[act.UnitKilled.UnitID]

		cp.Gold += unitUpdate(u.Level, u.Type, unit.Units[u.Type].Stats).Income
	case action.UpdateUnit:
		u := unit.Units[act.UpdateUnit.Type]
		buu := pstate.Players[act.UpdateUnit.PlayerID].UnitUpdates[act.UpdateUnit.Type]

		if !pstate.Players[act.UpdateUnit.PlayerID].CanUpdateUnit(act.UpdateUnit.Type) {
			break
		}

		pstate.Players[act.UpdateUnit.PlayerID].Gold -= buu.UpdateCost
		pstate.Players[act.UpdateUnit.PlayerID].UnitUpdates[act.UpdateUnit.Type] = UnitUpdate{
			Current:    buu.Next,
			Level:      buu.Level + 1,
			UpdateCost: updateCostFactor * buu.Next.Gold,
			Next:       unitUpdate(buu.Level+2, u.Type.String(), u.Stats),
		}

	case action.SyncState:
		ps.mxPlayers.Lock()
		defer ps.mxPlayers.Unlock()

		pids := make(map[string]struct{})
		for id := range pstate.Players {
			pids[id] = struct{}{}
		}
		for id, p := range act.SyncState.Players.Players {
			delete(pids, id)
			np := Player{
				ID:          p.ID,
				Name:        p.Name,
				Lives:       p.Lives,
				LineID:      p.LineID,
				Income:      p.Income,
				Gold:        p.Gold,
				Current:     p.Current,
				Winner:      p.Winner,
				UnitUpdates: make(map[string]UnitUpdate),
			}
			for t, uu := range p.UnitUpdates {
				np.UnitUpdates[t] = UnitUpdate(uu)
			}
			pstate.Players[id] = &np
		}
		for id := range pids {
			delete(pstate.Players, id)
		}
		pstate.IncomeTimer = act.SyncState.Players.IncomeTimer
	}

	return pstate
}

func unitUpdate(nlvl int, ut string, u unit.Stats) unit.Stats {
	bu := unit.Units[ut]

	u.Health = float64(levelToValue(nlvl, int(bu.Health)))
	u.Gold = levelToValue(nlvl, bu.Gold)
	u.Income = levelToValue(nlvl, bu.Income)

	return u
}

func levelToValue(lvl, base int) int {
	fb := float64(base)
	for i := 1; i < lvl; i++ {
		fb = fb * updateFactor
	}
	return int(fb)
}
