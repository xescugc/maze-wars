package store_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/store"
)

func TestNewStore(t *testing.T) {
	d := flux.NewDispatcher[*action.Action]()
	s := store.NewStore(d, newEmptyLogger())

	assert.NotNil(t, s)
}
