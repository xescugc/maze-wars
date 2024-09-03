package server

import (
	"log/slog"

	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/store"
)

type Store struct {
	Rooms   *RoomsStore
	Users   *UsersStore
	Lobbies *store.Lobbies
}

func NewStore(d *flux.Dispatcher, ws WSConnector, l *slog.Logger) *Store {
	ss := &Store{}

	rooms := NewRoomsStore(d, ss, ws, l)
	users := NewUsersStore(d, ss)
	lobbies := store.NewLobbies(d)

	ss.Rooms = rooms
	ss.Users = users
	ss.Lobbies = lobbies

	return ss
}
