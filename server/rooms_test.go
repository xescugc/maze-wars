package server_test

import (
	"testing"

	"github.com/coder/websocket"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/server"
	"github.com/xescugc/maze-wars/server/mock"
)

func TestRoom_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mwsc := mock.NewMockWSConnector(ctrl)

			_, s := initStore(mwsc)
			assert.Equal(t, []*server.Room{}, s.Rooms.List())
		})
		t.Run("WithRooms", func(t *testing.T) {
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

			cwrid := s.Rooms.FindCurrentVs6WaitingRoom().Name
			er := &server.Room{
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
			assert.Equal(t, []*server.Room{er}, s.Rooms.List())
		})
	})
}
func TestRoom_FindCurrentVs6WaitingRoom(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mwsc := mock.NewMockWSConnector(ctrl)

			_, s := initStore(mwsc)
			assert.Nil(t, s.Rooms.FindCurrentVs6WaitingRoom())
		})
		t.Run("WithRoom", func(t *testing.T) {
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

			cwrid := s.Rooms.FindCurrentVs6WaitingRoom().Name
			er := &server.Room{
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
			assert.Equal(t, er, s.Rooms.FindCurrentVs6WaitingRoom())
		})
	})
}

func TestRoom_GetNextID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("WithRoom", func(t *testing.T) {
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

			cwrid := s.Rooms.FindCurrentVs6WaitingRoom().Name
			assert.Equal(t, 1, s.Rooms.GetNextID(cwrid))

			un2 := "user name2"
			ra2 := "remote-address2"

			ad.UserSignUp(un2)
			ad.UserSignIn(un2, ra2, ws)

			a = action.NewJoinWaitingRoom(un2)
			ad.Dispatch(a)

			assert.Equal(t, 2, s.Rooms.GetNextID(cwrid))
		})
	})
}
