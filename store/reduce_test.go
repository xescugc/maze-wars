package store_test

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/unit"
	"github.com/xescugc/maze-wars/utils"
	"github.com/xescugc/maze-wars/utils/graph"
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
	atScale = true

	playersInitialState = func() store.PlayersState {
		return store.PlayersState{
			IncomeTimer: 15,
			Players:     make(map[string]*store.Player),
		}
	}

	linesInitialState = func() store.LinesState {
		return store.LinesState{
			Lines: make(map[int]*store.Line),
		}
	}

	mapInitialState = func() store.MapState {
		return store.MapState{
			Players: 2,
			Image:   store.MapImages[2],
		}
	}

	muActionsTested sync.Mutex
	actionsTested   = map[string]struct{}{
		// This are all the actions not involved on the Store
		action.CameraZoom.String():            {},
		action.CloseTowerMenu.String():        {},
		action.CursorMove.String():            {},
		action.DeselectTower.String():         {},
		action.GoHome.String():                {},
		action.NavigateTo.String():            {},
		action.OpenTowerMenu.String():         {},
		action.SelectTower.String():           {},
		action.SelectedTower.String():         {},
		action.SelectedTowerInvalid.String():  {},
		action.SignUpError.String():           {},
		action.ToggleStats.String():           {},
		action.ToggleStats.String():           {},
		action.ExitWaitingRoom.String():       {},
		action.JoinWaitingRoom.String():       {},
		action.UserSignIn.String():            {},
		action.UserSignOut.String():           {},
		action.UserSignUp.String():            {},
		action.WaitRoomCountdownTick.String(): {},
		action.WindowResizing.String():        {},
		action.SyncWaitingRoom.String():       {},
		action.SyncUsers.String():             {},
		action.VersionError.String():          {},
	}
)

func addAction(a string) {
	muActionsTested.Lock()
	defer muActionsTested.Unlock()

	actionsTested[a] = struct{}{}
}

func TestMain(m *testing.M) {
	code := m.Run()

	ma := make([]string, 0, 0)
	for _, a := range action.TypeValues() {
		if _, ok := actionsTested[a.String()]; !ok {
			ma = append(ma, a.String())
		}
	}
	if len(ma) != 0 {
		sort.Strings(ma)
		fmt.Printf("This actions are not tested: %s", ma)
		os.Exit(1)
	}
	os.Exit(code)
}

func TestEmpty(t *testing.T) {
	s := initStore()
	equalStore(t, s)
}

func TestStartGame(t *testing.T) {
	addAction(action.StartGame.String())
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p1 := addPlayer(s)
		p2 := addPlayer(s)
		p3 := addPlayer(s)

		s.Dispatch(action.NewStartGame())
		ms := mapInitialState()
		ms.Players = 3
		ms.Image = store.MapImages[3]

		ps := playersInitialState()
		ps.Players[p1.ID] = &p1
		ps.Players[p2.ID] = &p2
		ps.Players[p3.ID] = &p3

		ls := linesInitialState()
		for _, p := range ps.Players {
			x, y := s.Map.GetHomeCoordinates(p.LineID)
			g, err := graph.New(x+16, y+16, 16, 84, 16, 7, 74, 3)
			require.NoError(t, err)
			ls.Lines[p.LineID] = &store.Line{
				Towers: make(map[string]*store.Tower),
				Units:  make(map[string]*store.Unit),
				Graph:  g,
			}
		}

		equalStore(t, s, ps, ms, ls)
	})
}

