package server

import (
	"sync"

	"github.com/gofrs/uuid"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"nhooyr.io/websocket"
)

type RoomsStore struct {
	*flux.ReduceStore

	Store *Store

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

func NewRoomsStore(d *flux.Dispatcher, s *Store) *RoomsStore {
	rs := &RoomsStore{
		Store: s,
	}

	rs.ReduceStore = flux.NewReduceStore(d, rs.Reduce, RoomsState{
		Rooms: make(map[string]*Room),
	})

	return rs
}

func (rs *RoomsStore) List() []*Room {
	rs.mxRooms.RLock()
	defer rs.mxRooms.RUnlock()

	mrooms := rs.GetState().(RoomsState)
	rooms := make([]*Room, 0, len(mrooms.Rooms))
	for _, r := range mrooms.Rooms {
		rooms = append(rooms, r)
	}
	return rooms
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
		rd := flux.NewDispatcher()
		g := NewGame(rd)
		rstate.Rooms[act.StartGame.Room].Game = g
		// TODO:
		pcount := 0
		for pid, pc := range rstate.Rooms[act.StartGame.Room].Players {
			u, _ := rs.Store.Users.FindByRemoteAddress(pc.RemoteAddr)
			g.Dispatch(action.NewAddPlayer(pid, u.Username, pcount))
			pcount++
		}

	case action.RemovePlayer:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		pc := rstate.Rooms[act.RemovePlayer.Room].Players[act.RemovePlayer.ID]
		delete(rstate.Rooms[act.RemovePlayer.Room].Players, act.RemovePlayer.ID)
		delete(rstate.Rooms[act.RemovePlayer.Room].Connections, pc.RemoteAddr)

		rstate.Rooms[act.Room].Game.Dispatch(act)

		if len(rstate.Rooms[act.Room].Players) == 0 {
			delete(rstate.Rooms, act.Room)
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

		if len(wr.Players) == wr.Size {
			// As the size has been reached we remove
			// the current room as WR
			rstate.CurrentWaitingRoom = ""
		}
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

		if wr.Size == len(wr.Players) {
			rstate.CurrentWaitingRoom = ""
		}
	case action.ExitWaitingRoom:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		us, _ := rs.Store.Users.FindByUsername(act.ExitWaitingRoom.Username)
		delete(rstate.Rooms[rstate.CurrentWaitingRoom].Players, us.ID)

		// If there are no more players waiting remove the room
		if len(rstate.Rooms[rstate.CurrentWaitingRoom].Players) == 0 {
			delete(rstate.Rooms, rstate.CurrentWaitingRoom)
			rstate.CurrentWaitingRoom = ""
		}

	default:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		if r, ok := rstate.Rooms[act.Room]; ok {
			r.Game.Dispatch(act)
		}
		// If no room means that is a broadcast
		if act.Room == "" {
			for _, r := range rstate.Rooms {
				if r.Name != rstate.CurrentWaitingRoom {
					r.Game.Dispatch(act)
				}
			}
		}
	}

	return rstate
}
