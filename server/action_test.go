package server_test

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/server"
	"github.com/xescugc/maze-wars/server/mock"
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mwsc := mock.NewMockWSConnector(ctrl)

	_, s := initStore(mwsc)
	equalStore(t, s)
}

func TestUserSignUp(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mwsc := mock.NewMockWSConnector(ctrl)

		ad, s := initStore(mwsc)
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
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mwsc := mock.NewMockWSConnector(ctrl)

		ad, s := initStore(mwsc)
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
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mwsc := mock.NewMockWSConnector(ctrl)

		ad, s := initStore(mwsc)
		un := "user name"

		ad.UserSignUp(un)
		ad.UserSignOut(un)

		equalStore(t, s)
	})
}

func TestJoinWaitingRoom(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mwsc := mock.NewMockWSConnector(ctrl)

		ad, s := initStore(mwsc)
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
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mwsc := mock.NewMockWSConnector(ctrl)

		ad, s := initStore(mwsc)
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

		dbu, _ := s.Users.FindByUsername(un)
		u := &server.User{
			ID:         dbu.ID,
			Username:   un,
			Conn:       ws,
			RemoteAddr: ra,
		}
		dbu2, _ := s.Users.FindByUsername(un2)
		u2 := &server.User{
			ID:         dbu2.ID,
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
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mwsc := mock.NewMockWSConnector(ctrl)

		ad, s := initStore(mwsc)
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
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mwsc := mock.NewMockWSConnector(ctrl)

		ad, s := initStore(mwsc)
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
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mwsc := mock.NewMockWSConnector(ctrl)

		ad, s := initStore(mwsc)
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
		dbu, _ := s.Users.FindByUsername(un)
		u := &server.User{
			ID:         dbu.ID,
			Username:   un,
			Conn:       ws,
			RemoteAddr: ra,
		}
		dbu2, _ := s.Users.FindByUsername(un2)
		u2 := &server.User{
			ID:         dbu2.ID,
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

		equalStore(t, s, us, rs)
	})
}

func TestStartGame(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("When we reach max players", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mwsc := mock.NewMockWSConnector(ctrl)
			mwsc.EXPECT().Write(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			ad, s := initStore(mwsc)
			// Key is 'un' and value is 'ra'
			users := make(map[string]string)
			for i := 1; i <= 6; i++ {
				users[fmt.Sprintf("user name %d", i)] = fmt.Sprintf("remote-address%d", i)
			}
			ws := &websocket.Conn{}

			var cwr *server.Room
			for un, ra := range users {
				ad.UserSignUp(un)
				ad.UserSignIn(un, ra, ws)
				a := action.NewJoinWaitingRoom(un)
				ad.Dispatch(a)
				if cwr == nil {
					cwr = s.Rooms.FindCurrentWaitingRoom()
				}
			}

			us := usersInitialState()
			for un, ra := range users {
				u, _ := s.Users.FindByUsername(un)
				u = server.User{
					ID:            u.ID,
					Username:      un,
					Conn:          ws,
					RemoteAddr:    ra,
					CurrentRoomID: cwr.Name,
				}
				us.Users[un] = &u
			}

			rs := roomsInitialState()
			require.NotNil(t, cwr.Game)

			rs.Rooms[cwr.Name] = &server.Room{
				Name:        cwr.Name,
				Players:     make(map[string]server.PlayerConn),
				Connections: make(map[string]string),
				Size:        6,
				Countdown:   10,
				// We set it to the value already created, we just
				// check that it's not nil beforehand
				Game: cwr.Game,
			}
			for _, u := range us.Users {
				rs.Rooms[cwr.Name].Players[u.ID] = server.PlayerConn{
					Conn:       ws,
					RemoteAddr: u.RemoteAddr,
				}
				rs.Rooms[cwr.Name].Connections[u.RemoteAddr] = u.ID
			}
			equalStore(t, s, us, rs)
		})
		t.Run("When we reach countdown", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mwsc := mock.NewMockWSConnector(ctrl)
			mwsc.EXPECT().Write(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			ad, s := initStore(mwsc)
			// Key is 'un' and value is 'ra'
			users := make(map[string]string)
			for i := 1; i <= 2; i++ {
				users[fmt.Sprintf("user name %d", i)] = fmt.Sprintf("remote-address%d", i)
			}
			ws := &websocket.Conn{}

			var cwr *server.Room
			for un, ra := range users {
				ad.UserSignUp(un)
				ad.UserSignIn(un, ra, ws)
				a := action.NewJoinWaitingRoom(un)
				ad.Dispatch(a)
				if cwr == nil {
					cwr = s.Rooms.FindCurrentWaitingRoom()
				}
			}

			// Countdown from 6 users to 2 users
			for i := 1; i <= 44; i++ {
				ad.WaitRoomCountdownTick()
			}

			us := usersInitialState()
			for un, ra := range users {
				u, _ := s.Users.FindByUsername(un)
				u = server.User{
					ID:            u.ID,
					Username:      un,
					Conn:          ws,
					RemoteAddr:    ra,
					CurrentRoomID: cwr.Name,
				}
				us.Users[un] = &u
			}

			rs := roomsInitialState()
			require.NotNil(t, cwr.Game)

			rs.CurrentWaitingRoom = ""
			rs.Rooms[cwr.Name] = &server.Room{
				Name:        cwr.Name,
				Players:     make(map[string]server.PlayerConn),
				Connections: make(map[string]string),
				Size:        2,
				Countdown:   10,
				// We set it to the value already created, we just
				// check that it's not nil beforehand
				Game: cwr.Game,
			}
			for _, u := range us.Users {
				rs.Rooms[cwr.Name].Players[u.ID] = server.PlayerConn{
					Conn:       ws,
					RemoteAddr: u.RemoteAddr,
				}
				rs.Rooms[cwr.Name].Connections[u.RemoteAddr] = u.ID
			}
			equalStore(t, s, us, rs)
		})
	})
}

