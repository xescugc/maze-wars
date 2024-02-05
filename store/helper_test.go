package store_test

import (
	"fmt"
	"io"

	"github.com/gofrs/uuid"
	"github.com/sagikazarmark/slog-shim"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/unit"
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

func summonUnit(s *store.Store, p store.Player) (store.Player, store.Unit) {
	s.Dispatch(action.NewSummonUnit(unit.Spirit.String(), p.ID, p.LineID, p.LineID))

	// We know the Summon does this and as 'p' is not a pointer
	// we need to do it manually
	p.Gold -= unit.Units[unit.Spirit.String()].Gold
	p.Income += unit.Units[unit.Spirit.String()].Income

	return p, *s.Units.List()[0]
}

func placeTower(s *store.Store, p store.Player) (store.Player, store.Tower) {
	s.Dispatch(action.NewPlaceTower(tower.Soldier.String(), p.ID, 10, 20))

	// We know the PlaceTower does this and as 'p' is not a pointer
	// we need to do it manually
	p.Gold -= tower.Towers[tower.Soldier.String()].Gold

	return p, *s.Towers.List()[0]
}
