package server

import (
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/store"
)

type Game struct {
	*store.Store
}

func NewGame(d *flux.Dispatcher) *Game {
	g := &Game{
		Store: store.NewStore(d),
	}

	return g
}
