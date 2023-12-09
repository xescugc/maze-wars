package store_test

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/xescugc/ltw/action"
	"github.com/xescugc/ltw/store"
	"github.com/xescugc/ltw/tower"
	"github.com/xescugc/ltw/unit"
	"github.com/xescugc/ltw/utils"
	"nhooyr.io/websocket"
)

// This test are meant to check which Stores interact with Actions
// so we'll dispatch and action and expect changes to the stores
// that we want to have changes on and expect no changes to the rest
//
// Each test will require no preset data, it'll be independent.
// Each test block is for one action in case we want to have multiple conditions
// Not all action Types are for the 'store' to deal with so some may not have any
// relevance

var (
	playersInitialState = func() store.PlayersState {
		return store.PlayersState{
			IncomeTimer: 15,
			Players:     make(map[string]*store.Player),
		}
	}

	towersInitialState = func() store.TowersState {
		return store.TowersState{
			Towers: make(map[string]*store.Tower),
		}
	}

	unitsInitialState = func() store.UnitsState {
		return store.UnitsState{
			Units: make(map[string]*store.Unit),
		}
	}
)

func TestActions(t *testing.T) {
	assert.Len(t, action.TypeValues(), 27, "This is a hard check to remember that if a new action Type is added it should also have a test on this file")
}

func TestEmpty(t *testing.T) {
	s := initStore()
	equalStore(t, s)
}

func TestSummonUnit(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		clid := 2

		eu := &store.Unit{
			// As the ID is a UUID we cannot guess it
			//ID: units[0].ID,
			MovingObject: utils.MovingObject{
				Object: utils.Object{
					// This is also random
					//X: units[0].X, Y: units[0].Y,
					W: 16, H: 16,
				},
				Facing: utils.Down,
			},
			Type:          unit.Spirit.String(),
			PlayerID:      p.ID,
			PlayerLineID:  p.LineID,
			CurrentLineID: clid,
			Health:        unit.Units[unit.Spirit.String()].Health,
		}

		a := action.NewSummonUnit(eu.Type, p.ID, p.LineID, clid)
		s.Dispatch(a)

		units := s.Units.List()

		// As this are random assigned we cannot expect them
		eu.ID, eu.X, eu.Y = units[0].ID, units[0].X, units[0].Y

		// We need to set the path after the X, Y are set
		eu.Path = s.Units.Astar(s.Map, clid, eu.MovingObject, nil)
		eu.HashPath = utils.HashSteps(eu.Path)

		// AS the Unit is created we remove it from the gold
		// and add more income
		p.Gold -= unit.Units[unit.Spirit.String()].Gold
		p.Income += unit.Units[unit.Spirit.String()].Income

		ps := playersInitialState()
		ps.Players[p.ID] = &p

		us := unitsInitialState()
		us.Units[eu.ID] = eu

		equalStore(t, s, ps, us)
	})
	t.Run("Do not reach negative gold", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		clid := 2

		// We start with 40 gold, each Spirit
		// takes 10 gold so with that we can only
		// create 4 so we'll try to create 5
		for i := 0; i < 5; i++ {
			a := action.NewSummonUnit(unit.Spirit.String(), p.ID, p.LineID, clid)
			s.Dispatch(a)
		}

		// I don't want to EXPECT with Units just with Players
		us := s.Units.GetState()

		assert.Equal(t, 4, len(s.Units.List()))

		// We could only create 4 of the 5 so -40
		p.Gold -= 40
		// Only 4 can be created not 5
		p.Income += 4

		ps := playersInitialState()
		ps.Players[p.ID] = &p

		equalStore(t, s, ps, us)
	})
}

func TestTPS(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p, u := summonUnit(s, p)

		s.Dispatch(action.NewTPS())

		ps := playersInitialState()
		ps.Players[p.ID] = &p

		u.Path = u.Path[1:]
		u.MovingCount++
		us := unitsInitialState()
		us.Units[u.ID] = &u

		equalStore(t, s, ps, us)
	})
}

