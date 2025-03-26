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
	"github.com/xescugc/go-flux/v2"
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
	atScale  = true
	useCache = true

	incomeTimer = 15

	updateFactor     = 0.1
	updateCostFactor = 5
	incomeFactor     = 5

	resurrectionRank1 = 0.25

	noTowerID = ""

	projectileCannonball = "cannonball"
	projectileArrow      = "arrow"
)

var (
	tpsMS = (time.Second / 60).Milliseconds()
)

type Lines struct {
	*flux.ReduceStore[LinesState, *action.Action]

	store *Store

	mxLines sync.RWMutex
}

type LinesState struct {
	Lines   map[int]*Line
	Players map[string]*Player

	// IncomeTimer is the internal counter that goes from 15 to 0
	IncomeTimer int
	StartedAt   time.Time

	Error   string
	ErrorAt time.Time
}

type Line struct {
	ID          int
	Towers      map[string]*Tower
	Projectiles map[string]*Projectile
	Units       map[string]*Unit

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

type Projectile struct {
	ID string

	utils.Object

	TargetUnitID string
	Damage       float64

	AoE       int
	AoEDamage float64

	PlayerID string

	ImageKey string

	Type string
}

func (p *Projectile) CalculateImageKey(vx, vy float64) {
	if p.Type == projectileCannonball {
		p.ImageKey = "cannonball"
		return
	}
	at := math.Atan2(vy, vx)
	angle := (at * (180 / math.Pi)) * -1
	if angle > -22.5 && angle < 22.5 {
		p.ImageKey = "arrow-right"
	} else if angle < -22.5 && angle > -67.5 {
		p.ImageKey = "arrow-down-right"
	} else if angle < -67.5 && angle > -112.5 {
		p.ImageKey = "arrow-down"
	} else if angle < -112.5 && angle > -157.5 {
		p.ImageKey = "arrow-down-left"
	} else if angle > 22.5 && angle < 67.5 {
		p.ImageKey = "arrow-up-right"
	} else if angle > 67.5 && angle < 112.5 {
		p.ImageKey = "arrow-up"
	} else if angle > 112.5 && angle < 157.5 {
		p.ImageKey = "arrow-up-left"
	} else if angle > 157.5 && angle > -157.5 {
		p.ImageKey = "arrow-left"
	}
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

func (t *Tower) FacetKey() string   { return tower.Towers[t.Type].FacesetKey() }
func (t *Tower) IdleKey() string    { return tower.Towers[t.Type].IdleKey() }
func (t *Tower) ProfileKey() string { return tower.Towers[t.Type].ProfileKey() }
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
	centerTower := utils.Object{X: t.X + 16, Y: t.Y + 16}
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
	ID       string
	Name     string
	ImageKey string
	Lives    int
	LineID   int
	Income   int
	Gold     int
	IsBot    bool
	Current  bool
	Winner   bool
	Capacity int

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
	return (p.Gold-p.UnitUpdates[ut].Current.Gold) >= 0 && (p.Capacity+1 <= utils.MaxCapacity)
}
func (p Player) CanUpdateUnit(ut string) bool {
	return (p.Gold - p.UnitUpdates[ut].UpdateCost) >= 0
}
func (p Player) CanUpdateTower(tt string) bool {
	return (p.Gold - tower.Towers[tt].Gold) >= 0
}
func (p Player) CanPlaceTower(tt string) bool {
	return (p.Gold - tower.Towers[tt].Gold) >= 0
}

func NewLines(d *flux.Dispatcher[*action.Action], s *Store) *Lines {
	l := &Lines{
		store: s,
	}

	l.ReduceStore = flux.NewReduceStore(d, l.Reduce, LinesState{
		Lines:       make(map[int]*Line),
		Players:     make(map[string]*Player),
		IncomeTimer: incomeTimer,
		StartedAt:   time.Now(),
	})

	return l
}

func (ls *Lines) ListLines() []*Line {
	ls.mxLines.RLock()
	defer ls.mxLines.RUnlock()

	state := ls.GetState()
	lines := make([]*Line, 0, len(state.Lines))
	for _, l := range state.Lines {
		us := make(map[string]*Unit)
		ts := make(map[string]*Tower)
		ps := make(map[string]*Projectile)
		for uid, u := range l.Units {
			us[uid] = u
		}
		for tid, t := range l.Towers {
			ts[tid] = t
		}
		for pid, p := range l.Projectiles {
			ps[pid] = p
		}
		ll := *l
		ll.Units = us
		ll.Towers = ts
		ll.Projectiles = ps
		lines = append(lines, &ll)
	}
	return lines
}

func (ls *Lines) FindLineByID(id int) *Line {
	ls.mxLines.RLock()
	defer ls.mxLines.RUnlock()

	l := ls.GetState().Lines[id]

	us := make(map[string]*Unit)
	ts := make(map[string]*Tower)
	ps := make(map[string]*Projectile)
	for uid, u := range l.Units {
		us[uid] = u
	}
	for tid, t := range l.Towers {
		ts[tid] = t
	}
	for pid, p := range l.Projectiles {
		ps[pid] = p
	}
	ll := *l
	ll.Units = us
	ll.Towers = ts
	ll.Projectiles = ps

	return &ll
}

// ListPlayers returns the players list and it's meant for reading only purposes
func (ls *Lines) ListPlayers() []*Player {
	ls.mxLines.RLock()
	defer ls.mxLines.RUnlock()

	state := ls.GetState()
	players := make([]*Player, 0, len(state.Players))
	for _, p := range state.Players {
		players = append(players, p)
	}
	return players
}

func (ls *Lines) FindCurrentPlayer() Player {
	ls.mxLines.RLock()
	defer ls.mxLines.RUnlock()

	for _, p := range ls.GetState().Players {
		if p.Current {
			return *p
		}
	}
	return Player{}
}

func (ls *Lines) FindPlayerByID(id string) Player {
	ls.mxLines.RLock()
	defer ls.mxLines.RUnlock()

	p, ok := ls.GetState().Players[id]
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
	for _, p := range ls.GetState().Players {
		if p.LineID == lid {
			return *p
		}
	}
	return Player{}
}

func (ls *Lines) GetIncomeTimer() int {
	ls.mxLines.RLock()
	defer ls.mxLines.RUnlock()

	state := ls.GetState()
	return state.IncomeTimer
}

func (ls *Lines) GetStartedAt() time.Time {
	ls.mxLines.RLock()
	defer ls.mxLines.RUnlock()

	state := ls.GetState()
	return state.StartedAt
}

func (ls *Lines) Reduce(state LinesState, act *action.Action) LinesState {
	switch act.Type {
	case action.IncomeTick:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		state.IncomeTimer -= 1
		if state.IncomeTimer == 0 {
			state.IncomeTimer = incomeTimer
			for _, p := range state.Players {
				p.Gold += p.Income
			}
		}
	case action.AddPlayer:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		var found bool
		for _, p := range state.Players {
			if p.Name == act.AddPlayer.Name {
				found = true
				break
			}
		}

		if found {
			break
		}

		p := &Player{
			ID:       act.AddPlayer.ID,
			Name:     act.AddPlayer.Name,
			ImageKey: act.AddPlayer.ImageKey,
			Lives:    20,
			LineID:   act.AddPlayer.LineID,
			Income:   25,
			Gold:     40,
			IsBot:    act.AddPlayer.IsBot,

			UnitUpdates: make(map[string]UnitUpdate),
		}
		for _, u := range unit.Units {
			p.UnitUpdates[u.Type.String()] = CalculateUnitUpdate(u.Type.String(), UnitUpdate{
				Current: u.Stats,
				Level:   1,

				UpdateCost: updateCostFactor * u.Stats.Gold,
				Next:       unitUpdate(2, u.Type.String(), u.Stats),
			}, 1)
		}

		state.Players[act.AddPlayer.ID] = p
	case action.StartGame:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		for _, p := range state.Players {
			state.Lines[p.LineID] = ls.newLine(p.LineID)
		}
		ls.syncState(&state, act.StartGame.State)
	case action.PlaceTower:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		p := state.Players[act.PlaceTower.PlayerID]

		if !p.CanPlaceTower(act.PlaceTower.Type) {
			state.Error = fmt.Sprintf("Cannot place tower %s", tower.Towers[act.PlaceTower.Type].Name())
			state.ErrorAt = time.Now()
			break
		}

		var w, h int = 16 * 2, 16 * 2
		tid := uuid.Must(uuid.NewV4())
		tw := ls.newTower(act.PlaceTower.Type, p, utils.Object{
			X: float64(act.PlaceTower.X), Y: float64(act.PlaceTower.Y),
			W: w, H: h,
		})
		tw.ID = tid.String()

		l := state.Lines[p.LineID]
		err := l.Graph.AddTower(tw.ID, act.PlaceTower.X, act.PlaceTower.Y, tw.W, tw.H)
		if err != nil {
			state.Error = err.Error()
			state.ErrorAt = time.Now()
			break
		}

		p.Gold -= tower.Towers[act.PlaceTower.Type].Gold
		l.Towers[tw.ID] = tw

		ls.recalculateLineUnitStepsAndMove(state, p.LineID, noTowerID)
	case action.UpdateTower:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		p := state.Players[act.UpdateTower.PlayerID]
		l := state.Lines[p.LineID]
		t := l.Towers[act.UpdateTower.TowerID]

		tw := ls.newTower(act.UpdateTower.TowerType, p, t.Object)

		if !t.CanUpdateTo(act.UpdateTower.TowerType) || !p.CanUpdateTower(tw.Type) {
			state.Error = fmt.Sprintf("Cannot update to tower %s", tower.Towers[act.UpdateTower.TowerType].Name())
			state.ErrorAt = time.Now()
			break
		}

		p.Gold -= tower.Towers[tw.Type].Gold
		tw.ID = t.ID
		l.Towers[tw.ID] = tw

	case action.RemoveTower:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		p := state.Players[act.RemoveTower.PlayerID]
		l := state.Lines[p.LineID]
		t := l.Towers[act.RemoveTower.TowerID]

		state.Players[act.RemoveTower.PlayerID].Gold += tower.Towers[t.Type].Gold / 2

		// TODO: Add the LineID
		for lid, l := range state.Lines {
			if ok := l.Graph.RemoveTower(act.RemoveTower.TowerID); ok {
				delete(l.Towers, act.RemoveTower.TowerID)
				ls.recalculateLineUnitStepsAndMove(state, lid, act.RemoveTower.TowerID)
			}
		}
	case action.SummonUnit:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		cp := state.Players[act.SummonUnit.PlayerID]
		if !cp.CanSummonUnit(act.SummonUnit.Type) {
			state.Error = fmt.Sprintf("Cannot summon unit %s", unit.Units[act.SummonUnit.Type].Name())
			state.ErrorAt = time.Now()
			break
		}
		state.Players[act.SummonUnit.PlayerID].Income += cp.UnitUpdates[act.SummonUnit.Type].Current.Income
		state.Players[act.SummonUnit.PlayerID].Gold -= cp.UnitUpdates[act.SummonUnit.Type].Current.Gold
		//state.Players[act.SummonUnit.PlayerID].Capacity += 1
		cp.Capacity += 1

		uu := cp.UnitUpdates[act.SummonUnit.Type]
		bu := unit.Units[act.SummonUnit.Type]

		l := state.Lines[act.SummonUnit.CurrentLineID]

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

		u.Path, u.TargetTowerID = l.Graph.AStar(u.X, u.Y, u.MovementSpeed, u.Facing, l.Graph.DeathNode.X, l.Graph.DeathNode.Y, bu.Environment, u.HasAbility(ability.Attack), atScale, useCache)
		u.HashPath = graph.HashSteps(u.Path)
		l.Units[u.ID] = u
	case action.TPS:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		for lid := range state.Lines {
			ls.moveLineUnitsTo(state, lid, act.TPS.Time)
		}
	case action.RemovePlayer:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		p := state.Players[act.RemovePlayer.ID]

		// TODO: Add LineID
		delete(state.Lines, p.LineID)

		for _, l := range state.Lines {
			for _, u := range l.Units {
				if u.PlayerID == act.RemovePlayer.ID {
					delete(l.Units, u.ID)
				}
			}
		}

		delete(state.Players, act.RemovePlayer.ID)
		if len(state.Players) == 1 {
			for _, p := range state.Players {
				// As there is only 1 we can do it this way
				p.Winner = true
			}
		}

	case action.UpdateUnit:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		u := unit.Units[act.UpdateUnit.Type]
		buu := state.Players[act.UpdateUnit.PlayerID].UnitUpdates[act.UpdateUnit.Type]

		if !state.Players[act.UpdateUnit.PlayerID].CanUpdateUnit(act.UpdateUnit.Type) {
			state.Error = fmt.Sprintf("Cannot update unit %s", unit.Units[act.UpdateUnit.Type].Name())
			state.ErrorAt = time.Now()
			break
		}

		state.Players[act.UpdateUnit.PlayerID].Gold -= buu.UpdateCost
		state.Players[act.UpdateUnit.PlayerID].UnitUpdates[act.UpdateUnit.Type] = UnitUpdate{
			Current:    buu.Next,
			Level:      buu.Level + 1,
			UpdateCost: updateCostFactor * buu.Next.Gold,
			Next:       unitUpdate(buu.Level+2, u.Type.String(), u.Stats),
		}

	case action.SyncState:
		ls.mxLines.Lock()
		defer ls.mxLines.Unlock()

		ls.syncState(&state, *act.SyncState)

	}
	return state
}

// recalculateLineUnitStepsAndMove  will recalculate the paths on lid. The twID is if
// a tower, the one with the ID, was removed to the Attackers should move
// It also moves the units
func (ls *Lines) recalculateLineUnitStepsAndMove(state LinesState, lid int, twID string) {
	t := time.Now()
	ls.moveLineUnitsTo(state, lid, t)

	ls.recalculateLineUnitSteps(state, lid, twID)
}

// recalculateLineUnitSteps will recalculate the paths on lid. The twID is if
// a tower, the one with the ID, was removed to the Attackers should move
func (ls *Lines) recalculateLineUnitSteps(state LinesState, lid int, twID string) {
	l := state.Lines[lid]
	for _, u := range l.Units {
		// This means the unit is attacking the tower so no need to recalculate any path
		if u.HasAbility(ability.Attack) && ((twID == "" || u.TargetTowerID != twID) && len(u.Path) == 0) {
			continue
		}
		u.Path, u.TargetTowerID = l.Graph.AStar(u.X, u.Y, u.MovementSpeed, u.Facing, l.Graph.DeathNode.X, l.Graph.DeathNode.Y, unit.Units[u.Type].Environment, u.HasAbility(ability.Attack), atScale, useCache)
		u.HashPath = graph.HashSteps(u.Path)
	}
}

// moveLineUnitsTo will move the units of 'lid' to the 't' time compared to the last time they
// were moved.
// It'll also check:
// * If towers can attack
// * If units can use abilities
// * If units are dead, reached end or can go to the next line
func (ls *Lines) moveLineUnitsTo(state LinesState, lid int, t time.Time) {
	l := state.Lines[lid]
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
			//
			// This moves the unit to the next Path position
			if len(u.Path) != 0 {
				nextStep := u.Path[0]
				u.Path = u.Path[1:]
				u.MovingCount += 1
				u.Y = nextStep.Y
				u.X = nextStep.X
				u.Facing = nextStep.Facing
			}

			// We check if the new path is stepping into a burrowed
			// unit so we need to unburrow it
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
					// Attacking the tower
					u.AnimationCount += 1
					if u.CanAttack(t) {
						cu := state.Players[u.PlayerID].UnitUpdates[u.Type]
						tw, ok := l.Towers[u.TargetTowerID]
						if ok {
							tw.Health -= cu.Current.Damage
							u.LastAttack = t
							if tw.Health <= 0 {
								tw.Health = 0
								if ok := l.Graph.RemoveTower(tw.ID); ok {
									delete(l.Towers, tw.ID)
									ls.recalculateLineUnitSteps(state, lid, tw.ID)
								}
							}
						}
					}
				} else {
					// Reached the end of the line so we have to steal
					// one live and move to the next line (if any)
					var fpID string
					for pid, p := range state.Players {
						if p.LineID == lid {
							fpID = pid
							break
						}
					}
					// We steal a Live
					ls.stealLive(state, fpID, u.PlayerID)
					nlid := ls.store.Map.GetNextLineID(u.CurrentLineID)
					// If the next line is the owner of the Unit we remove it
					// If not then we change the unit to the next line
					if nlid == u.PlayerLineID {
						cp := state.Players[u.PlayerID]
						// We check if the unit is a Split and then we check if the other partner is in
						// the same line
						if u.HasAbility(ability.Split) && u.Abilities != nil {
							if as, ok := u.Abilities[ability.Split.String()].(AbilitySplit); ok {
								if _, ok := l.Units[as.UnitID]; !ok {
									// We know now the other is not in the line and it's the last one so we can reduce capacity
									cp.Capacity -= 1
								}
							}
						} else {
							cp.Capacity -= 1
						}
						delete(state.Lines[lid].Units, u.ID)
					} else {
						ls.changeUnitLine(state, u, nlid)
					}
				}
			}
		}
		// Now we need to move the Projectiles and check if they are hitting the target
		// we'll also check first if the projectile is already hitting the target as it
		// could happen that the unit move towards the projectile
		for pid, p := range l.Projectiles {
			u, ok := l.Units[p.TargetUnitID]
			if !ok {
				// If there is no unit for that projectile we just remove it
				delete(l.Projectiles, pid)
				continue
			}
			if p.IsColliding(u.Object) {
				ls.attackUnit(state, l, p, u, t)
				delete(l.Projectiles, pid)
				continue
			}
			distance := p.PDistance(u.Object)

			vx := (u.X - p.X)
			vy := (u.Y - p.Y)

			p.X = 3/distance*vx + p.X
			p.Y = 3/distance*vy + p.Y

			if p.IsColliding(u.Object) {
				ls.attackUnit(state, l, p, u, t)
				delete(l.Projectiles, pid)
				continue
			}

			p.CalculateImageKey(vx, vy)
		}

		//if !ls.store.isOnServer {
		//// If we are on the client we do not calculate any of those things
		//continue
		//}

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

				//targetUnit *Unit
			)
			if tw.TargetUnitID != "" {
				if u, ok := l.Units[tw.TargetUnitID]; ok {
					if tw.CanAttackUnit(u) {
						minCostUnit = u
						//targetUnit = u
					} else {
						tw.TargetUnitID = ""
					}
				} else {
					tw.TargetUnitID = ""
				}
			}
			if minCostUnit == nil {
				for _, u := range l.Units {
					if !tw.CanAttackUnit(u) {
						continue
					}
					// Target is based on the unit with the greatest cost
					up := state.Players[u.PlayerID]
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
			}
			if minCostUnit == nil && minCostCamUnit != nil {
				minCostUnit = minCostCamUnit
			}
			if minCostUnit != nil {
				// TODO: Should we change the current target if a priority target comes to range?
				// We replace the minCostUnit calculated with the targetUnit only if the minCostUnit has 'Attack' (top priority)
				// and if the targetUnit has 'Camouflage' (no priority)
				//if !minCostUnit.HasAbility(ability.Attack) && targetUnit != nil && !targetUnit.HasAbility(ability.Camouflage) {
				//minCostUnit = targetUnit
				//}
				ot := tower.Towers[tw.Type]
				pid := uuid.Must(uuid.NewV4()).String()
				p := &Projectile{
					// The Projectile starts at the middle of the tower
					Object: utils.Object{
						X: tw.X + 16,
						Y: tw.Y + 16,
						W: 13, H: 5,
					},
					TargetUnitID: minCostUnit.ID,
					Damage:       ot.Damage,
					AoE:          ot.AoE,
					AoEDamage:    ot.AoEDamage,
					PlayerID:     tw.PlayerID,
					Type:         projectileCannonball,
				}

				if ot.ShootsArrows() {
					p.Type = projectileArrow

					vx := (minCostUnit.X - p.X)
					vy := (minCostUnit.Y - p.Y)

					p.CalculateImageKey(vx, vy)
				}

				l.Projectiles[pid] = p
				// The attack was done so we register it
				tw.LastAttack = t
				tw.TargetUnitID = minCostUnit.ID
			}
		}
	}
	l.UpdatedAt = t
}

