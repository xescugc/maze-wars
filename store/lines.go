package store

import (
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/unit"
	"github.com/xescugc/maze-wars/unit/ability"
	"github.com/xescugc/maze-wars/unit/buff"
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

	resurrectionRank1 = 0.25

	noTowerID = ""
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
	ID     int
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

func (l *Line) ListSortedUnits() []*Unit {
	res := make([]*Unit, 0, len(l.Units))
	for _, u := range l.Units {
		res = append(res, u)
	}
	sort.Slice(res, func(i, j int) bool { return res[i].ID > res[j].ID })
	return res
}

type Tower struct {
	utils.Object

	ID string

	Type     string
	LineID   int
	PlayerID string

	Health float64

	TargetUnitID string
	LastAttack   time.Time
}

func (t *Tower) FacetKey() string { return tower.Towers[t.Type].FacesetKey() }
func (t *Tower) CanTarget(env environment.Environment) bool {
	return tower.Towers[t.Type].CanTarget(env)
}
func (t *Tower) CanUpdateTo(tt string) bool {
	for _, u := range tower.Towers[t.Type].Updates {
		if u.String() == tt {
			return true
		}
	}
	return false
}

func (t *Tower) CanAttackUnit(u *Unit) bool {
	if !t.CanTarget(unit.Units[u.Type].Environment) || u.HasBuff(buff.Burrowoed) || u.HasBuff(buff.Resurrecting) {
		return false
	}

	// If we do not take the center of the tower, towers would calculate everything from the top right
	// which is not good as a short range tower would not be able to attack from the right for example.
	// We add 16 as it's the distance from the center of the tower to the edge of it so the range ignores that part
	centerTower := utils.Object{X: t.X - 16, Y: t.Y - 16}
	return u.Object.IsCollidingCircle(centerTower, tower.Towers[t.Type].Range*32+16)
}

func (t *Tower) CanAttack(tm time.Time) bool {
	return tm.Sub(t.LastAttack) > time.Duration(int(tower.Towers[t.Type].AttackSpeed*float64(time.Second)))
}

type Unit struct {
	utils.MovingObject
	AnimationCount int

	ID            string
	Type          string
	PlayerID      string
	PlayerLineID  int
	CurrentLineID int

	MaxHealth float64
	Health    float64

	MaxShield float64
	Shield    float64

	MovementSpeed float64
	Bounty        int

	// The current level of the unit from the PlayerID
	Level int

	Path     []graph.Step
	HashPath string

	// CreatedAt has the time of creation so
	// on the next SyncState will be moved just
	// the diff amount and then it'll be set to 'nil'
	// so we know it's on sync
	CreatedAt time.Time

	// Abilities stores data from the abilities that
	// the unit has and need to be kept in check, for example
	// if it's a slime which is the other unit and if it died
	// then give the bounty. Or if the unit already resurected
	// The key is an ability.Ability.String() and the value is
	// a type that is specific for the ability.
	Abilities map[string]interface{}
	Buffs     map[string]interface{}

	// If the Unit has the ability 'Attack' it'll have a
	// TargetTowerID if it has a Tower to attack
	TargetTowerID string
	LastAttack    time.Time
}

func (u *Unit) FacesetKey() string                { return unit.Units[u.Type].FacesetKey() }
func (u *Unit) WalkKey() string                   { return unit.Units[u.Type].WalkKey() }
func (u *Unit) AttackKey() string                 { return unit.Units[u.Type].AttackKey() }
func (u *Unit) IdleKey() string                   { return unit.Units[u.Type].IdleKey() }
func (u *Unit) HasAbility(a ability.Ability) bool { return unit.Units[u.Type].HasAbility(a) }

func (u *Unit) AddBuff(b buff.Buff) {
	if u.Buffs == nil {
		u.Buffs = make(map[string]interface{})
	}
	u.Buffs[b.String()] = nil
}
func (u *Unit) HasBuff(b buff.Buff) bool { _, ok := u.Buffs[b.String()]; return ok }
func (u *Unit) RemoveBuff(b buff.Buff) {
	delete(u.Buffs, b.String())
}
func (u *Unit) CanBeAttacked(t time.Time) bool {
	if u.HasBuff(buff.Burrowoed) {
		if !u.MustUnburrow(t) {
			u.AnimationCount += 1
			return false
		}
		u.Unburrow()
	} else if u.HasBuff(buff.Resurrecting) {
		if !u.CanResurrect(t) {
			return false
		}
		u.Resurrect()
	}
	return true
}

func (u *Unit) TakeDamage(d float64) {
	if u.Shield != 0 {
		u.Shield -= d
		if u.Shield < 0 {
			u.Shield = 0
		}
	} else {
		u.Health -= d
	}
}
func (u *Unit) MustUnburrow(t time.Time) bool {
	// As it should have the buff but it does not we just
	// say yes
	if u.Abilities == nil {
		return true
	}
	amb, ok := u.Abilities[ability.Burrow.String()]
	if !ok {
		return true
	}
	ab, ok := amb.(AbilityBurrow)
	if !ok {
		return true
	}
	return ab.MustUnburrow(t)
}

func (u *Unit) CanUnburrow(t time.Time) bool {
	// As it should have the buff but it does not we just
	// say yes
	if u.Abilities == nil {
		return true
	}
	amb, ok := u.Abilities[ability.Burrow.String()]
	if !ok {
		return true
	}
	ab, ok := amb.(AbilityBurrow)
	if !ok {
		return true
	}
	return ab.CanUnburrow(t)
}

func (u *Unit) Unburrow() {
	u.RemoveBuff(buff.Burrowoed)
	amb, ok := u.Abilities[ability.Burrow.String()]
	if !ok {
		return
	}

	ab, ok := amb.(AbilityBurrow)
	if !ok {
		return
	}

	ab.Unburrowed = true
	u.Abilities[ability.Burrow.String()] = ab
	u.AnimationCount = 0
}

func (u *Unit) WasBurrowed() bool {
	// As it should have the buff but it does not we just
	// say yes
	if u.Abilities == nil {
		return false
	}
	amb, ok := u.Abilities[ability.Burrow.String()]
	if !ok {
		return false
	}
	ab, ok := amb.(AbilityBurrow)
	if !ok {
		return false
	}
	return ab.Unburrowed
}

func (u *Unit) CanResurrect(t time.Time) bool {
	// As it should have the buff but it does not we just
	// say yes
	if u.Abilities == nil {
		return true
	}
	amb, ok := u.Abilities[ability.Resurrection.String()]
	if !ok {
		return true
	}
	ab, ok := amb.(AbilityResurrection)
	if !ok {
		return true
	}
	return ab.CanResurrect(t)
}

func (u *Unit) Resurrect() {
	u.RemoveBuff(buff.Resurrecting)
	ar, ok := u.Abilities[ability.Resurrection.String()]
	if !ok {
		return
	}

	ab, ok := ar.(AbilityResurrection)
	if !ok {
		return
	}

	ab.Resurrected = true
	u.Abilities[ability.Resurrection.String()] = ab
	u.AnimationCount = 0
	// TODO: Figure out when to use the other ranks
	u.Health = u.MaxHealth * resurrectionRank1
}

func (u *Unit) WasResurrected() bool {
	// As it should have the buff but it does not we just
	// say yes
	if u.Abilities == nil {
		return false
	}
	amb, ok := u.Abilities[ability.Resurrection.String()]
	if !ok {
		return false
	}
	ab, ok := amb.(AbilityResurrection)
	if !ok {
		return false
	}
	return ab.Resurrected
}

func (u *Unit) Hybrid(cp, op int) {
	if cp < op {
		return
	}
	// This is the Percentage Difference
	// TODO: Potentially show this as a Buff
	p := float64(((cp - op) / ((cp + op) / 2)) * 100)
	bu := unit.Units[u.Type]
	uu := unitUpdate(u.Level, u.Type, bu.Stats)

	var (
		hp float64 = 1
		sp float64 = 1
	)
	if u.Shield != u.MaxShield {
		hp = (u.Health / u.MaxHealth)
		sp = (u.Shield / u.MaxShield)
	}
	// base + base * % increase / 100
	u.MaxHealth = (uu.Health + u.Health*p/100) * hp
	u.Health = u.MaxHealth * hp
	u.MovementSpeed = uu.MovementSpeed + uu.MovementSpeed*p/100
	u.MaxShield = uu.Shield + uu.Shield*p/100
	u.Shield = u.MaxShield * sp
}

func (u *Unit) CanAttack(tm time.Time) bool {
	if !u.HasAbility(ability.Attack) {
		return false
	}
	return tm.Sub(u.LastAttack) > time.Duration(int(unit.Units[u.Type].AttackSpeed*float64(time.Second)))
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
	return (p.Gold - p.UnitUpdates[ut].Current.Gold) >= 0
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

		var w, h int = 16 * 2, 16 * 2
		tid := uuid.Must(uuid.NewV4())
		tw := ls.newTower(act.PlaceTower.Type, p, utils.Object{
			X: float64(act.PlaceTower.X), Y: float64(act.PlaceTower.Y),
			W: w, H: h,
		})
		tw.ID = tid.String()

		l := lstate.Lines[p.LineID]
		err := l.Graph.AddTower(tw.ID, act.PlaceTower.X, act.PlaceTower.Y, tw.W, tw.H)
		if err != nil {
			break
		}

		p.Gold -= tower.Towers[act.PlaceTower.Type].Gold
		l.Towers[tw.ID] = tw

		ls.recalculateLineUnitStepsAndMove(lstate, p.LineID, noTowerID)
	case action.UpdateTower:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		p := lstate.Players[act.UpdateTower.PlayerID]
		l := lstate.Lines[p.LineID]
		t := l.Towers[act.UpdateTower.TowerID]

		if !t.CanUpdateTo(act.UpdateTower.TowerType) {
			break
		}

		tw := ls.newTower(act.UpdateTower.TowerType, p, t.Object)
		p.Gold -= tower.Towers[tw.Type].Gold
		tw.ID = t.ID
		l.Towers[tw.ID] = tw

	case action.RemoveTower:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		p := lstate.Players[act.RemoveTower.PlayerID]
		l := lstate.Lines[p.LineID]
		t := l.Towers[act.RemoveTower.TowerID]

		lstate.Players[act.RemoveTower.PlayerID].Gold += tower.Towers[t.Type].Gold / 2

		// TODO: Add the LineID
		for lid, l := range lstate.Lines {
			if ok := l.Graph.RemoveTower(act.RemoveTower.TowerID); ok {
				delete(l.Towers, act.RemoveTower.TowerID)
				ls.recalculateLineUnitStepsAndMove(lstate, lid, act.RemoveTower.TowerID)
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
			MaxHealth:     float64(uu.Current.Health),
			Health:        float64(uu.Current.Health),
			MaxShield:     uu.Current.Shield,
			Shield:        uu.Current.Shield,
			Level:         uu.Level,
			MovementSpeed: uu.Current.MovementSpeed,
			Bounty:        uu.Current.Income,
			CreatedAt:     time.Now(),
			LastAttack:    time.Now(),
		}

		if u.HasAbility(ability.Hybrid) {
			u.Hybrid(cp.Income, ls.findPlayerByLineID(act.SummonUnit.CurrentLineID).Income)
		}

		u.Path, u.TargetTowerID = l.Graph.AStar(u.X, u.Y, u.MovementSpeed, u.Facing, l.Graph.DeathNode.X, l.Graph.DeathNode.Y, bu.Environment, u.HasAbility(ability.Attack), atScale)
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
				if nu.Abilities != nil {
					for k, v := range nu.Abilities {
						switch k {
						case ability.Split.String():
							var a AbilitySplit
							_ = mapstructure.Decode(v, &a)
							nu.Abilities[k] = a
						case ability.Burrow.String():
							var a AbilityBurrow
							d, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
								DecodeHook:       mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
								WeaklyTypedInput: true,
								Result:           &a,
							})
							_ = d.Decode(v)
							nu.Abilities[k] = a
						case ability.Resurrection.String():
							var a AbilityResurrection
							d, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
								DecodeHook:       mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
								WeaklyTypedInput: true,
								Result:           &a,
							})
							_ = d.Decode(v)
							nu.Abilities[k] = a
						default:
							log.Fatal(fmt.Sprintf("ability %s not found", k))
						}
					}
				}
				//if nu.Buffs != nil {
				//for k, v := range nu.Buffs {
				//switch k {
				//case buff.Burrowoed.String():
				//var b BuffBurrowed
				//_ = mapstructure.Decode(v, &b)
				//nu.Buffs[k] = b
				//default:
				//log.Fatal(fmt.Sprintf("buff %s not found", k))
				//}
				//}
				//}

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
				at := t
				if _, ok := tids[id]; !ok {
					atws[id] = struct{}{}
				}
				delete(tids, id)
				nt := Tower(*at)
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

