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

	CurrentRoomID  string
	CurrentLobbyID string
}

func NewUsersStore(d *flux.Dispatcher, s *Store) *UsersStore {
	us := &UsersStore{
		Store: s,
	}

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
		us.GetDispatcher().WaitFor(us.Store.Rooms.GetDispatcherToken())

		us.mxUsers.Lock()
		defer us.mxUsers.Unlock()

		delete(ustate.Users, act.UserSignOut.Username)
	case action.StartRoom:
		us.mxUsers.Lock()
		defer us.mxUsers.Unlock()

		r := us.Store.Rooms.FindByID(act.StartRoom.RoomID)
		if r != nil {
			for pid := range r.Players {
				for _, u := range ustate.Users {
					if u.ID == pid {
						u.CurrentRoomID = r.Name
					}
				}
			}
		}
	case action.RemovePlayer:
		us.mxUsers.Lock()
		defer us.mxUsers.Unlock()

		for _, u := range ustate.Users {
			if u.ID == act.RemovePlayer.ID {
				u.CurrentRoomID = ""
				break
			}
		}

	// Lobby actions
	case action.JoinLobby:
		us.mxUsers.Lock()
		defer us.mxUsers.Unlock()

		if act.JoinLobby.IsBot {
			break
		}
		ustate.Users[act.JoinLobby.Username].CurrentLobbyID = act.JoinLobby.LobbyID
	case action.LeaveLobby:
		us.mxUsers.Lock()
		defer us.mxUsers.Unlock()

		if un, ok := ustate.Users[act.LeaveLobby.Username]; ok {
			un.CurrentLobbyID = ""
		}
	case action.DeleteLobby:
		us.mxUsers.Lock()
		defer us.mxUsers.Unlock()

		// TODO: Potentially make it better if this could have access to the
		// lobby and just target the users that we know need this.
		// It has access but it would need a WaitFor in order so synchronize
		for _, u := range ustate.Users {
			if u.CurrentLobbyID == act.DeleteLobby.LobbyID {
				u.CurrentLobbyID = ""
			}
		}
	case action.CreateLobby:
		ustate.Users[act.CreateLobby.Owner].CurrentLobbyID = act.CreateLobby.LobbyID
	}

	return ustate
}