func (ls Lines) attackUnit(state LinesState, l *Line, p *Projectile, tu *Unit, t time.Time) {
	// Tower Attack
	tu.TakeDamage(p.Damage)
	ls.checkAfterDamage(state, l, p, tu, t)
	// If the Tower does AoE Damage we need to check again all the units
	// except the current and damage them
	if p.AoE != 0 {
		centerUnit := utils.Object{X: tu.X + 8, Y: tu.Y + 8}
		for _, u := range l.Units {
			if u.ID == tu.ID {
				continue
			}
			if !u.CanBeAttacked(t) {
				continue
			}
			if !u.IsCollidingCircle(centerUnit, float64(p.AoE)*16) {
				continue
			}
			u.TakeDamage(p.AoEDamage)
			ls.checkAfterDamage(state, l, p, u, t)
		}
	}
}

func (ls Lines) checkAfterDamage(state LinesState, l *Line, p *Projectile, u *Unit, t time.Time) {
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
						cp := state.Players[p.PlayerID]
						cp.Gold += u.Bounty

						// Delete Unit as we know this is the last one of the split left
						// in the lane so we also reduce the capacity
						up := state.Players[u.PlayerID]
						up.Capacity -= 1
						delete(l.Units, u.ID)
						return
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
			cp := state.Players[p.PlayerID]
			cp.Gold += u.Bounty
		}

		// Delete Unit
		up := state.Players[u.PlayerID]
		up.Capacity -= 1
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
		ID:          lid,
		Towers:      make(map[string]*Tower),
		Units:       make(map[string]*Unit),
		Projectiles: make(map[string]*Projectile),
		Graph:       g,
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