// recalculateLineUnitStepsAndMove  will recalculate the paths on lid. The twID is if
// a tower, the one with the ID, was removed to the Attackers should move
// It also moves the units
func (ls *Lines) recalculateLineUnitStepsAndMove(lstate LinesState, lid int, twID string) {
	t := time.Now()
	ls.moveLineUnitsTo(lstate, lid, t)

	ls.recalculateLineUnitSteps(lstate, lid, twID)
}

// recalculateLineUnitSteps will recalculate the paths on lid. The twID is if
// a tower, the one with the ID, was removed to the Attackers should move
func (ls *Lines) recalculateLineUnitSteps(lstate LinesState, lid int, twID string) {
	l := lstate.Lines[lid]
	for _, u := range l.Units {
		// This means the unit is attacking the tower so no need to recalculate any path
		if u.HasAbility(ability.Attack) && ((twID == "" || u.TargetTowerID != twID) && len(u.Path) == 0) {
			continue
		}
		u.Path, u.TargetTowerID = l.Graph.AStar(u.X, u.Y, u.MovementSpeed, u.Facing, l.Graph.DeathNode.X, l.Graph.DeathNode.Y, unit.Units[u.Type].Environment, u.HasAbility(ability.Attack), atScale)
		u.HashPath = graph.HashSteps(u.Path)
	}
}