func TestSummonUnit(t *testing.T) {
	addAction(action.SummonUnit.String())
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p2 := addPlayer(s)
		s.Dispatch(action.NewStartGame())
		ms, ls := startGame(t, s)

		eu := &store.Unit{
			// As the ID is a UUID we cannot guess it
			// ID: units[0].ID,
			MovingObject: utils.MovingObject{
				Object: utils.Object{
					// This is also random
					// X: units[0].X, Y: units[0].Y,
					W: 16, H: 16,
				},
				Facing: utils.Down,
			},
			Type:          unit.Spirit.String(),
			PlayerID:      p.ID,
			PlayerLineID:  p.LineID,
			CurrentLineID: p2.LineID,
			Health:        unit.Units[unit.Spirit.String()].Health,
		}

		a := action.NewSummonUnit(eu.Type, p.ID, p.LineID, p2.LineID)
		s.Dispatch(a)

		var uid string
		l := s.Lines.FindByID(p2.LineID)
		units := l.Units
		for id := range units {
			uid = id
		}

		// As this are random assigned we cannot expect them
		eu.ID, eu.X, eu.Y = units[uid].ID, units[uid].X, units[uid].Y

		// We need to set the path after the X, Y are set
		eu.Path = l.Graph.AStar(eu.X, eu.Y, eu.Facing, l.Graph.DeathNode.X, l.Graph.DeathNode.Y, unit.Units[eu.Type].Environment, atScale)
		eu.HashPath = graph.HashSteps(eu.Path)

		// As the Unit is created we remove it from the gold
		// and add more income
		p.Gold -= unit.Units[unit.Spirit.String()].Gold
		p.Income += unit.Units[unit.Spirit.String()].Income

		ps := playersInitialState()
		ps.Players[p.ID] = &p
		ps.Players[p2.ID] = &p2

		ls.Lines[p2.LineID].Units[eu.ID] = eu

		equalStore(t, s, ps, ls, ms)
	})
	t.Run("Do not reach negative gold", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p2 := addPlayer(s)
		s.Dispatch(action.NewStartGame())
		ms, ls := startGame(t, s)

		// We start with 40 gold, each Spirit
		// takes 10 gold so with that we can only
		// create 4 so we'll try to create 5
		for i := 0; i < 5; i++ {
			a := action.NewSummonUnit(unit.Spirit.String(), p.ID, p.LineID, p2.LineID)
			s.Dispatch(a)
		}

		// I don't want to EXPECT with Units just with Players
		l := s.Lines.FindByID(p2.LineID)
		units := l.Units

		assert.Equal(t, 4, len(units))

		// We could only create 4 of the 5 so -40
		p.Gold -= 40
		// Only 4 can be created not 5
		p.Income += 4

		ps := playersInitialState()
		ps.Players[p.ID] = &p
		ps.Players[p2.ID] = &p2

		ls = s.Lines.GetState().(store.LinesState)

		equalStore(t, s, ps, ms, ls)
	})
}

func TestTPS(t *testing.T) {
	addAction(action.TPS.String())
	t.Run("Success", func(t *testing.T) {
		t.Run("Default", func(t *testing.T) {
			s := initStore()
			p := addPlayer(s)
			p2 := addPlayer(s)
			s.Dispatch(action.NewStartGame())
			ms, ls := startGame(t, s)
			p, u := summonUnit(s, p, p2)

			s.Dispatch(action.NewTPS(time.Time{}))

			ps := playersInitialState()
			ps.Players[p.ID] = &p
			ps.Players[p2.ID] = &p2

			u.Path = u.Path[1:]
			u.MovingCount++
			ls.Lines[p2.LineID].Units[u.ID] = &u

			equalStore(t, s, ps, ms, ls)
		})
		t.Run("WithTime", func(t *testing.T) {
			s := initStore()
			p := addPlayer(s)
			p2 := addPlayer(s)
			s.Dispatch(action.NewStartGame())
			ms, ls := startGame(t, s)
			p, u := summonUnit(s, p, p2)

			l2 := s.Lines.GetState().(store.LinesState).Lines[p2.LineID]

			tn := time.Now()
			l2.UpdatedAt = tn
			l2.Units[u.ID].CreatedAt = tn

			ta := tn.Add(time.Second)
			s.Dispatch(action.NewTPS(ta))

			ps := playersInitialState()
			ps.Players[p.ID] = &p
			ps.Players[p2.ID] = &p2

			np := u.Path[61]
			u.Path = u.Path[62:]
			u.MovingCount += 62
			u.X = np.X
			u.Y = np.Y
			u.Facing = np.Facing
			ls.Lines[p2.LineID].Units[u.ID] = &u

			equalStore(t, s, ps, ms, ls)
		})
	})
}

