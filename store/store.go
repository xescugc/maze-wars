package store

import (
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
)

type Store struct {
	Players *Players
	Towers  *Towers
	Units   *Units
	Map     *Map

	dispatcher *flux.Dispatcher
}

func NewStore(d *flux.Dispatcher) *Store {
	s := &Store{
		dispatcher: d,
	}
	s.Players = NewPlayers(d, s)
	s.Map = NewMap(d, s)
	s.Towers = NewTowers(d, s)
	s.Units = NewUnits(d, s)

	return s
}

func (s *Store) Dispatch(a *action.Action) {
	s.dispatcher.Dispatch(a)
}