func (ls *Lines) stealLive(state LinesState, fpID, tpID string) {
	fp := state.Players[fpID]
	tp := state.Players[tpID]

	fp.Lives -= 1
	if fp.Lives < 0 {
		fp.Lives = 0
	} else {
		tp.Lives += 1
	}

	var stillPlayersLeft bool
	for _, p := range state.Players {
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

func (ls *Lines) changeUnitLine(state LinesState, u *Unit, nlid int) {
	cl := state.Lines[u.CurrentLineID]
	// As we are gonna move it to another line
	// we remove it
	delete(cl.Units, u.ID)

	u.CurrentLineID = nlid

	nl := state.Lines[u.CurrentLineID]

	n := nl.Graph.GetRandomSpawnNode()
	u.X = float64(n.X)
	u.Y = float64(n.Y)

	u.Path, u.TargetTowerID = nl.Graph.AStar(u.X, u.Y, u.MovementSpeed, u.Facing, nl.Graph.DeathNode.X, nl.Graph.DeathNode.Y, unit.Units[u.Type].Environment, u.HasAbility(ability.Attack), atScale, useCache)
	u.HashPath = graph.HashSteps(u.Path)

	u.CreatedAt = time.Now()
	if u.HasAbility(ability.Hybrid) {
		u.Hybrid(state.Players[u.PlayerID].Income, ls.findPlayerByLineID(nlid).Income)
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

func (ls *Lines) syncState(state *LinesState, ss action.SyncStatePayload) {
	for lid, l := range ss.Lines.Lines {
		cl, ok := state.Lines[lid]
		if !ok {
			cl = ls.newLine(lid)
			state.Lines[lid] = cl
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
			nt := Tower(*at)

			if _, ok := tids[id]; !ok {
				atws[id] = struct{}{}
			}

			delete(tids, id)
			// So this way we have the projectiles on the Client
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

		//if !ls.store.isOnServer {
		//// If we are on the client we do not calculate any of those things
		//continue
		//}

		// NOTE: This is so the Client is creating the projectiles and keeping
		// track of them, if we only serve from the server as it returns every
		// 1/4s looks weird
		cl.Projectiles = make(map[string]*Projectile)
		pids := make(map[string]struct{})
		for id := range cl.Projectiles {
			pids[id] = struct{}{}
		}
		for id, p := range l.Projectiles {
			ap := p
			if _, ok := pids[id]; ok {
				continue
			}
			delete(pids, id)
			np := Projectile(*ap)
			cl.Projectiles[id] = &np
		}
		for id := range pids {
			delete(cl.Projectiles, id)
		}
	}

	// Sync Players
	pids := make(map[string]struct{})
	for id := range state.Players {
		pids[id] = struct{}{}
	}
	for id, p := range ss.Players.Players {
		delete(pids, id)
		np := Player{
			ID:          p.ID,
			Name:        p.Name,
			ImageKey:    p.ImageKey,
			Lives:       p.Lives,
			LineID:      p.LineID,
			Income:      p.Income,
			Gold:        p.Gold,
			IsBot:       p.IsBot,
			Current:     p.Current,
			Winner:      p.Winner,
			Capacity:    p.Capacity,
			UnitUpdates: make(map[string]UnitUpdate),
		}
		for t, uu := range p.UnitUpdates {
			np.UnitUpdates[t] = UnitUpdate(uu)
		}
		state.Players[id] = &np
	}
	for id := range pids {
		delete(state.Players, id)
	}
	state.IncomeTimer = ss.Players.IncomeTimer
	state.StartedAt = ss.StartedAt
}

// CalculateUnitUpdate will return the UnitUpdate of t
func CalculateUnitUpdate(t string, buu UnitUpdate, lvl int) UnitUpdate {
	u := unit.Units[t]
	if lvl == 1 {
		return UnitUpdate{
			Current:    u.Stats,
			Level:      lvl,
			UpdateCost: updateCostFactor * u.Gold,
			Next:       unitUpdate(lvl+1, t, u.Stats),
		}
	}
	uu := UnitUpdate{
		Current:    buu.Next,
		Level:      buu.Level + 1,
		UpdateCost: updateCostFactor * buu.Next.Gold,
		Next:       unitUpdate(buu.Level+2, u.Type.String(), u.Stats),
	}
	return uu
}
