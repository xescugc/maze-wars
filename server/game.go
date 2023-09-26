package main

import (
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/store"
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
