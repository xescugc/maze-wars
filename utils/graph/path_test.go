package graph_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xescugc/maze-wars/unit/environment"
	"github.com/xescugc/maze-wars/utils"
	"github.com/xescugc/maze-wars/utils/graph"
)

const (
	atScale    = true
	isAttacker = true
	useCache   = true
)

func TestGraph_Path(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("DefaultTerrestrial", func(t *testing.T) {
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
			enodes := make([]*graph.Node, 0, 0)
			for _, s := range esteps {
				enodes = append(enodes, g.GetNode(int(s.X), int(s.Y)))
			}
			steps, twid := g.Path(0, 0, 60, utils.Down, 0, 2, environment.Terrestrial, !isAttacker, !atScale, useCache)
			require.NotNil(t, steps)
			require.NotEmpty(t, steps)
			require.Len(t, steps, len(esteps))
			assert.Equal(t, esteps, steps)
			assert.Empty(t, twid)

			for i, n := range enodes {
				if i == len(enodes)-1 {
					assert.Nil(t, n.NextStep)
					continue
				}
				assert.Equal(t, enodes[i+1], n.NextStep.Node)
				assert.Equal(t, esteps[i+1].Facing, n.NextStep.Facing)
			}

			// There is no way for me to know if the NextStep logic is used
			// so I'm gonna force another Path so it uses it
			steps, twid = g.Path(0, 0, 60, utils.Down, 0, 2, environment.Terrestrial, !isAttacker, !atScale, useCache)
			require.NotNil(t, steps)
			require.NotEmpty(t, steps)
			require.Len(t, steps, len(esteps))
			assert.Equal(t, esteps, steps)
			assert.Empty(t, twid)
		})
		t.Run("DefaultTerrestrialAttacker", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)
			g.AddTower("id", 2, 1, 1, 1)

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
					Facing: utils.Right,
				},
			}
			enodes := make([]*graph.Node, 0, 0)
			for _, s := range esteps {
				enodes = append(enodes, g.GetNode(int(s.X), int(s.Y)))
			}
			steps, twid := g.Path(0, 0, 60, utils.Down, 0, 2, environment.Terrestrial, isAttacker, !atScale, useCache)
			require.NotNil(t, steps)
			require.NotEmpty(t, steps)
			require.Len(t, steps, len(esteps))
			assert.Equal(t, esteps, steps)
			assert.Equal(t, "id", twid)
		})
		t.Run("WithSlowerMS", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)
			g.AddTower("id", 0, 1, 1, 1)

			esteps := []graph.Step{
				{
					X: 0, Y: 0,
					Facing: utils.Down,
				},
				{
					X: 0.5, Y: 0,
					Facing: utils.Right,
				},
				{
					X: 1, Y: 0,
					Facing: utils.Right,
				},
				{
					X: 1, Y: 0.5,
					Facing: utils.Down,
				},
				{
					X: 1, Y: 1,
					Facing: utils.Down,
				},
				{
					X: 1, Y: 1.5,
					Facing: utils.Down,
				},
				{
					X: 1, Y: 2,
					Facing: utils.Down,
				},
				{
					X: 0.5, Y: 2,
					Facing: utils.Left,
				},
				{
					X: 0, Y: 2,
					Facing: utils.Left,
				},
			}
			steps, twid := g.Path(0, 0, 30, utils.Down, 0, 2, environment.Terrestrial, !isAttacker, atScale, useCache)
			require.NotNil(t, steps)
			require.NotEmpty(t, steps)
			require.Len(t, steps, len(esteps))
			assert.Equal(t, esteps, steps)
			assert.Empty(t, twid)

			// There is no way for me to know if the NextStep logic is used
			// so I'm gonna force another Path so it uses it
			steps, twid = g.Path(0, 0, 30, utils.Down, 0, 2, environment.Terrestrial, !isAttacker, atScale, useCache)
			require.NotNil(t, steps)
			require.NotEmpty(t, steps)
			require.Len(t, steps, len(esteps))
			assert.Equal(t, esteps, steps)
			assert.Empty(t, twid)
		})
		t.Run("DefaultAerial", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)
			g.AddTower("id", 0, 1, 1, 1)

			esteps := []graph.Step{
				{
					X: 0, Y: 0,
					Facing: utils.Down,
				},
				{
					X: 0, Y: 1,
					Facing: utils.Down,
				},
				{
					X: 0, Y: 2,
					Facing: utils.Down,
				},
			}
			enodes := make([]*graph.Node, 0, 0)
			for _, s := range esteps {
				enodes = append(enodes, g.GetNode(int(s.X), int(s.Y)))
			}
			steps, twid := g.Path(0, 0, 60, utils.Down, 0, 2, environment.Aerial, !isAttacker, !atScale, useCache)
			require.NotNil(t, steps)
			require.NotEmpty(t, steps)
			require.Len(t, steps, len(esteps))
			assert.Equal(t, esteps, steps)
			assert.Empty(t, twid)

			for _, n := range enodes {
				assert.Nil(t, n.NextStep)
			}
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
			steps, twid := g.Path(10, 10, 60, utils.Down, 10, 14, environment.Terrestrial, !isAttacker, atScale, useCache)
			require.NotNil(t, steps)
			require.NotEmpty(t, steps)
			require.Len(t, steps, len(esteps))
			assert.Equal(t, esteps, steps)
			assert.Empty(t, twid)
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
			steps, twid := g.Path(11, 10, 60, utils.Down, 10, 14, environment.Terrestrial, !isAttacker, atScale, useCache)
			require.NotNil(t, steps)
			require.NotEmpty(t, steps)
			require.Len(t, steps, len(esteps))
			assert.Equal(t, esteps, steps)
			assert.Empty(t, twid)
		})
		t.Run("RemoveNextStepWhenAddTower", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			_, _ = g.Path(0, 0, 60, utils.Down, 0, 2, environment.Terrestrial, !isAttacker, !atScale, useCache)
			g.AddTower("id", 0, 1, 1, 1)
			for _, yn := range g.Nodes {
				for _, n := range yn {
					assert.Nil(t, n.NextStep)
				}
			}
		})
		t.Run("RemoveNextStepWhenRemoveTower", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			g.AddTower("id", 0, 1, 1, 1)
			_, _ = g.Path(0, 0, 60, utils.Down, 0, 2, environment.Terrestrial, !isAttacker, !atScale, useCache)
			g.RemoveTower("id")
			for _, yn := range g.Nodes {
				for _, n := range yn {
					assert.Nil(t, n.NextStep)
				}
			}
		})
		t.Run("NotRemoveNextStepWhenRemoveTower_NotFound", func(t *testing.T) {
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
			enodes := make([]*graph.Node, 0, 0)
			for _, s := range esteps {
				enodes = append(enodes, g.GetNode(int(s.X), int(s.Y)))
			}

			steps, twid := g.Path(0, 0, 60, utils.Down, 0, 2, environment.Terrestrial, !isAttacker, !atScale, useCache)
			g.RemoveTower("not-found")

			require.NotNil(t, steps)
			require.NotEmpty(t, steps)
			require.Len(t, steps, len(esteps))
			assert.Equal(t, esteps, steps)
			assert.Empty(t, twid)

			for i, n := range enodes {
				if i == len(enodes)-1 {
					assert.Nil(t, n.NextStep)
					continue
				}
				assert.Equal(t, enodes[i+1], n.NextStep.Node)
				assert.Equal(t, esteps[i+1].Facing, n.NextStep.Facing)
			}
		})
	})
}
