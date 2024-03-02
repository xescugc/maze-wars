package store_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/sagikazarmark/slog-shim"
	"github.com/stretchr/testify/require"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/unit"
	"github.com/xescugc/maze-wars/utils/graph"
)

func newEmptyLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func initStore() *store.Store {
	d := flux.NewDispatcher()
	return store.NewStore(d, newEmptyLogger())
}

func addPlayer(s *store.Store) store.Player {
	id := uuid.Must(uuid.NewV4())
	name := fmt.Sprintf("name-%d", len(s.Players.List()))
	lid := len(s.Players.List())
	s.Dispatch(action.NewAddPlayer(id.String(), name, lid))

	return s.Players.FindByID(id.String())
}

func startGame(t *testing.T, s *store.Store) (store.MapState, store.LinesState) {
	t.Helper()

	ms := mapInitialState()
	ms.Players = len(s.Players.List())
	ms.Image = store.MapImages[ms.Players]

	ls := linesInitialState()
	for _, p := range s.Players.List() {
		x, y := s.Map.GetHomeCoordinates(p.LineID)
		g, err := graph.New(int(x+16), int(y+16), 16, 84, 16, 7, 74, 3)
		require.NoError(t, err)
		ls.Lines[p.LineID] = &store.Line{
			Towers: make(map[string]*store.Tower),
			Units:  make(map[string]*store.Unit),
			Graph:  g,
		}
	}

	return ms, ls
}

func summonUnit(s *store.Store, fp, tp store.Player) (store.Player, store.Unit) {
	s.Dispatch(action.NewSummonUnit(unit.Spirit.String(), fp.ID, fp.LineID, tp.LineID))

	// We know the Summon does this and as 'p' is not a pointer
	// we need to do it manually
	fp.Gold -= unit.Units[unit.Spirit.String()].Gold
	fp.Income += unit.Units[unit.Spirit.String()].Income

	units := s.Lines.FindByID(tp.LineID).Units
	var u *store.Unit
	for _, un := range units {
		u = un
	}

	return fp, *u
}

func placeTower(s *store.Store, p store.Player) (store.Player, store.Tower) {
	l := s.Lines.FindByID(p.LineID)

	s.Dispatch(action.NewPlaceTower(tower.Soldier.String(), p.ID, l.Graph.OffsetX, l.Graph.OffsetY+(l.Graph.Scale*l.Graph.SpawnZoneH)))

	// We know the PlaceTower does this and as 'p' is not a pointer
	// we need to do it manually
	p.Gold -= tower.Towers[tower.Soldier.String()].Gold

	// We cannot reuse the 'l' as the FindByID returns a copy
	// so it's not updated when the NewPlaceTower is triggered
	towers := s.Lines.FindByID(p.LineID).Towers
	var tw *store.Tower
	for _, tn := range towers {
		tw = tn
	}

	return p, *tw
}
