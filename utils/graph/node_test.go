package graph_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xescugc/maze-wars/utils/graph"
)

func TestNewNode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		n := graph.NewNode(10, 20)
		en := &graph.Node{
			X: 10, Y: 20,
			ID: "1020",
		}

		assert.Equal(t, en, n)
	})
}

func TestNode_MDistance(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		n1 := graph.NewNode(0, 0)
		n2 := graph.NewNode(0, 1)
		n3 := graph.NewNode(1, 1)

		assert.Equal(t, 1, n1.MDistance(n2))
		assert.Equal(t, 2, n1.MDistance(n3))

		assert.Equal(t, 1, n2.MDistance(n1))
		assert.Equal(t, 1, n2.MDistance(n3))

		assert.Equal(t, 2, n3.MDistance(n1))
		assert.Equal(t, 1, n3.MDistance(n2))
	})
}

func TestNode_HasTower(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		n := graph.NewNode(0, 0)

		assert.False(t, n.HasTower())

		n.TowerID = "id"

		assert.True(t, n.HasTower())
	})
}