func TestRemoveUnit(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p, u := summonUnit(s, p)

		s.Dispatch(action.NewRemoveUnit(u.ID))

		ps := playersInitialState()
		ps.Players[p.ID] = &p

		equalStore(t, s, ps)
	})
}

func TestStealLive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p1 := addPlayer(s)
		p2 := addPlayer(s)

		s.Dispatch(action.NewStealLive(p1.ID, p2.ID))

		p1.Lives--
		p2.Lives++

		ps := playersInitialState()
		ps.Players[p1.ID] = &p1
		ps.Players[p2.ID] = &p2

		equalStore(t, s, ps)
	})
	t.Run("DeclareWinner", func(t *testing.T) {
		s := initStore()
		p1 := addPlayer(s)
		p2 := addPlayer(s)

		// The lives of a Player are 20 so we go on 30
		// to see it cannot overflow the lives
		for i := 0; i <= 30; i++ {
			s.Dispatch(action.NewStealLive(p1.ID, p2.ID))
		}

		// It should only be 20
		p1.Lives = 0
		p2.Lives += 20
		p2.Winner = true

		ps := playersInitialState()
		ps.Players[p1.ID] = &p1
		ps.Players[p2.ID] = &p2

		equalStore(t, s, ps)
	})
}

func TestPlaceTower(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)

		s.Dispatch(action.NewPlaceTower(tower.Soldier.String(), p.ID, 10, 20))

		p.Gold -= tower.Towers[tower.Soldier.String()].Gold
		ps := playersInitialState()
		ps.Players[p.ID] = &p

		tw := store.Tower{
			ID: s.Towers.List()[0].ID,
			Object: utils.Object{
				X: 10, Y: 20,
				W: 32, H: 32,
			},
			Type:     tower.Soldier.String(),
			LineID:   p.LineID,
			PlayerID: p.ID,
		}
		ts := towersInitialState()
		ts.Towers[tw.ID] = &tw

		equalStore(t, s, ps, ts)
	})
	t.Run("Do not reach negative gold", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)

		// The Player gold is 40 so it should only create 4
		// towers and not 10
		for i := 0; i <= 10; i++ {
			s.Dispatch(action.NewPlaceTower(tower.Soldier.String(), p.ID, 10, 20))
		}

		p.Gold = 0
		ps := playersInitialState()
		ps.Players[p.ID] = &p

		// I don't want to EXPECT with Towers just with Players
		ts := s.Towers.GetState()

		assert.Equal(t, 4, len(s.Towers.List()))

		equalStore(t, s, ps, ts)
	})
	t.Run("Change unit course", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p, u := summonUnit(s, p)

		// We place it in the 10th (any place would be fine) path position so we can force the
		// unit to recalculate the path
		s.Dispatch(action.NewPlaceTower(tower.Soldier.String(), p.ID, int(u.Path[10].X), int(u.Path[10].Y)))

		p.Gold -= tower.Towers[tower.Soldier.String()].Gold
		ps := playersInitialState()
		ps.Players[p.ID] = &p

		tw := store.Tower{
			ID: s.Towers.List()[0].ID,
			Object: utils.Object{
				X: u.Path[10].X, Y: u.Path[10].Y,
				W: 32, H: 32,
			},
			Type:     tower.Soldier.String(),
			LineID:   p.LineID,
			PlayerID: p.ID,
		}
		ts := towersInitialState()
		ts.Towers[tw.ID] = &tw

		u.Path = s.Units.Astar(s.Map, u.CurrentLineID, u.MovingObject, []utils.Object{tw.Object})
		us := unitsInitialState()
		us.Units[u.ID] = &u

		equalStore(t, s, ps, ts, us)
	})
}

func TestRemoveTower(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p, tw := placeTower(s, p)

		s.Dispatch(action.NewRemoveTower(p.ID, tw.ID, tw.Type))

		p.Gold += tower.Towers[tw.Type].Gold / 2

		ps := playersInitialState()
		ps.Players[p.ID] = &p

		equalStore(t, s, ps)
	})
}

