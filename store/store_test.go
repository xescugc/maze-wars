package store_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/store"
)

func TestNewStore(t *testing.T) {
	d := flux.NewDispatcher()
	s := store.NewStore(d)

	assert.NotNil(t, s)
}
