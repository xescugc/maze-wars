package server_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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
			assert.Equal(t, []*server.Room{}, s.Rooms.ListRooms())
		})
	})
}
