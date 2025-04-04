package store

import (
	"time"

	"github.com/sagikazarmark/slog-shim"
	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/utils"
)

type Store struct {
	Game    *Game
	Map     *Map
	Lobbies *Lobbies

	dispatcher *flux.Dispatcher[*action.Action]
	logger     *slog.Logger

	isOnServer bool
}

func NewStore(d *flux.Dispatcher[*action.Action], l *slog.Logger, server bool) *Store {
	s := &Store{
		dispatcher: d,
		logger:     l,
		isOnServer: server,
	}
	s.Map = NewMap(d, s)
	s.Game = NewGame(d, s)
	s.Lobbies = NewLobbies(d)

	return s
}

func (s *Store) Dispatch(a *action.Action) {
	b := time.Now()
	defer utils.LogTime(s.logger, b, "action dispatch", "action", a.Type)

	s.dispatcher.Dispatch(a)
}