func TestIncomeTick(t *testing.T) {
	t.Run("NormalTick", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)

		s.Dispatch(action.NewIncomeTick())

		ps := playersInitialState()
		ps.Players[p.ID] = &p
		ps.IncomeTimer = 14

		equalStore(t, s, ps)
	})
	t.Run("TicksToTriggerIncome", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)

		for i := 0; i <= 14; i++ {
			s.Dispatch(action.NewIncomeTick())
		}
		p.Gold += p.Income

		ps := playersInitialState()
		ps.Players[p.ID] = &p
		ps.IncomeTimer = 15

		equalStore(t, s, ps)
	})
}

func TestTowerAttack(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p, tw := placeTower(s, p)
		p, u := summonUnit(s, p)

		s.Dispatch(action.NewTowerAttack(u.ID, tw.Type))
		u.Health -= tower.Towers[tw.Type].Damage

		ps := playersInitialState()
		ps.Players[p.ID] = &p

		us := unitsInitialState()
		us.Units[u.ID] = &u

		ts := towersInitialState()
		ts.Towers[tw.ID] = &tw

		equalStore(t, s, ps, us, ts)
	})
}

func TestUnitKilled(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p, u := summonUnit(s, p)

		s.Dispatch(action.NewUnitKilled(p.ID, u.Type))
		p.Gold += unit.Units[u.Type].Income

		ps := playersInitialState()
		ps.Players[p.ID] = &p

		us := unitsInitialState()
		us.Units[u.ID] = &u

		equalStore(t, s, ps, us)
	})
}

func TestPlayerReady(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)

		s.Dispatch(action.NewPlayerReady(p.ID))
		p.Ready = true

		ps := playersInitialState()
		ps.Players[p.ID] = &p

		equalStore(t, s, ps)
	})
}

func TestAddPlayer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		sid := "sid"
		id := uuid.Must(uuid.NewV4())
		name := "name"
		lid := 2
		ws := &websocket.Conn{}
		ra := "localhost"
		s.Dispatch(action.NewAddPlayer(sid, id.String(), name, lid, ws, ra))

		p := store.Player{
			ID:     id.String(),
			Name:   name,
			LineID: lid,
			Lives:  20,
			Income: 25,
			Gold:   40,
		}

		ps := playersInitialState()
		ps.Players[p.ID] = &p

		equalStore(t, s, ps)
	})
	t.Run("AlreadyExists", func(t *testing.T) {
		s := initStore()
		sid := "sid"
		sid2 := "sid2"
		id := uuid.Must(uuid.NewV4())
		id2 := uuid.Must(uuid.NewV4())
		name := "name"
		lid := 2
		ws := &websocket.Conn{}
		ra1 := "localhost"
		ra2 := "localhost2"
		s.Dispatch(action.NewAddPlayer(sid, id.String(), name, lid, ws, ra1))
		s.Dispatch(action.NewAddPlayer(sid2, id2.String(), name, lid+1, ws, ra2))

		p := store.Player{
			ID:     id.String(),
			Name:   name,
			LineID: lid,
			Lives:  20,
			Income: 25,
			Gold:   40,
		}

		ps := playersInitialState()
		ps.Players[p.ID] = &p

		equalStore(t, s, ps)
	})
}

func TestRemovePlayer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p, _ = placeTower(s, p)
		p, _ = summonUnit(s, p)

		s.Dispatch(action.NewRemovePlayer("room", p.ID))

		equalStore(t, s)
	})
}

func equalStore(t *testing.T, sto *store.Store, states ...interface{}) {
	pis := playersInitialState()
	tis := towersInitialState()
	uis := unitsInitialState()
	for _, st := range states {
		switch s := st.(type) {
		case store.PlayersState:
			pis = s
		case store.TowersState:
			tis = s
		case store.UnitsState:
			uis = s
		default:
			t.Fatalf("State with type %T is unknown", st)
		}
	}

	assert.Equal(t, pis, sto.Players.GetState().(store.PlayersState))
	assert.Equal(t, tis, sto.Towers.GetState().(store.TowersState))
	assert.Equal(t, uis, sto.Units.GetState().(store.UnitsState))
}
