package store

import (
	"time"

	"github.com/sagikazarmark/slog-shim"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/utils"
)

type Store struct {
	Players *Players
	Towers  *Towers
	Units   *Units
	Map     *Map

	dispatcher *flux.Dispatcher
	logger     *slog.Logger
}

func NewStore(d *flux.Dispatcher, l *slog.Logger) *Store {
	s := &Store{
		dispatcher: d,
		logger:     l,
	}
	s.Players = NewPlayers(d, s)
	s.Map = NewMap(d, s)
	s.Towers = NewTowers(d, s)
	s.Units = NewUnits(d, s)

	return s
}

func (s *Store) Dispatch(a *action.Action) {
	b := time.Now()
	defer utils.LogTime(s.logger, b, "action dispatch", "action", a.Type)

	s.dispatcher.Dispatch(a)
}