func (ls *Lines) moveLineUnitsTo(lstate LinesState, lid int, t time.Time) {
	l := lstate.Lines[lid]
	lmoves := 1
	if !t.IsZero() && !l.UpdatedAt.IsZero() {
		lmoves = int(t.Sub(l.UpdatedAt).Milliseconds() / tpsMS)
	}
	// First of all we need to get all the units that could be Unburrowed
	// in this TPS so then we can check if any unit after moving is steping
	// on one of them and unburrow it
	burrowedUnits := make(map[string]*Unit)
	for _, u := range l.Units {
		if u.CanUnburrow(t) {
			burrowedUnits[u.ID] = u
		}
	}
	// We'll move all the Units 1 by 1 so we can calculate if they have
	// an aura around and if they have been attacked/killed and if they
	// reached the end to steal a live and change lines
	for i := 1; i < lmoves+1; i++ {
		for _, u := range l.Units {
			if !u.CanBeAttacked(t) {
				continue
			}
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

			for bid, bu := range burrowedUnits {
				if u.IsColliding(bu.Object) {
					bu.Unburrow()
					delete(burrowedUnits, bid)
				}
			}

			// We reached the end of the line
			// If it has the ability 'Attack' we have to check if it has
			// a TargetTowerID as then it means is attacking and not that
			// reached the end of the line
			if len(u.Path) == 0 {
				if u.HasAbility(ability.Attack) && u.TargetTowerID != "" {
					u.AnimationCount += 1
					if u.CanAttack(t) {
						cu := lstate.Players[u.PlayerID].UnitUpdates[u.Type]
						tw, ok := l.Towers[u.TargetTowerID]
						if ok {
							tw.Health -= cu.Current.Damage
							u.LastAttack = t
							if tw.Health <= 0 {
								tw.Health = 0
								if ok := l.Graph.RemoveTower(tw.ID); ok {
									delete(l.Towers, tw.ID)
									ls.recalculateLineUnitSteps(lstate, lid, tw.ID)
								}
							}
						}
					}
				} else {
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
		}
		// Now that the unit has moved we'll calculate if any
		// tower can attack any Unit in their new positions
		for _, tw := range l.Towers {
			if !tw.CanAttack(t) {
				continue
			}
			// Get the closes unit to the current tower to attack it
			var (
				minCost     int = 0
				minCostUnit *Unit

				// The potential Camouflage units
				isAttacker     bool
				minCostCam     int = 0
				minCostCamUnit *Unit

				targetUnit *Unit
			)
			if tw.TargetUnitID != "" {
				if u, ok := l.Units[tw.TargetUnitID]; ok {
					if tw.CanAttackUnit(u) {
						targetUnit = u
					} else {
						tw.TargetUnitID = ""
					}
				} else {
					tw.TargetUnitID = ""
				}
			}
			for _, u := range l.Units {
				if !tw.CanAttackUnit(u) {
					continue
				}
				// Target is based on the unit with the greatest cost
				up := lstate.Players[u.PlayerID]
				ug := up.UnitUpdates[u.Type].Current.Gold
				if u.HasAbility(ability.Camouflage) {
					if minCostCam == 0 {
						minCostCam = ug
						minCostCamUnit = u
					}
					if ug > minCostCam {
						minCostCam = ug
						minCostCamUnit = u
					}
				} else {
					if u.HasAbility(ability.Attack) {
						if !isAttacker {
							minCost = ug
							minCostUnit = u
						} else {
							if minCost == 0 {
								minCost = ug
							}
							if ug >= minCost {
								minCost = ug
								minCostUnit = u
							}
						}
					} else {
						if minCost == 0 {
							minCost = ug
						}
						if ug >= minCost {
							minCost = ug
							minCostUnit = u
						}
					}
				}
			}
			if minCostUnit == nil && minCostCamUnit != nil {
				minCostUnit = minCostCamUnit
			}
			if minCostUnit != nil {
				// We replace the minCostUnit calculated with the targetUnit only if the minCostUnit has 'Attack' (top priority)
				// and if the targetUnit has 'Camouflage' (no priority)
				if !minCostUnit.HasAbility(ability.Attack) && targetUnit != nil && !targetUnit.HasAbility(ability.Camouflage) {
					minCostUnit = targetUnit
				}
				// Tower Attack
				minCostUnit.TakeDamage(tower.Towers[tw.Type].Damage)
				// The attack was done so we register it
				tw.LastAttack = t
				tw.TargetUnitID = minCostUnit.ID

				ls.checkAfterDamage(lstate, l, tw, minCostUnit, t)
				// If the Tower does AoE Damage we need to check again all the units
				// except the current and damage them
				if tower.Towers[tw.Type].AoE != 0 {
					centerUnit := utils.Object{X: minCostUnit.X + 8, Y: minCostUnit.Y + 8}
					for _, u := range l.Units {
						if u.ID == minCostUnit.ID {
							continue
						}
						if !u.CanBeAttacked(t) {
							continue
						}
						bt := tower.Towers[tw.Type]
						if !u.IsCollidingCircle(centerUnit, float64(bt.AoE)*16) {
							continue
						}
						u.TakeDamage(tower.Towers[tw.Type].AoEDamage)
						ls.checkAfterDamage(lstate, l, tw, u, t)
					}
				}
			}
		}
	}
	l.UpdatedAt = t
}

func (ls Lines) checkAfterDamage(lstate LinesState, l *Line, tw *Tower, u *Unit, t time.Time) {
	if u.Health <= u.MaxHealth/2 && u.HasAbility(ability.Burrow) && !u.WasBurrowed() {
		if u.Abilities == nil {
			u.Abilities = make(map[string]interface{})
		}
		u.Abilities[ability.Burrow.String()] = AbilityBurrow{
			BurrowAt: time.Now(),
		}
		u.AddBuff(buff.Burrowoed)
	}
	// Unit is killed
	if u.Health <= 0 {
		if u.HasAbility(ability.Split) {
			if u.Abilities != nil {
				if as, ok := u.Abilities[ability.Split.String()].(AbilitySplit); ok {
					// TODO: Check if it moved into another line
					if _, ok := l.Units[as.UnitID]; !ok {
						u.Health = 0

						// Unit Killed by player so we give gold to the player
						cp := lstate.Players[tw.PlayerID]
						cp.Gold += u.Bounty
					}
				}
			} else {
				// TODO: This should only be done on the server not on the client port
				u1 := *u
				u2 := *u

				u1.ID = uuid.Must(uuid.NewV4()).String()
				u2.ID = uuid.Must(uuid.NewV4()).String()

				h := u.MaxHealth / 2

				u1.MaxHealth = h
				u1.Health = h
				u2.MaxHealth = h
				u2.Health = h

				u1.MovementSpeed = u1.MovementSpeed * 1.20
				u2.MovementSpeed = u2.MovementSpeed * 1.20

				// The second unit created we move it 10 positions forward if possible
				for i := 0; i < 20; i++ {
					if len(u2.Path) != 0 {
						nextStep := u2.Path[0]
						u2.Path = u2.Path[1:]
						u2.MovingCount += 1
						u2.Y = nextStep.Y
						u2.X = nextStep.X
						u2.Facing = nextStep.Facing
					}
				}

				if u1.Abilities == nil {
					u1.Abilities = make(map[string]interface{})
					u2.Abilities = make(map[string]interface{})
				}

				u1.Abilities[ability.Split.String()] = AbilitySplit{
					UnitID: u2.ID,
				}
				u2.Abilities[ability.Split.String()] = AbilitySplit{
					UnitID: u1.ID,
				}

				l.Units[u1.ID] = &u1
				l.Units[u2.ID] = &u2
			}
		} else if u.HasAbility(ability.Resurrection) && !u.WasResurrected() {
			u.Health = 0
			if u.Abilities == nil {
				u.Abilities = make(map[string]interface{})
			}
			u.Abilities[ability.Resurrection.String()] = AbilityResurrection{
				KilledAt: t,
			}
			u.AddBuff(buff.Resurrecting)
			return
		} else {
			u.Health = 0

			// Unit Killed by player so we give gold to the player
			cp := lstate.Players[tw.PlayerID]
			cp.Gold += u.Bounty
		}

		// Delete Unit
		delete(l.Units, u.ID)
	}
}

func (ls *Lines) newLine(lid int) *Line {
	x, y := ls.store.Map.GetHomeCoordinates(lid)
	g, err := graph.New(x+16, y+16, 16, 84, 16, 7, 74, 3)
	if err != nil {
		panic(err)
	}
	return &Line{
		ID:     lid,
		Towers: make(map[string]*Tower),
		Units:  make(map[string]*Unit),
		Graph:  g,
	}
}

func unitUpdate(nlvl int, ut string, u unit.Stats) unit.Stats {
	bu := unit.Units[ut]

	u.Health = float64(levelToValue(nlvl, int(bu.Health)))
	u.Damage = float64(levelToValue(nlvl, int(bu.Damage)))
	u.Gold = levelToValue(nlvl, bu.Gold)
	u.Income = int(math.Round(float64(u.Gold) / float64(incomeFactor)))
	u.Shield = float64(levelToValue(nlvl, int(bu.Shield)))

	return u
}

func levelToValue(lvl, base int) int {
	fb := float64(base)
	for i := 1; i < lvl; i++ {
		fb = fb * math.Pow(math.E, updateFactor)
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

	u.Path, u.TargetTowerID = nl.Graph.AStar(u.X, u.Y, u.MovementSpeed, u.Facing, nl.Graph.DeathNode.X, nl.Graph.DeathNode.Y, unit.Units[u.Type].Environment, u.HasAbility(ability.Attack), atScale)
	u.HashPath = graph.HashSteps(u.Path)

	u.CreatedAt = time.Now()
	if u.HasAbility(ability.Hybrid) {
		u.Hybrid(lstate.Players[u.PlayerID].Income, ls.findPlayerByLineID(nlid).Income)
	}
	nl.Units[u.ID] = u
}

func (ls *Lines) newTower(tt string, p *Player, o utils.Object) *Tower {
	ot := tower.Towers[tt]
	return &Tower{
		Object:     o,
		Type:       tt,
		LineID:     p.LineID,
		PlayerID:   p.ID,
		Health:     ot.Health,
		LastAttack: time.Now(),
	}
}
