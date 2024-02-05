package server_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/server"
	"nhooyr.io/websocket"
)

var (
	roomsInitialState = func() server.RoomsState {
		return server.RoomsState{
			Rooms: make(map[string]*server.Room),
		}
	}

	usersInitialState = func() server.UsersState {
		return server.UsersState{
			Users: make(map[string]*server.User),
		}
	}
)

func TestEmpty(t *testing.T) {
	_, s := initStore()
	equalStore(t, s)
}

func TestUserSignUp(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ad, s := initStore()
		un := "user name"

		ad.UserSignUp(un)

		u := &server.User{
			ID:       s.Users.List()[0].ID,
			Username: un,
		}
		us := usersInitialState()
		us.Users[un] = u
		equalStore(t, s, us)
	})
}

func TestUserSignIn(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ad, s := initStore()
		un := "user name"
		ra := "remote-address"
		ws := &websocket.Conn{}

		ad.UserSignUp(un)
		ad.UserSignIn(un, ra, ws)

		u := &server.User{
			ID:         s.Users.List()[0].ID,
			Username:   un,
			Conn:       ws,
			RemoteAddr: ra,
		}
		us := usersInitialState()
		us.Users[un] = u
		equalStore(t, s, us)
	})
}

func TestUserSignOut(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ad, s := initStore()
		un := "user name"

		ad.UserSignUp(un)
		ad.UserSignOut(un)

		equalStore(t, s)
	})
}

func TestJoinWaitingRoom(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ad, s := initStore()
		un := "user name"
		ra := "remote-address"
		ws := &websocket.Conn{}

		ad.UserSignUp(un)
		ad.UserSignIn(un, ra, ws)

		a := action.NewJoinWaitingRoom(un)
		ad.Dispatch(a)

		u := &server.User{
			ID:         s.Users.List()[0].ID,
			Username:   un,
			Conn:       ws,
			RemoteAddr: ra,
		}
		us := usersInitialState()
		us.Users[un] = u

		rs := roomsInitialState()
		cwrid := s.Rooms.FindCurrentWaitingRoom().Name
		rs.CurrentWaitingRoom = cwrid
		rs.Rooms[cwrid] = &server.Room{
			Name: cwrid,
			Players: map[string]server.PlayerConn{
				u.ID: server.PlayerConn{
					Conn:       ws,
					RemoteAddr: ra,
				},
			},
			Connections: map[string]string{
				ra: u.ID,
			},

			Size:      6,
			Countdown: 10,
		}

		equalStore(t, s, us, rs)
	})
	t.Run("Success2Players", func(t *testing.T) {
		ad, s := initStore()
		un := "user name"
		ra := "remote-address"
		un2 := "user name2"
		ra2 := "remote-address2"
		ws := &websocket.Conn{}

		ad.UserSignUp(un)
		ad.UserSignIn(un, ra, ws)
		ad.UserSignUp(un2)
		ad.UserSignIn(un2, ra2, ws)

		a := action.NewJoinWaitingRoom(un)
		ad.Dispatch(a)
		a = action.NewJoinWaitingRoom(un2)
		ad.Dispatch(a)

		u := &server.User{
			ID:         s.Users.List()[0].ID,
			Username:   un,
			Conn:       ws,
			RemoteAddr: ra,
		}
		u2 := &server.User{
			ID:         s.Users.List()[1].ID,
			Username:   un2,
			Conn:       ws,
			RemoteAddr: ra2,
		}
		us := usersInitialState()
		us.Users[un] = u
		us.Users[un2] = u2

		rs := roomsInitialState()
		cwrid := s.Rooms.FindCurrentWaitingRoom().Name
		rs.CurrentWaitingRoom = cwrid
		rs.Rooms[cwrid] = &server.Room{
			Name: cwrid,
			Players: map[string]server.PlayerConn{
				u.ID: server.PlayerConn{
					Conn:       ws,
					RemoteAddr: ra,
				},
				u2.ID: server.PlayerConn{
					Conn:       ws,
					RemoteAddr: ra2,
				},
			},
			Connections: map[string]string{
				ra:  u.ID,
				ra2: u2.ID,
			},

			Size:      6,
			Countdown: 10,
		}

		equalStore(t, s, us, rs)
	})
}

