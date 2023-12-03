package store_test

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
	"github.com/xescugc/ltw/store"
	"github.com/xescugc/ltw/tower"
	"github.com/xescugc/ltw/unit"
)

func initStore() *store.Store {
	d := flux.NewDispatcher()
	return store.NewStore(d)
}

func addPlayer(s *store.Store) store.Player {
	sid := "sid"
	id := uuid.Must(uuid.NewV4())
	name := fmt.Sprintf("name-%s", id.String())
	lid := 2
	ws := &websocket.Conn{}
	s.Dispatch(action.NewAddPlayer(sid, id.String(), name, lid, ws))

	return s.Players.FindByID(id.String())
}

func summonUnit(s *store.Store, p store.Player) (store.Player, store.Unit) {
	clid := 2
	s.Dispatch(action.NewSummonUnit(unit.Spirit.String(), p.ID, p.LineID, clid))

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
