package client

import (
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
)

type UserStore struct {
	*flux.ReduceStore
}

type UserState struct {
	Username string
}

func NewUserStore(d *flux.Dispatcher) *UserStore {
	u := &UserStore{}
	u.ReduceStore = flux.NewReduceStore(d, u.Reduce, UserState{})

	return u
}

func (us *UserStore) Username() string { return us.GetState().(UserState).Username }

func (u *UserStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	ustate, ok := state.(UserState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.UserSignIn:
		ustate.Username = act.UserSignIn.Username
	}

	return ustate
}
