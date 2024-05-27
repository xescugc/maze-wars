package store

import (
	"math"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/unit"
	"github.com/xescugc/maze-wars/unit/environment"
	"github.com/xescugc/maze-wars/utils"
	"github.com/xescugc/maze-wars/utils/graph"
)

const (
	atScale = true

	incomeTimer = 15

	updateFactor     = 0.1
	updateCostFactor = 5
	incomeFactor     = 5
)

var (
	tpsMS = (time.Second / 60).Milliseconds()
)

type Lines struct {
	*flux.ReduceStore

	store *Store

	mxLines sync.RWMutex
}

type LinesState struct {
	Lines   map[int]*Line
	Players map[string]*Player

	// IncomeTimer is the internal counter that goes from 15 to 0
	IncomeTimer int
}

type Line struct {
	Towers map[string]*Tower
	Units  map[string]*Unit

	Graph *graph.Graph

	// UpdatedAt is the last time
	// something was updated on this Line.
	// Towers added, Units added or
	// when the Units position was updated
	// the last time.
	// Used for the SyncState to know how much
	// time has passed since the last update
	// and move the Units accordingly
	// (60 moves per second pass)
	UpdatedAt time.Time
}

type Tower struct {
	utils.Object

	ID string

	Type     string
	LineID   int
	PlayerID string

	Stats tower.Stats

	Level int
}

func (t *Tower) FacetKey() string { return tower.Towers[t.Type].FacesetKey() }
func (t *Tower) CanTarget(env environment.Environment) bool {
	return tower.Towers[t.Type].CanTarget(env)
}

type Unit struct {
	utils.MovingObject

	ID            string
	Type          string
	PlayerID      string
	PlayerLineID  int
	CurrentLineID int

	Health        float64
	MovementSpeed float64

	// The current level of the unit from the PlayerID
	Level int

	Path     []graph.Step
	HashPath string

	// CreatedAt has the time of creation so
	// on the next SyncState will be moved just
	// the diff amount and then it'll be set to 'nil'
	// so we know it's on sync
	CreatedAt time.Time
}

func (u *Unit) FacesetKey() string { return unit.Units[u.Type].FacesetKey() }
func (u *Unit) WalkKey() string    { return unit.Units[u.Type].WalkKey() }

