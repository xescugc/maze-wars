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
	s := &Store{
		Players: NewPlayers(d),

		dispatcher: d,
	}
	s.Map = NewMap(d, s)
	s.Towers = NewTowers(d, s)
	s.Units = NewUnits(d, s)

	return s
}

func (s *Store) Dispatch(a *action.Action) {
	s.dispatcher.Dispatch(a)
}
