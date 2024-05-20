package store

import (
	"time"

	"github.com/sagikazarmark/slog-shim"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/utils"
)

type Store struct {
	Lines   *Lines
	Map     *Map
	Lobbies *Lobbies

	dispatcher *flux.Dispatcher
	logger     *slog.Logger
}

func NewStore(d *flux.Dispatcher, l *slog.Logger) *Store {
	s := &Store{
		dispatcher: d,
		logger:     l,
	}
	s.Map = NewMap(d, s)
	s.Lines = NewLines(d, s)
	s.Lobbies = NewLobbies(d)

	return s
}

func (s *Store) Dispatch(a *action.Action) {
	b := time.Now()
	defer utils.LogTime(s.logger, b, "action dispatch", "action", a.Type)

	s.dispatcher.Dispatch(a)
}
