package server

import (
	"log/slog"

	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/store"
)

type Game struct {
	*store.Store
}

func NewGame(d *flux.Dispatcher, l *slog.Logger) *Game {
	g := &Game{
		Store: store.NewStore(d, l),
	}

	return g
}
