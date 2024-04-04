package store

import (
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
	Lines map[int]*Line
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

	Health float64

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
func (u *Unit) SpriteKey() string  { return unit.Units[u.Type].SpriteKey() }

func NewLines(d *flux.Dispatcher, s *Store) *Lines {
	l := &Lines{
		store: s,
	}

	l.ReduceStore = flux.NewReduceStore(d, l.Reduce, LinesState{
		Lines: make(map[int]*Line),
	})

	return l
}

func (ls *Lines) List() []*Line {
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

func (ls *Lines) FindByID(id int) *Line {
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
	case action.StartGame:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		for _, p := range ls.store.Players.List() {
			lstate.Lines[p.LineID] = ls.newLine(p.LineID)
		}
	case action.PlaceTower:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		p := ls.store.Players.FindByID(act.PlaceTower.PlayerID)

		if !p.CanPlaceTower(act.PlaceTower.Type) {
			break
		}

		var w, h int = 16 * 2, 16 * 2
		tid := uuid.Must(uuid.NewV4())
		tw := &Tower{
			ID: tid.String(),
			Object: utils.Object{
				X: act.PlaceTower.X, Y: act.PlaceTower.Y,
				W: w, H: h,
			},
			Type:     act.PlaceTower.Type,
			LineID:   p.LineID,
			PlayerID: p.ID,
		}

		l := lstate.Lines[p.LineID]
		// TODO: Check this errors
		_ = l.Graph.AddTower(tw.ID, tw.X, tw.Y, tw.W, tw.H)

		l.Towers[tw.ID] = tw

		recalculateLineUnitSteps(l)
	case action.RemoveTower:
		// TODO: Add the LineID
		for _, l := range lstate.Lines {
			if ok := l.Graph.RemoveTower(act.RemoveTower.TowerID); ok {
				delete(l.Towers, act.RemoveTower.TowerID)
				recalculateLineUnitSteps(l)
			}
		}
	case action.TowerAttack:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		// TODO: Add the LineID
		for _, l := range lstate.Lines {
			if u, ok := l.Units[act.TowerAttack.UnitID]; ok {
				u.Health -= tower.Towers[act.TowerAttack.TowerType].Damage
				if u.Health <= 0 {
					u.Health = 0
				}
				break
			}
		}
	case action.SummonUnit:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		p := ls.store.Players.FindByID(act.SummonUnit.PlayerID)
		if !p.CanSummonUnit(act.SummonUnit.Type) {
			break
		}

		uu := p.UnitUpdates[act.SummonUnit.Type]
		bu := unit.Units[act.SummonUnit.Type]

		l := lstate.Lines[act.SummonUnit.CurrentLineID]

		var w, h int = 16, 16
		var n = l.Graph.GetRandomSpawnNode()
		uid := uuid.Must(uuid.NewV4())
		u := &Unit{
			MovingObject: utils.MovingObject{
				Object: utils.Object{
					X: n.X, Y: n.Y,
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
			CreatedAt:     time.Now(),
		}

		u.Path = l.Graph.AStar(u.X, u.Y, u.Facing, l.Graph.DeathNode.X, l.Graph.DeathNode.Y, bu.Environment, atScale)
		u.HashPath = graph.HashSteps(u.Path)
		l.Units[u.ID] = u
	case action.ChangeUnitLine:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		// TODO: Add the unit LineID
		for lid, l := range lstate.Lines {
			if u, ok := l.Units[act.ChangeUnitLine.UnitID]; ok {
				// As we are gonna move it to another line
				// we remove it
				delete(l.Units, u.ID)

				u.CurrentLineID = ls.store.Map.GetNextLineID(lid)

				nl := lstate.Lines[u.CurrentLineID]

				n := nl.Graph.GetRandomSpawnNode()
				u.X = n.X
				u.Y = n.Y

				u.Path = nl.Graph.AStar(u.X, u.Y, u.Facing, nl.Graph.DeathNode.X, nl.Graph.DeathNode.Y, unit.Units[u.Type].Environment, atScale)
				u.HashPath = graph.HashSteps(u.Path)

				u.CreatedAt = time.Now()
				nl.Units[u.ID] = u

				break
			}
		}
	case action.RemoveUnit:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		// As I don't know which line is it
		// we just remove it from all
		// TODO: Add the LineID on the action
		for _, l := range lstate.Lines {
			delete(l.Units, act.RemoveUnit.UnitID)
		}

	case action.TPS:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		for _, l := range lstate.Lines {
			moveLineUnitsTo(l, act.TPS.Time)
		}
	case action.RemovePlayer:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		p := ls.store.Players.FindByID(act.RemovePlayer.ID)

		// TODO: Add LineID
		delete(lstate.Lines, p.LineID)

		for _, l := range lstate.Lines {
			for _, u := range l.Units {
				if u.PlayerID == act.RemovePlayer.ID {
					delete(l.Units, u.ID)
				}
			}
		}
	case action.SyncState:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

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
				cl.Graph.AddTower(id, t.X, t.Y, t.W, t.H)
			}
		}
	}

	return lstate
}

func recalculateLineUnitSteps(l *Line) {
	t := time.Now()
	moveLineUnitsTo(l, t)

	for _, u := range l.Units {
		u.Path = l.Graph.AStar(u.X, u.Y, u.Facing, l.Graph.DeathNode.X, l.Graph.DeathNode.Y, unit.Units[u.Type].Environment, atScale)
		u.HashPath = graph.HashSteps(u.Path)
	}
}

func moveLineUnitsTo(l *Line, t time.Time) {
	lmoves := 1
	if !t.IsZero() && !l.UpdatedAt.IsZero() {
		lmoves = int(t.Sub(l.UpdatedAt).Milliseconds() / tpsMS)
	}
	for _, u := range l.Units {
		if len(u.Path) > 0 {
			umoves := lmoves
			if !t.IsZero() && !u.CreatedAt.IsZero() {
				umoves = int(t.Sub(u.CreatedAt).Milliseconds() / tpsMS)
				// This way we mean it's up to date now
				u.CreatedAt = time.Time{}
			}
			// If we have less moves remaining that the expected amount
			// we just move to the last position
			if len(u.Path) < umoves {
				umoves = len(u.Path) - 1
			}
			if umoves == 0 {
				continue
			}
			nextStep := u.Path[umoves-1]
			u.Path = u.Path[umoves:]
			u.MovingCount += umoves
			u.Y = nextStep.Y
			u.X = nextStep.X
			u.Facing = nextStep.Facing
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
