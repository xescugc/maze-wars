package main

import (
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
)

type ScreenStore struct {
	*flux.ReduceStore
}

type ScreenState struct {
	W int
	H int
}

func NewScreenStore(d *flux.Dispatcher, w, h int) *ScreenStore {
	ss := &ScreenStore{}
	ss.ReduceStore = flux.NewReduceStore(d, ss.Reduce, ScreenState{
		W: w,
		H: h,
	})

	return ss
}

func (ss *ScreenStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	cstate, ok := state.(ScreenState)
	if !ok {
		return state
	}

	switch act.Type {
	default:
	}
	return cstate
}