type Player struct {
	ID      string
	Name    string
	Lives   int
	LineID  int
	Income  int
	Gold    int
	Current bool
	Winner  bool

	// UnitUpdates holds the current unit level
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

func NewLines(d *flux.Dispatcher, s *Store) *Lines {
	l := &Lines{
		store: s,
	}

	l.ReduceStore = flux.NewReduceStore(d, l.Reduce, LinesState{
		Lines:       make(map[int]*Line),
		Players:     make(map[string]*Player),
		IncomeTimer: incomeTimer,
	})

	return l
}

func (ls *Lines) ListLines() []*Line {
	ls.mxLines.RLock()
	defer ls.mxLines.RUnlock()

	mlines := ls.GetState().(LinesState)
	lines := make([]*Line, 0, len(mlines.Lines))
	for _, l := range mlines.Lines {
		us := make(map[string]*Unit)
		ts := make(map[string]*Tower)
		for uid, u := range l.Units {
			us[uid] = u
		}
		for tid, t := range l.Towers {
			ts[tid] = t
		}
		ll := *l
		ll.Units = us
		ll.Towers = ts
		lines = append(lines, &ll)
	}
	return lines
}

func (ls *Lines) FindLineByID(id int) *Line {
	ls.mxLines.RLock()
	defer ls.mxLines.RUnlock()

	l := ls.GetState().(LinesState).Lines[id]

	us := make(map[string]*Unit)
	ts := make(map[string]*Tower)
	for uid, u := range l.Units {
		us[uid] = u
	}
	for tid, t := range l.Towers {
		ts[tid] = t
	}
	ll := *l
	ll.Units = us
	ll.Towers = ts

	return &ll
}

// ListPlayers returns the players list and it's meant for reading only purposes
func (ls *Lines) ListPlayers() []*Player {
	ls.mxLines.RLock()
	defer ls.mxLines.RUnlock()

	mlines := ls.GetState().(LinesState)
	players := make([]*Player, 0, len(mlines.Players))
	for _, p := range mlines.Players {
		players = append(players, p)
	}
	return players
}

func (ls *Lines) FindCurrentPlayer() Player {
	ls.mxLines.RLock()
	defer ls.mxLines.RUnlock()
	for _, p := range ls.GetState().(LinesState).Players {
		if p.Current {
			return *p
		}
	}
	return Player{}
}

func (ls *Lines) FindPlayerByID(id string) Player {
	ls.mxLines.RLock()
	defer ls.mxLines.RUnlock()
	p, ok := ls.GetState().(LinesState).Players[id]
	if !ok {
		return Player{}
	}
	return *p
}

func (ls *Lines) FindPlayerByLineID(lid int) Player {
	ls.mxLines.RLock()
	defer ls.mxLines.RUnlock()

	return ls.findPlayerByLineID(lid)
}

func (ls *Lines) findPlayerByLineID(lid int) Player {
	for _, p := range ls.GetState().(LinesState).Players {
		if p.LineID == lid {
			return *p
		}
	}
	return Player{}
}

func (ls *Lines) GetIncomeTimer() int {
	ls.mxLines.RLock()
	defer ls.mxLines.RUnlock()

	lsstate := ls.GetState().(LinesState)
	return lsstate.IncomeTimer
}

func (ls *Lines) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	lstate, ok := state.(LinesState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.IncomeTick:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		lstate.IncomeTimer -= 1
		if lstate.IncomeTimer == 0 {
			lstate.IncomeTimer = incomeTimer
			for _, p := range lstate.Players {
				p.Gold += p.Income
			}
		}
	case action.AddPlayer:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		var found bool
		for _, p := range lstate.Players {
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

		lstate.Players[act.AddPlayer.ID] = p
	case action.StartGame:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		for _, p := range lstate.Players {
			lstate.Lines[p.LineID] = ls.newLine(p.LineID)
		}
	case action.PlaceTower:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		p := lstate.Players[act.PlaceTower.PlayerID]

		if !p.CanPlaceTower(act.PlaceTower.Type) {
			break
		}

		p.Gold -= tower.Towers[act.PlaceTower.Type].Gold

		var w, h int = 16 * 2, 16 * 2
		tid := uuid.Must(uuid.NewV4())
		tw := &Tower{
			ID: tid.String(),
			Object: utils.Object{
				X: float64(act.PlaceTower.X), Y: float64(act.PlaceTower.Y),
				W: w, H: h,
			},
			Type:     act.PlaceTower.Type,
			LineID:   p.LineID,
			PlayerID: p.ID,
			Level:    1,
			Stats:    tower.Towers[act.PlaceTower.Type].Stats,
		}

		l := lstate.Lines[p.LineID]
		// TODO: Check this errors
		_ = l.Graph.AddTower(tw.ID, act.PlaceTower.X, act.PlaceTower.Y, tw.W, tw.H)

		l.Towers[tw.ID] = tw

		ls.recalculateLineUnitSteps(lstate, p.LineID)
	case action.UpdateTower:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		p := lstate.Players[act.UpdateTower.PlayerID]
		l := lstate.Lines[p.LineID]
		t := l.Towers[act.UpdateTower.TowerID]

		tu := tower.FindUpdateByLevel(t.Type, t.Level+1)
		if tu != nil {
			if p.Gold-tu.UpdateCost > 0 {
				t.Stats = tu.Stats
				t.Level += 1
				p.Gold -= tu.UpdateCost
			}
		}

	case action.RemoveTower:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		p := lstate.Players[act.RemoveTower.PlayerID]
		l := lstate.Lines[p.LineID]
		t := l.Towers[act.RemoveTower.TowerID]

		removeTowerGoldReturn := tower.Towers[t.Type].Gold / 2
		tc := tower.FindUpdateByLevel(t.Type, t.Level)
		if tc != nil {
			removeTowerGoldReturn = tc.UpdateCost / 2
		}
		lstate.Players[act.RemoveTower.PlayerID].Gold += removeTowerGoldReturn

		// TODO: Add the LineID
		for lid, l := range lstate.Lines {
			if ok := l.Graph.RemoveTower(act.RemoveTower.TowerID); ok {
				delete(l.Towers, act.RemoveTower.TowerID)
				ls.recalculateLineUnitSteps(lstate, lid)
			}
		}
	case action.SummonUnit:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		cp := lstate.Players[act.SummonUnit.PlayerID]
		if !cp.CanSummonUnit(act.SummonUnit.Type) {
			break
		}
		lstate.Players[act.SummonUnit.PlayerID].Income += cp.UnitUpdates[act.SummonUnit.Type].Current.Income
		lstate.Players[act.SummonUnit.PlayerID].Gold -= cp.UnitUpdates[act.SummonUnit.Type].Current.Gold

		uu := cp.UnitUpdates[act.SummonUnit.Type]
		bu := unit.Units[act.SummonUnit.Type]

		l := lstate.Lines[act.SummonUnit.CurrentLineID]

		var w, h int = 16, 16
		var n = l.Graph.GetRandomSpawnNode()
		uid := uuid.Must(uuid.NewV4())
		u := &Unit{
			MovingObject: utils.MovingObject{
				Object: utils.Object{
					X: float64(n.X), Y: float64(n.Y),
					W: w, H: h,
				},
				Facing: utils.Down,
			},
			ID:            uid.String(),
			Type:          act.SummonUnit.Type,
			PlayerID:      act.SummonUnit.PlayerID,
			PlayerLineID:  act.SummonUnit.PlayerLineID,
			CurrentLineID: act.SummonUnit.CurrentLineID,
			Health:        float64(uu.Current.Health),
			Level:         uu.Level,
			MovementSpeed: uu.Current.MovementSpeed,
			CreatedAt:     time.Now(),
		}

		u.Path = l.Graph.AStar(u.X, u.Y, u.MovementSpeed, u.Facing, l.Graph.DeathNode.X, l.Graph.DeathNode.Y, bu.Environment, atScale)
		u.HashPath = graph.HashSteps(u.Path)
		l.Units[u.ID] = u
	case action.TPS:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		for lid := range lstate.Lines {
			ls.moveLineUnitsTo(lstate, lid, act.TPS.Time)
		}
	case action.RemovePlayer:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		p := lstate.Players[act.RemovePlayer.ID]

		// TODO: Add LineID
		delete(lstate.Lines, p.LineID)

		for _, l := range lstate.Lines {
			for _, u := range l.Units {
				if u.PlayerID == act.RemovePlayer.ID {
					delete(l.Units, u.ID)
				}
			}
		}

		delete(lstate.Players, act.RemovePlayer.ID)
		if len(lstate.Players) == 1 {
			for _, p := range lstate.Players {
				// As there is only 1 we can do it this way
				p.Winner = true
			}
		}

	case action.UpdateUnit:
		u := unit.Units[act.UpdateUnit.Type]
		buu := lstate.Players[act.UpdateUnit.PlayerID].UnitUpdates[act.UpdateUnit.Type]

		if !lstate.Players[act.UpdateUnit.PlayerID].CanUpdateUnit(act.UpdateUnit.Type) {
			break
		}

		lstate.Players[act.UpdateUnit.PlayerID].Gold -= buu.UpdateCost
		lstate.Players[act.UpdateUnit.PlayerID].UnitUpdates[act.UpdateUnit.Type] = UnitUpdate{
			Current:    buu.Next,
			Level:      buu.Level + 1,
			UpdateCost: updateCostFactor * buu.Next.Gold,
			Next:       unitUpdate(buu.Level+2, u.Type.String(), u.Stats),
		}
		//lstate.PlayerID[act.UnitUpdate.PlayerID].UnitUpdate[act.UnitUpdate.Type].

	case action.SyncState:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		// Sync Lines
		for lid, l := range act.SyncState.Lines.Lines {
			cl, ok := lstate.Lines[lid]
			if !ok {
				cl = ls.newLine(lid)
				lstate.Lines[lid] = cl
			}

			// Units
			uids := make(map[string]struct{})
			for id := range cl.Units {
				uids[id] = struct{}{}
			}
			for id, u := range l.Units {
				delete(uids, id)
				nu := Unit(*u)
				ou, ok := cl.Units[id]

				if ok {
					_ = ou
					//If the unit already exists and have the same Hash then ignore the server
					//coordinates and path
					if ou.HashPath == nu.HashPath {
						nu.Path = ou.Path
						nu.X = ou.X
						nu.Y = ou.Y
					}
				}
				cl.Units[id] = &nu

			}
			for id := range uids {
				delete(cl.Units, id)
			}

			// Towers
			tids := make(map[string]struct{})
			for id := range cl.Towers {
				tids[id] = struct{}{}
			}
			atws := make(map[string]struct{})
			for id, t := range l.Towers {
				if _, ok := tids[id]; !ok {
					atws[id] = struct{}{}
				}
				delete(tids, id)
				nt := Tower(*t)
				cl.Towers[id] = &nt
			}
			for id := range tids {
				cl.Graph.RemoveTower(id)
				delete(cl.Towers, id)
			}
			for id := range atws {
				t := cl.Towers[id]
				cl.Graph.AddTower(id, int(t.X), int(t.Y), t.W, t.H)
			}
		}

		// Sync Players
		pids := make(map[string]struct{})
		for id := range lstate.Players {
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
			lstate.Players[id] = &np
		}
		for id := range pids {
			delete(lstate.Players, id)
		}
		lstate.IncomeTimer = act.SyncState.Players.IncomeTimer
	}

	return lstate
}

func (ls *Lines) recalculateLineUnitSteps(lstate LinesState, lid int) {
	t := time.Now()
	ls.moveLineUnitsTo(lstate, lid, t)

	l := lstate.Lines[lid]
	for _, u := range l.Units {
		u.Path = l.Graph.AStar(u.X, u.Y, u.MovementSpeed, u.Facing, l.Graph.DeathNode.X, l.Graph.DeathNode.Y, unit.Units[u.Type].Environment, atScale)
		u.HashPath = graph.HashSteps(u.Path)
	}
}

func (ls *Lines) moveLineUnitsTo(lstate LinesState, lid int, t time.Time) {
	l := lstate.Lines[lid]
	lmoves := 1
	if !t.IsZero() && !l.UpdatedAt.IsZero() {
		lmoves = int(t.Sub(l.UpdatedAt).Milliseconds() / tpsMS)
	}
	// We'll move all the Units 1 by 1 so we can calculate if they have
	// an aura around and if they have been attacked/killed and if they
	// reached the end to steal a live and change lines
	for i := 1; i < lmoves+1; i++ {
		for _, u := range l.Units {
			// If a unit was added in between the TPS checks
			// we only need to move partially
			if !t.IsZero() && !u.CreatedAt.IsZero() {
				um := int(t.Sub(u.CreatedAt).Milliseconds() / tpsMS)
				// We only want to consider this unit move when
				// it should do it
				// If for some reason it gets desynchronized even just for
				// 1 ms then the unit it'll be too old to ever get moved
				// so if the expected time to move is too big we just move it
				if um != (lmoves-i) && um < lmoves {
					continue
				}
				// This way we mean it's up to date now
				u.CreatedAt = time.Time{}
			}
			// TODO: Investigate why this is a case to check
			// as if it's 0 it should be read for the next if
			// and delete/change line
			if len(u.Path) != 0 {
				nextStep := u.Path[0]
				u.Path = u.Path[1:]
				u.MovingCount += 1
				u.Y = nextStep.Y
				u.X = nextStep.X
				u.Facing = nextStep.Facing
			}

			// We reached the end of the line
			if len(u.Path) == 0 {
				var fpID string
				for pid, p := range lstate.Players {
					if p.LineID == lid {
						fpID = pid
						break
					}
				}
				// We steal a Live
				ls.stealLive(lstate, fpID, u.PlayerID)
				nlid := ls.store.Map.GetNextLineID(u.CurrentLineID)
				// If the next line is the owner of the Unit we remove it
				// If not then we change the unit to the next line
				if nlid == u.PlayerLineID {
					delete(lstate.Lines[lid].Units, u.ID)
				} else {
					ls.changeUnitLine(lstate, u, nlid)
				}
			}
		}
		// Now that the unit has moved we'll calculate if any
		// tower can attack any Unit in their new positions
		for _, t := range l.Towers {
			// Get the closes unit to the current tower to attack it
			var (
				minDist     float64 = 0
				minDistUnit *Unit
			)
			for _, u := range l.Units {
				if !t.CanTarget(unit.Units[u.Type].Environment) {
					continue
				}
				d := t.PDistance(u.Object)
				if minDist == 0 {
					minDist = d
				}
				if d <= tower.Towers[t.Type].Range && d <= minDist {
					minDist = d
					minDistUnit = u
				}
			}
			if minDistUnit != nil {
				// Tower Attack
				minDistUnit.Health -= t.Stats.Damage
				if minDistUnit.Health <= 0 {
					minDistUnit.Health = 0
					// Delete Unit
					delete(lstate.Lines[lid].Units, minDistUnit.ID)

					// Unit Killed by player so we give gold to the player
					cp := lstate.Players[t.PlayerID]
					cp.Gold += unitUpdate(minDistUnit.Level, minDistUnit.Type, unit.Units[minDistUnit.Type].Stats).Income
				}
			}
		}
	}
	l.UpdatedAt = t
}

func (ls *Lines) newLine(lid int) *Line {
	x, y := ls.store.Map.GetHomeCoordinates(lid)
	g, err := graph.New(x+16, y+16, 16, 84, 16, 7, 74, 3)
	if err != nil {
		panic(err)
	}
	return &Line{
		Towers: make(map[string]*Tower),
		Units:  make(map[string]*Unit),
		Graph:  g,
	}
}

func unitUpdate(nlvl int, ut string, u unit.Stats) unit.Stats {
	bu := unit.Units[ut]

	u.Health = float64(levelToValue(nlvl, int(bu.Health)))
	u.Gold = levelToValue(nlvl, bu.Gold)
	u.Income = int(math.Round(float64(u.Gold) / float64(incomeFactor)))

	return u
}

func levelToValue(lvl, base int) int {
	fb := float64(base)
	for i := 1; i < lvl; i++ {
		fb = math.Round(fb) * math.Pow(math.E, updateFactor)
	}
	return int(math.Round(fb))
}

func (ls *Lines) stealLive(lstate LinesState, fpID, tpID string) {
	fp := lstate.Players[fpID]
	tp := lstate.Players[tpID]

	fp.Lives -= 1
	if fp.Lives < 0 {
		fp.Lives = 0
	} else {
		tp.Lives += 1
	}

	var stillPlayersLeft bool
	for _, p := range lstate.Players {
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
}

func (ls *Lines) changeUnitLine(lstate LinesState, u *Unit, nlid int) {
	cl := lstate.Lines[u.CurrentLineID]
	// As we are gonna move it to another line
	// we remove it
	delete(cl.Units, u.ID)

	u.CurrentLineID = nlid

	nl := lstate.Lines[u.CurrentLineID]

	n := nl.Graph.GetRandomSpawnNode()
	u.X = float64(n.X)
	u.Y = float64(n.Y)

	u.Path = nl.Graph.AStar(u.X, u.Y, u.MovementSpeed, u.Facing, nl.Graph.DeathNode.X, nl.Graph.DeathNode.Y, unit.Units[u.Type].Environment, atScale)
	u.HashPath = graph.HashSteps(u.Path)

	u.CreatedAt = time.Now()
	nl.Units[u.ID] = u
}