func TestRemoveUnit(t *testing.T) {
	addAction(action.RemoveUnit.String())
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p2 := addPlayer(s)
		s.Dispatch(action.NewStartGame())
		ms, ls := startGame(t, s)
		p, u := summonUnit(s, p, p2)

		s.Dispatch(action.NewRemoveUnit(u.ID))

		ps := playersInitialState()
		ps.Players[p.ID] = &p
		ps.Players[p2.ID] = &p2

		equalStore(t, s, ps, ms, ls)
	})
}

func TestStealLive(t *testing.T) {
	addAction(action.StealLive.String())
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
	addAction(action.PlaceTower.String())
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p2 := addPlayer(s)
		s.Dispatch(action.NewStartGame())
		ms, ls := startGame(t, s)

		s.Dispatch(action.NewPlaceTower(tower.Soldier.String(), p.ID, 10, 20))

		p.Gold -= tower.Towers[tower.Soldier.String()].Gold

		ps := playersInitialState()
		ps.Players[p.ID] = &p
		ps.Players[p2.ID] = &p2

		tid := ""
		towers := s.Lines.FindByID(p.LineID).Towers
		for id := range towers {
			tid = id
		}

		tw := store.Tower{
			ID: tid,
			Object: utils.Object{
				X: 10, Y: 20,
				W: 32, H: 32,
			},
			Type:     tower.Soldier.String(),
			LineID:   p.LineID,
			PlayerID: p.ID,
		}
		ls.Lines[p.LineID].Towers[tw.ID] = &tw

		equalStore(t, s, ps, ms, ls)
	})
	t.Run("Do not reach negative gold", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p2 := addPlayer(s)
		s.Dispatch(action.NewStartGame())
		ms, ls := startGame(t, s)

		// The Player gold is 40 so it should only create 4
		// towers and not 10
		for i := 0; i <= 10; i++ {
			s.Dispatch(action.NewPlaceTower(tower.Soldier.String(), p.ID, 10, 20))
		}

		p.Gold = 0
		ps := playersInitialState()
		ps.Players[p.ID] = &p
		ps.Players[p2.ID] = &p2

		// I don't want to EXPECT with Towers just with Players
		ls = s.Lines.GetState().(store.LinesState)

		assert.Equal(t, 4, len(ls.Lines[p.LineID].Towers))

		equalStore(t, s, ps, ms, ls)
	})
	t.Run("Change unit course", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p2 := addPlayer(s)
		s.Dispatch(action.NewStartGame())
		ms, ls := startGame(t, s)
		p, u := summonUnit(s, p, p2)

		// We place it in the 10th (any place would be fine) path position so we can force the
		// unit to recalculate the path
		s.Dispatch(action.NewPlaceTower(tower.Soldier.String(), p.ID, u.Path[10].X, u.Path[10].Y))

		p.Gold -= tower.Towers[tower.Soldier.String()].Gold

		ps := playersInitialState()
		ps.Players[p.ID] = &p
		ps.Players[p2.ID] = &p2

		tid := ""
		towers := s.Lines.FindByID(p.LineID).Towers
		for id := range towers {
			tid = id
		}

		tw := store.Tower{
			ID: tid,
			Object: utils.Object{
				X: u.Path[10].X, Y: u.Path[10].Y,
				W: 32, H: 32,
			},
			Type:     tower.Soldier.String(),
			LineID:   p.LineID,
			PlayerID: p.ID,
		}
		l := ls.Lines[p.LineID]
		l.Towers[tw.ID] = &tw

		l2 := ls.Lines[p2.LineID]
		u.Path = l2.Graph.AStar(u.X, u.Y, u.Facing, l2.Graph.DeathNode.X, l2.Graph.DeathNode.Y, unit.Units[u.Type].Environment, atScale)
		u.HashPath = graph.HashSteps(u.Path)

		l2.Units[u.ID] = &u

		equalStore(t, s, ps, ms, ls)
	})
}

