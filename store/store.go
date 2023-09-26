package store

import (
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
)

type Store struct {
	Players *Players
	Towers  *Towers
	Units   *Units
	Map     *Map

	dispatcher *flux.Dispatcher
}

func NewStore(d *flux.Dispatcher) *Store {
	g := &Store{
		Players: NewPlayers(d),

		dispatcher: d,
	}
	// TODO: Handle error?
	g.Map, _ = NewMap()
	g.Towers = NewTowers(d, g)
	g.Units = NewUnits(d, g)

	return g
}

func (g *Store) Dispatch(a *action.Action) {
	g.dispatcher.Dispatch(a)
}
