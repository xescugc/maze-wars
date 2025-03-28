package server_test

import (
	"testing"

	"github.com/coder/websocket"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/xescugc/maze-wars/server"
	"github.com/xescugc/maze-wars/server/mock"
)

var (
	roomsInitialState = func() server.RoomsState {
		return server.RoomsState{
			Searching: make(map[string]*server.Room),
			Waiting:   make(map[string]*server.Room),
			Rooms:     make(map[string]*server.Room),
			Users:     make(map[string]*server.User),
		}
	}
)

func TestEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mwsc := mock.NewMockWSConnector(ctrl)

	_, s := initStore(mwsc)
	assert.Equal(t, roomsInitialState(), s.Rooms.GetState())
}

func TestUserSignUp(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mwsc := mock.NewMockWSConnector(ctrl)

		ad, s := initStore(mwsc)
		un := "user name"

		ad.UserSignUp(un, "ImageKey")

		u := &server.User{
			ID:       s.Rooms.ListUsers()[0].ID,
			Username: un,
			ImageKey: "ImageKey",
		}
		rs := roomsInitialState()
		rs.Users[un] = u

		assert.Equal(t, rs, s.Rooms.GetState())
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

		ad.UserSignUp(un, "ImageKey")
		ad.UserSignIn(un, ra, ws)

		u := &server.User{
			ID:         s.Rooms.ListUsers()[0].ID,
			Username:   un,
			Conn:       ws,
			RemoteAddr: ra,
			ImageKey:   "ImageKey",
		}
		rs := roomsInitialState()
		rs.Users[un] = u

		assert.Equal(t, rs, s.Rooms.GetState())
	})
	t.Run("NotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mwsc := mock.NewMockWSConnector(ctrl)

		ad, s := initStore(mwsc)
		un := "user name"
		ra := "remote-address"
		ws := &websocket.Conn{}

		ad.UserSignUp(un, "ImageKey")
		ad.UserSignIn("not-found", ra, ws)

		rs := roomsInitialState()
		u := &server.User{
			ID:       s.Rooms.ListUsers()[0].ID,
			Username: un,
			ImageKey: "ImageKey",
		}
		rs.Users[un] = u

		assert.Equal(t, rs, s.Rooms.GetState())
	})
}

func TestUserSignOut(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mwsc := mock.NewMockWSConnector(ctrl)

		ad, s := initStore(mwsc)
		un := "user name"

		ad.UserSignUp(un, "ImageKey")
		ad.UserSignOut(un)

		assert.Equal(t, roomsInitialState(), s.Rooms.GetState())
	})
	t.Run("InRoom", func(t *testing.T) {
		// TODO: Missing context to create a room now
		t.Skip("Missing context to create a room now")
	})
	t.Run("NotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mwsc := mock.NewMockWSConnector(ctrl)

		ad, s := initStore(mwsc)
		un := "user name"

		ad.UserSignUp(un, "ImageKey")
		ad.UserSignOut("not-found")

		rs := roomsInitialState()
		u := &server.User{
			ID:       s.Rooms.ListUsers()[0].ID,
			Username: un,
			ImageKey: "ImageKey",
		}
		rs.Users[un] = u

		assert.Equal(t, rs, s.Rooms.GetState())
	})
}

func TestFindGame(t *testing.T) {
	//t.Run("Success", func(t *testing.T) {
	//ctrl := gomock.NewController(t)
	//defer ctrl.Finish()
	//mwsc := mock.NewMockWSConnector(ctrl)

	//ad, s := initStore(mwsc)
	//un := "user name"
	//ra := "remote-address"
	//ws := &websocket.Conn{}

	//ad.UserSignUp(un, "ImageKey")
	//ad.UserSignIn(un, ra, ws)

	//u := &server.User{
	//ID:         s.Rooms.ListUsers()[0].ID,
	//Username:   un,
	//Conn:       ws,
	//RemoteAddr: ra,
	//ImageKey:   "ImageKey",
	//}

	//var (
	//v1     = true
	//rank   = false
	//vsBots = true
	//)
	//ad.Dispatch(action.NewFindGame(un, v1, rank, vsBots))

	//rs := roomsInitialState()
	//rs.Users[un] = u

	//assert.Equal(t, rs, s.Rooms.GetState())
	//})
}

// TODO: Lobbies
