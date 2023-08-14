package main

import (
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
)

type ScreenStore struct {
	*flux.ReduceStore
}

type ScreenState struct {
	W    int
	H    int
	Zoom int
}

func NewScreenStore(d *flux.Dispatcher, w, h int) *ScreenStore {
	ss := &ScreenStore{}
	ss.ReduceStore = flux.NewReduceStore(d, ss.Reduce, ScreenState{
		W: w,
		H: h,
	})

	return ss
}

func (ss *ScreenStore) GetWidth() int {
	s := ss.GetState().(ScreenState)
	return s.W + s.Zoom
}
func (ss *ScreenStore) GetHeight() int {
	s := ss.GetState().(ScreenState)
	return s.H + s.Zoom
}

func (ss *ScreenStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	sstate, ok := state.(ScreenState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.CameraZoom:
		sstate.Zoom += act.CameraZoom.Direction
	default:
	}
	return sstate
}
