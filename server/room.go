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

	muCons  sync.RWMutex
	Players map[string]*websocket.Conn

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
				Name:    act.JoinRoom.Room,
				Players: make(map[string]*websocket.Conn),
				Game:    NewGame(rd),
			}
		}
	case action.AddPlayer:
		rstate.Rooms[act.AddPlayer.Room].Players[act.AddPlayer.ID] = act.AddPlayer.Websocket
		fallthrough
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
