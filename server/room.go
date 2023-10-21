package main

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
)

type RoomsStore struct {
	*flux.ReduceStore
}

type RoomsState struct {
	Rooms map[string]*Room
}

type Room struct {
	Name string

	muPlayers sync.RWMutex
	Players   map[string]*websocket.Conn

	Connections map[string]string

	Game *Game
}

func NewRoomsStore(d *flux.Dispatcher) *RoomsStore {
	rs := &RoomsStore{}

	rs.ReduceStore = flux.NewReduceStore(d, rs.Reduce, RoomsState{Rooms: make(map[string]*Room)})

	return rs
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
	case action.JoinRoom:
		rd := flux.NewDispatcher()
		if _, ok := rstate.Rooms[act.JoinRoom.Room]; !ok {
			rstate.Rooms[act.JoinRoom.Room] = &Room{
				Name:        act.JoinRoom.Room,
				Players:     make(map[string]*websocket.Conn),
				Connections: make(map[string]string),
				Game:        NewGame(rd),
			}
		}
	case action.AddPlayer:
		if len(rstate.Rooms[act.AddPlayer.Room].Players) == 6 {
			// The limit for now will be to 6 players but realistically
			// it could have no limit
			break
		}
		rstate.Rooms[act.AddPlayer.Room].Players[act.AddPlayer.ID] = act.AddPlayer.Websocket
		rstate.Rooms[act.AddPlayer.Room].Connections[act.AddPlayer.Websocket.RemoteAddr().String()] = act.AddPlayer.ID

		rstate.Rooms[act.Room].Game.Dispatch(act)
	case action.RemovePlayer:
		ws := rstate.Rooms[act.RemovePlayer.Room].Players[act.RemovePlayer.ID]
		delete(rstate.Rooms[act.RemovePlayer.Room].Players, act.RemovePlayer.ID)
		delete(rstate.Rooms[act.RemovePlayer.Room].Connections, ws.RemoteAddr().String())

		rstate.Rooms[act.Room].Game.Dispatch(act)

		if len(rstate.Rooms[act.Room].Players) == 0 {
			delete(rstate.Rooms, act.Room)
		}
	default:
		if r, ok := rstate.Rooms[act.Room]; ok {
			r.Game.Dispatch(act)
		}
		// If no room means that is a broadcast
		if act.Room == "" {
			for _, r := range rstate.Rooms {
				r.Game.Dispatch(act)
			}
		}
	}

	return rstate
}
