package store_test

import (
	"bytes"
	"image"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/assets"
	"github.com/xescugc/maze-wars/store"
)

func TestNewMap(t *testing.T) {
	d := flux.NewDispatcher[*action.Action]()
	st := store.NewStore(d, newEmptyLogger(), isSerer)
	ms := store.NewMap(d, st)
	mstate := ms.GetState()
	m2, _, err := image.Decode(bytes.NewReader(assets.Map_2_png))
	if err != nil {
		log.Fatal(err)
	}
	emstate := store.MapState{
		Players: 2,
		Image:   m2,
	}
	assert.Equal(t, emstate, mstate)
}

//func Test_GetNextLineID(t *testing.T) {
//s := initStore()
//_ = addPlayer(s)
//_ = addPlayer(s)
//_ = addPlayer(s)

//s.Dispatch(action.NewStartGame())

//sms := s.Map.GetState()
//assert.Equal(t, 3, sms.Players)
//assert.Equal(t, 1, s.Map.GetNextLineID(0))
//assert.Equal(t, 2, s.Map.GetNextLineID(1))
//assert.Equal(t, 0, s.Map.GetNextLineID(2))
//}
