package server

import (
	"log/slog"

	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/store"
)

const isOnServer = true

type Game struct {
	*store.Store
}

func NewGame(d *flux.Dispatcher[*action.Action], l *slog.Logger) *Game {
	g := &Game{
		Store: store.NewStore(d, l, isOnServer),
	}

	return g
}
