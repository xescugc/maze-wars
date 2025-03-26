package server

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/store"
)

type Store struct {
	Rooms   *RoomsStore
	Lobbies *store.Lobbies
}

func NewStore(d *flux.Dispatcher[*action.Action], ws WSConnector, dgo *discordgo.Session, opt Options, l *slog.Logger) *Store {
	ss := &Store{}

	rooms := NewRoomsStore(d, ss, ws, dgo, opt, l)
	lobbies := store.NewLobbies(d)

	ss.Rooms = rooms
	ss.Lobbies = lobbies

	return ss
}
