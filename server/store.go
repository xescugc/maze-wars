package server

import (
	"log/slog"

	"github.com/xescugc/go-flux"
)

type Store struct {
	Rooms *RoomsStore
	Users *UsersStore
}

func NewStore(d *flux.Dispatcher, l *slog.Logger) *Store {
	ss := &Store{}

	rooms := NewRoomsStore(d, ss, l)
	users := NewUsersStore(d, ss)

	ss.Rooms = rooms
	ss.Users = users

	return ss
}
