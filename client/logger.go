package main

import (
	"fmt"

	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
)

// LoggerStore is a logger store in charge of logging all the actions
type LoggerStore struct {
	*flux.ReduceStore
}

// NewLoggerStore creates a new LoggerStore with the Dispatcher d
func NewLoggerStore(d *flux.Dispatcher) *LoggerStore {
	ss := &LoggerStore{}
	ss.ReduceStore = flux.NewReduceStore(d, ss.Reduce, nil)

	return ss
}

func (ss *LoggerStore) Reduce(cstate, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return cstate
	}

	// Prints all the action types
	fmt.Println(act.Type)

	return cstate
}