func TestRemovePlayer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Remove 1 player", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mwsc := mock.NewMockWSConnector(ctrl)
			mwsc.EXPECT().Write(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			ad, s := initStore(mwsc)
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

			cwr := s.Rooms.FindCurrentWaitingRoom()

			// Countdown from 6 users to 2 users
			for i := 1; i <= 44; i++ {
				ad.WaitRoomCountdownTick()
			}

			dbu, _ := s.Users.FindByUsername(un)
			u := &server.User{
				ID:            dbu.ID,
				Username:      un,
				Conn:          ws,
				RemoteAddr:    ra,
				CurrentRoomID: cwr.Name,
			}
			dbu2, _ := s.Users.FindByUsername(un2)
			u2 := &server.User{
				ID:         dbu2.ID,
				Username:   un2,
				Conn:       ws,
				RemoteAddr: ra2,
			}

			a = action.NewRemovePlayer(u2.ID)
			a.Room = cwr.Name
			ad.Dispatch(a)

			us := usersInitialState()
			us.Users[un] = u
			us.Users[un2] = u2

			rs := roomsInitialState()
			rs.Rooms[cwr.Name] = &server.Room{
				Name:        cwr.Name,
				Players:     make(map[string]server.PlayerConn),
				Connections: make(map[string]string),
				Size:        2,
				Countdown:   10,
				Game:        cwr.Game,
			}
			rs.Rooms[cwr.Name].Players[u.ID] = server.PlayerConn{
				Conn:       ws,
				RemoteAddr: u.RemoteAddr,
			}
			rs.Rooms[cwr.Name].Connections[u.RemoteAddr] = u.ID

			equalStore(t, s, us, rs)
		})
		t.Run("Remove 2 players", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mwsc := mock.NewMockWSConnector(ctrl)
			mwsc.EXPECT().Write(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			ad, s := initStore(mwsc)
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

			cwr := s.Rooms.FindCurrentWaitingRoom()

			// Countdown from 6 users to 2 users
			for i := 1; i <= 44; i++ {
				ad.WaitRoomCountdownTick()
			}

			dbu, _ := s.Users.FindByUsername(un)
			u := &server.User{
				ID:         dbu.ID,
				Username:   un,
				Conn:       ws,
				RemoteAddr: ra,
			}
			dbu2, _ := s.Users.FindByUsername(un2)
			u2 := &server.User{
				ID:         dbu2.ID,
				Username:   un2,
				Conn:       ws,
				RemoteAddr: ra2,
			}

			a = action.NewRemovePlayer(u.ID)
			a.Room = cwr.Name
			ad.Dispatch(a)
			a = action.NewRemovePlayer(u2.ID)
			a.Room = cwr.Name
			ad.Dispatch(a)

			us := usersInitialState()
			us.Users[un] = u
			us.Users[un2] = u2

			equalStore(t, s, us)
		})
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
