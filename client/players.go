package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/ltw/store"
)

var (
	unitIncome = map[string]int{
		"cyclope": 1,
	}
)

type Players struct {
	store *store.Store
}

func NewPlayers(s *store.Store) *Players {
	return &Players{
		store: s,
	}
}

func (ps *Players) GetCurrentPlayer() store.Player {
	for _, p := range ps.store.Players.GetState().(store.PlayersState).Players {
		if p.Current {
			return *p
		}
	}
	return store.Player{}
}

func (ps *Players) Update() error {
	return nil
}

func (ps *Players) Draw(screen *ebiten.Image) {}
