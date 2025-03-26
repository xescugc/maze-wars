package client

import (
	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/maze-wars/action"
)

type UserStore struct {
	*flux.ReduceStore[UserState, *action.Action]
}

type UserState struct {
	Username string
	ImageKey string
}

func NewUserStore(d *flux.Dispatcher[*action.Action]) *UserStore {
	u := &UserStore{}
	u.ReduceStore = flux.NewReduceStore(d, u.Reduce, UserState{})

	return u
}

func (us *UserStore) Username() string { return us.GetState().Username }
func (us *UserStore) ImageKey() string { return us.GetState().ImageKey }

func (u *UserStore) Reduce(state UserState, act *action.Action) UserState {
	switch act.Type {
	case action.UserSignIn:
		state.Username = act.UserSignIn.Username
		state.ImageKey = act.UserSignIn.ImageKey
	}

	return state
}
