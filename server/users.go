package server

import (
	"sync"

	"github.com/gofrs/uuid"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"nhooyr.io/websocket"
)

type UsersStore struct {
	*flux.ReduceStore

	Store *Store

	mxUsers sync.RWMutex
}

type UsersState struct {
	Users map[string]*User
}

type User struct {
	ID       string
	Username string

	Conn       *websocket.Conn
	RemoteAddr string
}

func NewUsersStore(d *flux.Dispatcher) *UsersStore {
	us := &UsersStore{}

	us.ReduceStore = flux.NewReduceStore(d, us.Reduce, UsersState{
		Users: make(map[string]*User),
	})

	return us
}

func (us *UsersStore) FindByUsername(un string) (User, bool) {
	us.mxUsers.RLock()
	defer us.mxUsers.RUnlock()

	u, ok := us.GetState().(UsersState).Users[un]
	if !ok {
		return User{}, false
	}
	return *u, true
}

func (us *UsersStore) FindByRemoteAddress(ra string) (User, bool) {
	us.mxUsers.RLock()
	defer us.mxUsers.RUnlock()

	for _, u := range us.GetState().(UsersState).Users {
		if u.RemoteAddr == ra {
			return *u, true
			break
		}
	}
	return User{}, false
}

func (us *UsersStore) List() []*User {
	us.mxUsers.RLock()
	defer us.mxUsers.RUnlock()

	musers := us.GetState().(UsersState)
	users := make([]*User, 0, len(musers.Users))
	for _, u := range musers.Users {
		users = append(users, u)
	}
	return users
}

func (us *UsersStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	ustate, ok := state.(UsersState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.UserSignUp:
		us.mxUsers.Lock()
		defer us.mxUsers.Unlock()

		id := uuid.Must(uuid.NewV4())
		ustate.Users[act.UserSignUp.Username] = &User{
			ID:       id.String(),
			Username: act.UserSignUp.Username,
		}
	case action.UserSignIn:
		us.mxUsers.Lock()
		defer us.mxUsers.Unlock()

		if u, ok := ustate.Users[act.UserSignIn.Username]; ok {
			u.Conn = act.UserSignIn.Websocket
			u.RemoteAddr = act.UserSignIn.RemoteAddr
		}
	case action.UserSignOut:
		us.mxUsers.Lock()
		defer us.mxUsers.Unlock()

		delete(ustate.Users, act.UserSignOut.Username)
	}

	return ustate
}
