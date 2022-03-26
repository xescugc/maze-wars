package main

import (
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
)

type ActionDispatcher struct {
	dispatcher *flux.Dispatcher
}

func NewActionDispatcher(d *flux.Dispatcher) *ActionDispatcher {
	return &ActionDispatcher{
		dispatcher: d,
	}
}

func (ac *ActionDispatcher) CameraMove(x, y int) {
	cma := action.NewCameraMove(x, y)
	ac.dispatcher.Dispatch(cma)
}
