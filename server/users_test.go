package server_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/xescugc/maze-wars/server"
	"github.com/xescugc/maze-wars/server/mock"
	"nhooyr.io/websocket"
)

func TestUsers_FindByUsername(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Found", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mwsc := mock.NewMockWSConnector(ctrl)

			ad, s := initStore(mwsc)
			un := "user name"
			ra := "remote-address"
			ws := &websocket.Conn{}

			ad.UserSignUp(un)

			u := server.User{
				ID:       s.Users.List()[0].ID,
				Username: un,
			}

			dbu, ok := s.Users.FindByUsername(un)
			assert.True(t, ok)
			assert.Equal(t, u, dbu)

			ad.UserSignIn(un, ra, ws)
			u.Conn = ws
			u.RemoteAddr = ra

			dbu, ok = s.Users.FindByUsername(un)
			assert.True(t, ok)
			assert.Equal(t, u, dbu)
		})
		t.Run("NotFound", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mwsc := mock.NewMockWSConnector(ctrl)
			_, s := initStore(mwsc)

			dbu, ok := s.Users.FindByUsername("not-found")
			assert.False(t, ok)
			assert.Equal(t, server.User{}, dbu)
		})
	})
}

func TestUsers_FindByRemoteAddress(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Found", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mwsc := mock.NewMockWSConnector(ctrl)

			ad, s := initStore(mwsc)
			un := "user name"
			ra := "remote-address"
			ws := &websocket.Conn{}

			ad.UserSignUp(un)
			ad.UserSignIn(un, ra, ws)

			u := server.User{
				ID:         s.Users.List()[0].ID,
				Username:   un,
				Conn:       ws,
				RemoteAddr: ra,
			}

			dbu, ok := s.Users.FindByRemoteAddress(ra)
			assert.True(t, ok)
			assert.Equal(t, u, dbu)
		})
		t.Run("NotFound", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mwsc := mock.NewMockWSConnector(ctrl)
			_, s := initStore(mwsc)

			dbu, ok := s.Users.FindByRemoteAddress("not-found")
			assert.False(t, ok)
			assert.Equal(t, server.User{}, dbu)
		})
	})
}

func TestUsers_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Found", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mwsc := mock.NewMockWSConnector(ctrl)

			ad, s := initStore(mwsc)
			un := "user name"
			ra := "remote-address"
			ws := &websocket.Conn{}

			ad.UserSignUp(un)
			ad.UserSignIn(un, ra, ws)

			u := server.User{
				ID:         s.Users.List()[0].ID,
				Username:   un,
				Conn:       ws,
				RemoteAddr: ra,
			}

			assert.Equal(t, []*server.User{&u}, s.Users.List())
		})
		t.Run("Empty", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mwsc := mock.NewMockWSConnector(ctrl)
			_, s := initStore(mwsc)

			assert.Equal(t, []*server.User{}, s.Users.List())
		})
	})
}
