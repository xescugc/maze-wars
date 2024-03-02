package graph_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xescugc/maze-wars/utils"
	"github.com/xescugc/maze-wars/utils/graph"
)

const atScale = true

func TestGraph_AStar(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Default", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)
			g.AddTower("id", 0, 1, 1, 1)

			esteps := []graph.Step{
				{
					X: 0, Y: 0,
					Facing: utils.Down,
				},
				{
					X: 1, Y: 0,
					Facing: utils.Right,
				},
				{
					X: 1, Y: 1,
					Facing: utils.Down,
				},
				{
					X: 1, Y: 2,
					Facing: utils.Down,
				},
				{
					X: 0, Y: 2,
					Facing: utils.Left,
				},
			}
			steps := g.AStar(0, 0, utils.Down, 0, 2, !atScale)
			require.NotNil(t, steps)
			require.NotEmpty(t, steps)
			require.Len(t, steps, len(esteps))
			assert.Equal(t, esteps, steps)
		})
		t.Run("WithScale", func(t *testing.T) {
			g, err := graph.New(10, 10, 3, 3, 2, 1, 1, 1)
			require.NoError(t, err)
			g.AddTower("id", 10, 12, 2, 2)

			esteps := []graph.Step{
				{
					X: 10, Y: 10,
					Facing: utils.Down,
				},
				{
					X: 11, Y: 10,
					Facing: utils.Right,
				},

				{
					X: 12, Y: 10,
					Facing: utils.Right,
				},
				{
					X: 12, Y: 11,
					Facing: utils.Down,
				},

				{
					X: 12, Y: 12,
					Facing: utils.Down,
				},
				{
					X: 12, Y: 13,
					Facing: utils.Down,
				},

				{
					X: 12, Y: 14,
					Facing: utils.Down,
				},
				{
					X: 11, Y: 14,
					Facing: utils.Left,
				},

				{
					X: 10, Y: 14,
					Facing: utils.Left,
				},
			}
			steps := g.AStar(10, 10, utils.Down, 10, 14, atScale)
			require.NotNil(t, steps)
			require.NotEmpty(t, steps)
			require.Len(t, steps, len(esteps))
			assert.Equal(t, esteps, steps)
		})
		t.Run("WithScaleAndSourceNotNode", func(t *testing.T) {
			g, err := graph.New(10, 10, 3, 3, 2, 1, 1, 1)
			require.NoError(t, err)
			g.AddTower("id", 10, 12, 2, 2)

			esteps := []graph.Step{
				{
					X: 11, Y: 10,
					Facing: utils.Right,
				},

				{
					X: 12, Y: 10,
					Facing: utils.Right,
				},
				{
					X: 12, Y: 11,
					Facing: utils.Down,
				},

				{
					X: 12, Y: 12,
					Facing: utils.Down,
				},
				{
					X: 12, Y: 13,
					Facing: utils.Down,
				},

				{
					X: 12, Y: 14,
					Facing: utils.Down,
				},
				{
					X: 11, Y: 14,
					Facing: utils.Left,
				},

				{
					X: 10, Y: 14,
					Facing: utils.Left,
				},
			}
			steps := g.AStar(11, 10, utils.Down, 10, 14, atScale)
			require.NotNil(t, steps)
			require.NotEmpty(t, steps)
			require.Len(t, steps, len(esteps))
			assert.Equal(t, esteps, steps)
		})
	})
}
