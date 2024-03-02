package server

import (
	"log/slog"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"nhooyr.io/websocket"
)

type RoomsStore struct {
	*flux.ReduceStore

	Store *Store

	logger  *slog.Logger
	mxRooms sync.RWMutex
}

type RoomsState struct {
	Rooms              map[string]*Room
	CurrentWaitingRoom string
}

type Room struct {
	Name string

	Players map[string]PlayerConn

	Connections map[string]string

	Size      int
	Countdown int

	Game *Game
}

type PlayerConn struct {
	Conn       *websocket.Conn
	RemoteAddr string
}

func NewRoomsStore(d *flux.Dispatcher, s *Store, l *slog.Logger) *RoomsStore {
	rs := &RoomsStore{
		Store:  s,
		logger: l,
	}

	rs.ReduceStore = flux.NewReduceStore(d, rs.Reduce, RoomsState{
		Rooms: make(map[string]*Room),
	})

	return rs
}

func (rs *RoomsStore) List() []*Room {
	rs.mxRooms.RLock()
	defer rs.mxRooms.RUnlock()

	srooms := rs.GetState().(RoomsState)
	rooms := make([]*Room, 0, len(srooms.Rooms))
	for _, r := range srooms.Rooms {
		rooms = append(rooms, r)
	}
	return rooms
}

func (rs *RoomsStore) FindCurrentWaitingRoom() *Room {
	rs.mxRooms.RLock()
	defer rs.mxRooms.RUnlock()

	srooms := rs.GetState().(RoomsState)
	r, ok := srooms.Rooms[srooms.CurrentWaitingRoom]
	if !ok {
		return nil
	}
	return r
}

func (rs *RoomsStore) GetNextID(room string) int {
	r, _ := rs.GetState().(RoomsState).Rooms[room]
	return len(r.Players)
}

func (rs *RoomsStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	rstate, ok := state.(RoomsState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.StartGame:
		rs.GetDispatcher().WaitFor(rs.Store.Users.GetDispatcherToken())

		rd := flux.NewDispatcher()
		g := NewGame(rd, rs.logger)
		rstate.Rooms[rstate.CurrentWaitingRoom].Game = g
		pcount := 0
		for pid, pc := range rstate.Rooms[rstate.CurrentWaitingRoom].Players {
			u, _ := rs.Store.Users.FindByRemoteAddress(pc.RemoteAddr)
			g.Dispatch(action.NewAddPlayer(pid, u.Username, pcount))
			pcount++
		}
		rstate.CurrentWaitingRoom = ""
		g.Dispatch(action.NewStartGame())

	case action.RemovePlayer:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		removePlayer(&rstate, act.RemovePlayer.ID, act.Room)
	case action.UserSignOut:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		u, ok := rs.Store.Users.FindByUsername(act.UserSignOut.Username)
		if ok && u.CurrentRoomID != "" {
			removePlayer(&rstate, u.ID, u.CurrentRoomID)
		}
	case action.JoinWaitingRoom:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		if rstate.CurrentWaitingRoom == "" {
			rid := uuid.Must(uuid.NewV4())
			rstate.Rooms[rid.String()] = &Room{
				Name:        rid.String(),
				Players:     make(map[string]PlayerConn),
				Connections: make(map[string]string),

				Size:      6,
				Countdown: 10,
			}
			rstate.CurrentWaitingRoom = rid.String()
		}

		us, _ := rs.Store.Users.FindByUsername(act.JoinWaitingRoom.Username)
		wr := rstate.Rooms[rstate.CurrentWaitingRoom]
		wr.Players[us.ID] = PlayerConn{
			Conn:       us.Conn,
			RemoteAddr: us.RemoteAddr,
		}
		wr.Connections[us.RemoteAddr] = us.ID
	case action.WaitRoomCountdownTick:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		if rstate.CurrentWaitingRoom == "" {
			break
		}

		wr := rstate.Rooms[rstate.CurrentWaitingRoom]
		wr.Countdown--
		if wr.Countdown == -1 {
			if wr.Size > 2 {
				wr.Countdown = 10
				wr.Size--
			} else {
				wr.Countdown = 0
			}
		}
	case action.ExitWaitingRoom:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		us, _ := rs.Store.Users.FindByUsername(act.ExitWaitingRoom.Username)
		delete(rstate.Rooms[rstate.CurrentWaitingRoom].Players, us.ID)
		delete(rstate.Rooms[rstate.CurrentWaitingRoom].Connections, us.RemoteAddr)

		// If there are no more players waiting remove the room
		if len(rstate.Rooms[rstate.CurrentWaitingRoom].Players) == 0 {
			delete(rstate.Rooms, rstate.CurrentWaitingRoom)
			rstate.CurrentWaitingRoom = ""
		}

	default:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		// If no room means that is a broadcast
		if act.Room == "" {
			for _, r := range rstate.Rooms {
				if r.Name != rstate.CurrentWaitingRoom {
					go r.Game.Dispatch(act)
				}
			}
		} else {
			if r, ok := rstate.Rooms[act.Room]; ok {
				go r.Game.Dispatch(act)
			}
		}
	}

	return rstate
}

func removePlayer(rstate *RoomsState, pid, room string) {
	pc := rstate.Rooms[room].Players[pid]
	delete(rstate.Rooms[room].Players, pid)
	delete(rstate.Rooms[room].Connections, pc.RemoteAddr)

	rstate.Rooms[room].Game.Dispatch(action.NewRemovePlayer(pid))

	if len(rstate.Rooms[room].Players) == 0 {
		delete(rstate.Rooms, room)
	}
}