func TestWaitRoomCountdownTick(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ad, s := initStore()
		un := "user name"
		ra := "remote-address"
		ws := &websocket.Conn{}

		ad.UserSignUp(un)
		ad.UserSignIn(un, ra, ws)

		a := action.NewJoinWaitingRoom(un)
		ad.Dispatch(a)
		a = action.NewWaitRoomCountdownTick()
		ad.Dispatch(a)

		u := &server.User{
			ID:         s.Users.List()[0].ID,
			Username:   un,
			Conn:       ws,
			RemoteAddr: ra,
		}
		us := usersInitialState()
		us.Users[un] = u

		rs := roomsInitialState()
		cwrid := s.Rooms.FindCurrentWaitingRoom().Name
		rs.CurrentWaitingRoom = cwrid
		rs.Rooms[cwrid] = &server.Room{
			Name: cwrid,
			Players: map[string]server.PlayerConn{
				u.ID: server.PlayerConn{
					Conn:       ws,
					RemoteAddr: ra,
				},
			},
			Connections: map[string]string{
				ra: u.ID,
			},

			Size:      6,
			Countdown: 9,
		}

		equalStore(t, s, us, rs)

	})
	t.Run("ReduceSize", func(t *testing.T) {
		ad, s := initStore()
		un := "user name"
		ra := "remote-address"
		ws := &websocket.Conn{}

		ad.UserSignUp(un)
		ad.UserSignIn(un, ra, ws)

		a := action.NewJoinWaitingRoom(un)
		ad.Dispatch(a)

		for i := 0; i < 11; i++ {
			a = action.NewWaitRoomCountdownTick()
			ad.Dispatch(a)
		}

		u := &server.User{
			ID:         s.Users.List()[0].ID,
			Username:   un,
			Conn:       ws,
			RemoteAddr: ra,
		}
		us := usersInitialState()
		us.Users[un] = u

		rs := roomsInitialState()
		cwrid := s.Rooms.FindCurrentWaitingRoom().Name
		rs.CurrentWaitingRoom = cwrid
		rs.Rooms[cwrid] = &server.Room{
			Name: cwrid,
			Players: map[string]server.PlayerConn{
				u.ID: server.PlayerConn{
					Conn:       ws,
					RemoteAddr: ra,
				},
			},
			Connections: map[string]string{
				ra: u.ID,
			},

			Size:      5,
			Countdown: 10,
		}

		equalStore(t, s, us, rs)
	})
}

func TestExitWaitingRoom(t *testing.T) {
	t.Run("Success2Players", func(t *testing.T) {
		ad, s := initStore()
		un := "user name"
		ra := "remote-address"
		un2 := "user name2"
		ra2 := "remote-address2"
		ws := &websocket.Conn{}

		ad.UserSignUp(un)
		ad.UserSignIn(un, ra, ws)
		ad.UserSignUp(un2)
		ad.UserSignIn(un2, ra2, ws)

		a := action.NewJoinWaitingRoom(un)
		ad.Dispatch(a)
		a = action.NewJoinWaitingRoom(un2)
		ad.Dispatch(a)
		a = action.NewExitWaitingRoom(un)
		ad.Dispatch(a)

		u := &server.User{
			ID:         s.Users.List()[0].ID,
			Username:   un,
			Conn:       ws,
			RemoteAddr: ra,
		}
		u2 := &server.User{
			ID:         s.Users.List()[1].ID,
			Username:   un2,
			Conn:       ws,
			RemoteAddr: ra2,
		}
		us := usersInitialState()
		us.Users[un] = u
		us.Users[un2] = u2

		rs := roomsInitialState()
		cwrid := s.Rooms.FindCurrentWaitingRoom().Name
		rs.CurrentWaitingRoom = cwrid
		rs.Rooms[cwrid] = &server.Room{
			Name: cwrid,
			Players: map[string]server.PlayerConn{
				u2.ID: server.PlayerConn{
					Conn:       ws,
					RemoteAddr: ra2,
				},
			},
			Connections: map[string]string{
				ra2: u2.ID,
			},

			Size:      6,
			Countdown: 10,
		}

		spew.Dump(s.Users.GetState(), us)
		equalStore(t, s, us, rs)
	})
}

func equalStore(t *testing.T, sto *server.Store, states ...interface{}) {
	t.Helper()

	ris := roomsInitialState()
	uis := usersInitialState()
	for _, st := range states {
		switch s := st.(type) {
		case server.RoomsState:
			ris = s
		case server.UsersState:
			uis = s
		default:
			t.Fatalf("State with type %T is unknown", st)
		}
	}

	assert.Equal(t, ris, sto.Rooms.GetState().(server.RoomsState))
	assert.Equal(t, uis, sto.Users.GetState().(server.UsersState))
}