func TestRemoveTower(t *testing.T) {
	addAction(action.RemoveTower.String())
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p2 := addPlayer(s)
		s.Dispatch(action.NewStartGame())
		ms, ls := startGame(t, s)
		p, tw := placeTower(s, p)

		s.Dispatch(action.NewRemoveTower(p.ID, tw.ID, tw.Type))

		p.Gold += tower.Towers[tw.Type].Gold / 2

		ps := playersInitialState()
		ps.Players[p.ID] = &p
		ps.Players[p2.ID] = &p2

		equalStore(t, s, ps, ms, ls)
	})
}

func TestIncomeTick(t *testing.T) {
	addAction(action.IncomeTick.String())
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
	addAction(action.TowerAttack.String())
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p2 := addPlayer(s)
		s.Dispatch(action.NewStartGame())
		// TODO: Each summon/place updates the l.UpdatedAt so we should
		// manually add the value from the store line
		ms, ls := startGame(t, s)
		p, tw := placeTower(s, p)
		p, u := summonUnit(s, p, p2)

		s.Dispatch(action.NewTowerAttack(u.ID, tw.Type))
		u.Health -= tower.Towers[tw.Type].Damage

		ps := playersInitialState()
		ps.Players[p.ID] = &p
		ps.Players[p2.ID] = &p2

		l := ls.Lines[p.LineID]
		l.Towers[tw.ID] = &tw

		l2 := ls.Lines[p2.LineID]
		l2.Units[u.ID] = &u

		equalStore(t, s, ps, ms, ls)
	})
}

func TestUnitKilled(t *testing.T) {
	addAction(action.UnitKilled.String())
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p2 := addPlayer(s)
		s.Dispatch(action.NewStartGame())
		ms, ls := startGame(t, s)
		p, u := summonUnit(s, p, p2)

		s.Dispatch(action.NewUnitKilled(p.ID, u.Type))
		p.Gold += unit.Units[u.Type].Income

		ps := playersInitialState()
		ps.Players[p.ID] = &p
		ps.Players[p2.ID] = &p2

		l2 := ls.Lines[p2.LineID]
		l2.Units[u.ID] = &u

		equalStore(t, s, ps, ms, ls)
	})
}

func TestAddPlayer(t *testing.T) {
	addAction(action.AddPlayer.String())
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		id := uuid.Must(uuid.NewV4())
		name := "name"
		lid := 2
		s.Dispatch(action.NewAddPlayer(id.String(), name, lid))

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
		id := uuid.Must(uuid.NewV4())
		id2 := uuid.Must(uuid.NewV4())
		name := "name"
		lid := 2
		s.Dispatch(action.NewAddPlayer(id.String(), name, lid))
		s.Dispatch(action.NewAddPlayer(id2.String(), name, lid+1))

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
	addAction(action.RemovePlayer.String())
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p := addPlayer(s)
		p2 := addPlayer(s)
		s.Dispatch(action.NewStartGame())
		ms, ls := startGame(t, s)
		p, _ = placeTower(s, p)
		p, _ = summonUnit(s, p, p2)

		s.Dispatch(action.NewRemovePlayer(p.ID))

		p2.Winner = true
		ps := playersInitialState()
		ps.Players[p2.ID] = &p2

		delete(ls.Lines, p.LineID)

		equalStore(t, s, ps, ms, ls)
	})
}

func TestChangeUnitLine(t *testing.T) {
	addAction(action.ChangeUnitLine.String())
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		p1 := addPlayer(s)
		p2 := addPlayer(s)
		p3 := addPlayer(s)
		s.Dispatch(action.NewStartGame())
		ms, ls := startGame(t, s)
		p1, u1 := summonUnit(s, p1, p2)

		s.Dispatch(action.NewChangeUnitLine(u1.ID))

		ps := playersInitialState()
		ps.Players[p1.ID] = &p1
		ps.Players[p2.ID] = &p2
		ps.Players[p3.ID] = &p3

		u1.CurrentLineID = p3.LineID

		var uid string
		l := s.Lines.FindByID(p3.LineID)
		units := l.Units
		for id := range units {
			uid = id
		}

		// As this are random assigned we cannot expect them
		u1.ID, u1.X, u1.Y = units[uid].ID, units[uid].X, units[uid].Y
		u1.CreatedAt = units[uid].CreatedAt

		// We need to set the path after the X, Y are set
		u1.Path = l.Graph.AStar(u1.X, u1.Y, u1.Facing, l.Graph.DeathNode.X, l.Graph.DeathNode.Y, unit.Units[u1.Type].Environment, atScale)
		u1.HashPath = graph.HashSteps(u1.Path)

		ls.Lines[p3.LineID].Units[u1.ID] = &u1

		equalStore(t, s, ps, ms, ls)
	})
}

