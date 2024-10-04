package server

import (
	"log/slog"

	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/store"
)

type Store struct {
	Rooms   *RoomsStore
	Lobbies *store.Lobbies
}

func NewStore(d *flux.Dispatcher, ws WSConnector, l *slog.Logger) *Store {
	ss := &Store{}

	rooms := NewRoomsStore(d, ss, ws, l)
	lobbies := store.NewLobbies(d)

	ss.Rooms = rooms
	ss.Lobbies = lobbies

	return ss
}