func TestSyncState(t *testing.T) {
	addAction(action.SyncState.String())
	t.Run("Success", func(t *testing.T) {
		s := initStore()
		ssa := &action.Action{
			Type: action.SyncState,
			SyncState: &action.SyncStatePayload{
				Players: &action.SyncStatePlayersPayload{
					Players: map[string]*action.SyncStatePlayerPayload{
						"123": &action.SyncStatePlayerPayload{
							ID:      "123",
							Name:    "Player name",
							Lives:   10,
							LineID:  2,
							Income:  3,
							Gold:    10,
							Current: true,
							Winner:  false,
						},
					},
					IncomeTimer: 5,
				},
				Lines: &action.SyncStateLinesPayload{
					Lines: map[int]*action.SyncStateLinePayload{
						1: &action.SyncStateLinePayload{
							Towers: map[string]*action.SyncStateTowerPayload{
								"456": &action.SyncStateTowerPayload{
									Object: utils.Object{
										X: 1, Y: 2, W: 3, H: 4,
									},
									ID:       "456",
									Type:     "soldier",
									PlayerID: "123",
									LineID:   2,
								},
							},
							Units: map[string]*action.SyncStateUnitPayload{
								"789": &action.SyncStateUnitPayload{
									ID:            "789",
									Type:          "cyclope",
									PlayerID:      "10",
									PlayerLineID:  10,
									CurrentLineID: 2,
									Health:        2,
								},
							},
						},
					},
				},
			},
		}
		ps := playersInitialState()
		ps.Players["123"] = &store.Player{
			ID:      "123",
			Name:    "Player name",
			Lives:   10,
			LineID:  2,
			Income:  3,
			Gold:    10,
			Current: true,
			Winner:  false,
		}
		ps.IncomeTimer = 5
		ls := linesInitialState()
		ls.Lines[1] = &store.Line{
			Towers: map[string]*store.Tower{
				"456": &store.Tower{
					Object: utils.Object{
						X: 1, Y: 2, W: 3, H: 4,
					},
					ID:       "456",
					Type:     "soldier",
					PlayerID: "123",
					LineID:   2,
				},
			},
			Units: map[string]*store.Unit{
				"789": &store.Unit{
					ID:            "789",
					Type:          "cyclope",
					PlayerID:      "10",
					PlayerLineID:  10,
					CurrentLineID: 2,
					Health:        2,
				},
			},
		}
		s.Dispatch(ssa)
		equalStore(t, s, ps, ls)
	})
}

func equalStore(t *testing.T, sto *store.Store, states ...interface{}) {
	t.Helper()

	pis := playersInitialState()
	lis := linesInitialState()
	mis := mapInitialState()
	for _, st := range states {
		switch s := st.(type) {
		case store.PlayersState:
			pis = s
		case store.LinesState:
			lis = s
		case store.MapState:
			mis = s
		default:
			t.Fatalf("State with type %T is unknown", st)
		}
	}

	assert.Equal(t, pis, sto.Players.GetState().(store.PlayersState))
	// We do not want to compare the Graphs as it's too big and
	// it takes ages to compare
	// At the end the Graph it's only used to calculate Paths but not
	// to have the Units/Towers init
	for _, l := range lis.Lines {
		l.Graph = nil
		l.UpdatedAt = time.Time{}
		for _, u := range l.Units {
			u.CreatedAt = time.Time{}
		}
	}
	for _, l := range sto.Lines.GetState().(store.LinesState).Lines {
		l.Graph = nil
		l.UpdatedAt = time.Time{}
		for _, u := range l.Units {
			u.CreatedAt = time.Time{}
		}
	}
	assert.Equal(t, lis, sto.Lines.GetState().(store.LinesState))
	assert.Equal(t, mis, sto.Map.GetState().(store.MapState))
}
