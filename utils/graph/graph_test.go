package graph_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xescugc/maze-wars/utils/graph"
)

func TestNew(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Default", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			assert.NotNil(t, g)
			assert.Len(t, g.Nodes, 3)
			assert.Len(t, g.Nodes[0], 3)
			assert.Len(t, g.Nodes[1], 3)
			assert.Len(t, g.Nodes[2], 3)

			// We load all the graph Nodes into variables
			x0y0 := g.GetNode(0, 0)
			x1y0 := g.GetNode(1, 0)
			x2y0 := g.GetNode(2, 0)
			x0y1 := g.GetNode(0, 1)
			x1y1 := g.GetNode(1, 1)
			x2y1 := g.GetNode(2, 1)
			x0y2 := g.GetNode(0, 2)
			x1y2 := g.GetNode(1, 2)
			x2y2 := g.GetNode(2, 2)

			nodes := []*graph.Node{x0y0, x1y0, x2y0, x0y1, x1y1, x2y1, x0y2, x1y2, x2y2}

			topNil := map[*graph.Node]struct{}{
				x0y0: {}, x1y0: {}, x2y0: {},
			}
			bottomNil := map[*graph.Node]struct{}{
				x0y2: {}, x1y2: {}, x2y2: {},
			}
			rightNil := map[*graph.Node]struct{}{
				x2y0: {}, x2y1: {}, x2y2: {},
			}
			leftNil := map[*graph.Node]struct{}{
				x0y0: {}, x0y1: {}, x0y2: {},
			}

			neighbors2 := []*graph.Node{
				x0y0, x0y2, x2y0, x2y2,
			}
			neighbors3 := []*graph.Node{
				x1y0, x0y1, x1y2, x2y1,
			}
			neighbors4 := []*graph.Node{
				x1y1,
			}

			spawnZone := map[*graph.Node]struct{}{
				x0y0: {}, x1y0: {}, x2y0: {},
			}
			buildingZone := map[*graph.Node]struct{}{
				x0y1: {}, x1y1: {}, x2y1: {},
			}
			deathZone := map[*graph.Node]struct{}{
				x0y2: {}, x1y2: {}, x2y2: {},
			}

			t.Run("TopNeighborNil", func(t *testing.T) {
				for _, n := range nodes {
					require.NotNil(t, n)
					if _, ok := topNil[n]; ok {
						assert.Nil(t, n.TopNeighbor)
					} else {
						assert.NotNil(t, n.TopNeighbor)
					}
				}
			})
			t.Run("BottomNeighborNil", func(t *testing.T) {
				for _, n := range nodes {
					require.NotNil(t, n)
					if _, ok := bottomNil[n]; ok {
						assert.Nil(t, n.BottomNeighbor)
					} else {
						assert.NotNil(t, n.BottomNeighbor)
					}
				}
			})
			t.Run("LeftNeighborNil", func(t *testing.T) {
				for _, n := range nodes {
					require.NotNil(t, n)
					if _, ok := leftNil[n]; ok {
						assert.Nil(t, n.LeftNeighbor)
					} else {
						assert.NotNil(t, n.LeftNeighbor)
					}
				}
			})
			t.Run("RightNeighborNil", func(t *testing.T) {
				for _, n := range nodes {
					require.NotNil(t, n)
					if _, ok := rightNil[n]; ok {
						assert.Nil(t, n.RightNeighbor)
					} else {
						assert.NotNil(t, n.RightNeighbor)
					}
				}
			})
			t.Run("AllNeighbors", func(t *testing.T) {
				assert.NotNil(t, x1y1.TopNeighbor)
				assert.NotNil(t, x1y1.BottomNeighbor)
				assert.NotNil(t, x1y1.LeftNeighbor)
				assert.NotNil(t, x1y1.RightNeighbor)
			})
			t.Run("2Neighbors", func(t *testing.T) {
				for _, n := range neighbors2 {
					assert.Len(t, n.Neighbors, 2)
				}
			})
			t.Run("3Neighbors", func(t *testing.T) {
				for _, n := range neighbors3 {
					assert.Len(t, n.Neighbors, 3)
				}
			})
			t.Run("4Neighbors", func(t *testing.T) {
				for _, n := range neighbors4 {
					assert.Len(t, n.Neighbors, 4)
				}
			})
			t.Run("SpawnZone", func(t *testing.T) {
				for _, n := range nodes {
					require.NotNil(t, n)
					if _, ok := spawnZone[n]; ok {
						assert.True(t, n.IsSpawnZone)
					} else {
						assert.False(t, n.IsSpawnZone)
					}
				}
			})
			t.Run("BuildingZone", func(t *testing.T) {
				for _, n := range nodes {
					require.NotNil(t, n)
					if _, ok := buildingZone[n]; ok {
						assert.True(t, n.IsBuildingZone)
					} else {
						assert.False(t, n.IsBuildingZone)
					}
				}
			})
			t.Run("DeathZone", func(t *testing.T) {
				for _, n := range nodes {
					require.NotNil(t, n)
					if _, ok := deathZone[n]; ok {
						assert.True(t, n.IsDeathZone)
					} else {
						assert.False(t, n.IsDeathZone)
					}
				}
			})
			t.Run("DeathNode", func(t *testing.T) {
				assert.Equal(t, x1y2, g.DeathNode)
			})
		})
		t.Run("WithInvalidScale", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, -1, 1, 1, 1)
			require.NoError(t, err)
			assert.Equal(t, 1, g.Scale)
		})
	})
	t.Run("Error", func(t *testing.T) {
		t.Run("ZonesDoNotMatchHeight", func(t *testing.T) {
			_, err := graph.New(0, 0, 3, 3, 1, 0, 0, 0)
			assert.True(t, errors.Is(err, graph.ErrInvalidZoneHeights))

			_, err = graph.New(0, 0, 3, 3, 1, 10, 10, 10)
			assert.True(t, errors.Is(err, graph.ErrInvalidZoneHeights))
		})
	})
}

func TestGraph_GetNode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Basic", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			assert.NotNil(t, g.GetNode(1, 2))
		})
		t.Run("WithOffset", func(t *testing.T) {
			g, err := graph.New(10, 10, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			assert.NotNil(t, g.GetNode(11, 12))
		})
		t.Run("WithScale", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 10, 1, 1, 1)
			require.NoError(t, err)

			assert.NotNil(t, g.GetNode(10, 20))
		})
		t.Run("WithScaleAndOffset", func(t *testing.T) {
			g, err := graph.New(20, 20, 3, 3, 10, 1, 1, 1)
			require.NoError(t, err)

			assert.NotNil(t, g.GetNode(30, 40))
		})
	})
	t.Run("X", func(t *testing.T) {
		t.Run("X", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			assert.Nil(t, g.GetNode(10, 2))
		})
		t.Run("Y", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			assert.Nil(t, g.GetNode(2, 20))
		})
	})
}

func TestGraph_GetNodeOf(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("WithNoScale", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			assert.Equal(t, g.GetNode(1, 2), g.GetNodeOf(1, 2))
		})
		t.Run("WithScale", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 10, 1, 1, 1)
			require.NoError(t, err)

			assert.Equal(t, g.GetNode(10, 20), g.GetNodeOf(19, 20))
		})
		t.Run("WithScaleAndInside", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 10, 1, 1, 1)
			require.NoError(t, err)

			assert.Equal(t, g.GetNode(10, 20), g.GetNodeOf(11, 21))
		})
	})
	t.Run("Error", func(t *testing.T) {
		t.Run("WithNoScale", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			assert.Nil(t, g.GetNodeOf(10, 22))
		})
		t.Run("WithScale", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 10, 1, 1, 1)
			require.NoError(t, err)

			assert.Nil(t, g.GetNodeOf(41, 41))
		})
	})
}

func TestGraph_GetRandomSpawnNode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Default", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			snds := map[*graph.Node]struct{}{
				g.GetNode(0, 0): {}, g.GetNode(1, 0): {}, g.GetNode(2, 0): {},
			}

			for i := 0; i < 20; i++ {
				n := g.GetRandomSpawnNode()
				require.NotNil(t, n)

				_, ok := snds[n]
				assert.True(t, ok, fmt.Sprintf("X: %d, Y: %d", n.X, n.Y))
			}
		})
		t.Run("WithScaleAndOffset", func(t *testing.T) {
			g, err := graph.New(10, 10, 3, 3, 16, 1, 1, 1)
			require.NoError(t, err)

			snds := map[*graph.Node]struct{}{
				g.GetNode(10, 10): {}, g.GetNode(26, 10): {}, g.GetNode(42, 10): {},
			}

			for i := 0; i < 20; i++ {
				n := g.GetRandomSpawnNode()
				require.NotNil(t, n)

				_, ok := snds[n]
				assert.True(t, ok, fmt.Sprintf("X: %d, Y: %d", n.X, n.Y))
			}
		})
	})
}

func TestGraph_AddTower(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Basic", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			err = g.AddTower("id", 1, 1, 1, 1)
			require.NoError(t, err)

			x0y0 := g.GetNode(0, 0)
			x1y0 := g.GetNode(1, 0)
			x2y0 := g.GetNode(2, 0)
			x0y1 := g.GetNode(0, 1)
			x1y1 := g.GetNode(1, 1)
			x2y1 := g.GetNode(2, 1)
			x0y2 := g.GetNode(0, 2)
			x1y2 := g.GetNode(1, 2)
			x2y2 := g.GetNode(2, 2)

			nodes := []*graph.Node{x0y0, x1y0, x2y0, x0y1, x1y1, x2y1, x0y2, x1y2, x2y2}

			hasTower := map[*graph.Node]struct{}{
				x1y1: {},
			}

			for _, n := range nodes {
				require.NotNil(t, n)
				if _, ok := hasTower[n]; ok {
					assert.True(t, n.HasTower())
				} else {
					assert.False(t, n.HasTower())
				}
			}
		})
		t.Run("With2x2", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 4, 1, 1, 2, 1)
			require.NoError(t, err)

			err = g.AddTower("id", 1, 1, 2, 2)
			require.NoError(t, err)

			x0y0 := g.GetNode(0, 0)
			x1y0 := g.GetNode(1, 0)
			x2y0 := g.GetNode(2, 0)
			x0y1 := g.GetNode(0, 1)
			x1y1 := g.GetNode(1, 1)
			x2y1 := g.GetNode(2, 1)
			x0y2 := g.GetNode(0, 2)
			x1y2 := g.GetNode(1, 2)
			x2y2 := g.GetNode(2, 2)
			x0y3 := g.GetNode(0, 3)
			x1y3 := g.GetNode(1, 3)
			x2y3 := g.GetNode(2, 3)

			nodes := []*graph.Node{x0y0, x1y0, x2y0, x0y1, x1y1, x2y1, x0y2, x1y2, x2y2, x0y3, x1y3, x2y3}

			hasTower := map[*graph.Node]struct{}{
				x1y1: {}, x2y1: {}, x1y2: {}, x2y2: {},
			}

			for _, n := range nodes {
				require.NotNil(t, n)
				if _, ok := hasTower[n]; ok {
					assert.True(t, n.HasTower())
				} else {
					assert.False(t, n.HasTower())
				}
			}
		})
		t.Run("WithScaleAndOffset2x2", func(t *testing.T) {
			g, err := graph.New(10, 10, 3, 4, 16, 1, 2, 1)
			require.NoError(t, err)

			err = g.AddTower("id", 26, 26, 32, 32)
			require.NoError(t, err)

			x0y0 := g.GetNode(10, 10)
			x1y0 := g.GetNode(26, 10)
			x2y0 := g.GetNode(42, 10)
			x0y1 := g.GetNode(10, 26)
			x1y1 := g.GetNode(26, 26)
			x2y1 := g.GetNode(42, 26)
			x0y2 := g.GetNode(10, 42)
			x1y2 := g.GetNode(26, 42)
			x2y2 := g.GetNode(42, 42)
			x0y3 := g.GetNode(10, 58)
			x1y3 := g.GetNode(26, 58)
			x2y3 := g.GetNode(42, 58)

			nodes := []*graph.Node{x0y0, x1y0, x2y0, x0y1, x1y1, x2y1, x0y2, x1y2, x2y2, x0y3, x1y3, x2y3}

			hasTower := map[*graph.Node]struct{}{
				x1y1: {}, x2y1: {}, x1y2: {}, x2y2: {},
			}

			for _, n := range nodes {
				require.NotNil(t, n)
				if _, ok := hasTower[n]; ok {
					assert.True(t, n.HasTower())
				} else {
					assert.False(t, n.HasTower())
				}
			}
		})
	})
	t.Run("Error", func(t *testing.T) {
		t.Run("ErrInvalidBoundaries", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			err = g.AddTower("id", 2, 2, 2, 2)
			assert.True(t, errors.Is(err, graph.ErrInvalidBoundaries))
		})
		t.Run("ErrInvalidPosition", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			err = g.AddTower("id", 1, 1, 1, 1)
			require.NoError(t, err)

			err = g.AddTower("id", 1, 1, 1, 1)
			assert.True(t, errors.Is(err, graph.ErrInvalidPosition))

			err = g.AddTower("id", 0, 0, 1, 1)
			assert.True(t, errors.Is(err, graph.ErrInvalidPosition), "Cannot place on Spawn Zone")

			err = g.AddTower("id", 2, 2, 1, 1)
			assert.True(t, errors.Is(err, graph.ErrInvalidPosition), "Cannot place on Death Zone")
		})
		t.Run("ErrInvalidBlockingPath", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			err = g.AddTower("id", 0, 1, 1, 1)
			require.NoError(t, err)
			err = g.AddTower("id", 1, 1, 1, 1)
			require.NoError(t, err)
			err = g.AddTower("id", 2, 1, 1, 1)
			assert.True(t, errors.Is(err, graph.ErrInvalidBlockingPath))
		})
	})
}

func TestGraph_CanAddTower(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Basic", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			assert.True(t, g.CanAddTower(1, 1, 1, 1))
		})
		t.Run("With2x2", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 4, 1, 1, 2, 1)
			require.NoError(t, err)

			assert.True(t, g.CanAddTower(1, 1, 2, 2))
		})
		t.Run("WithScaleAndOffset2x2", func(t *testing.T) {
			g, err := graph.New(10, 10, 3, 4, 16, 1, 2, 1)
			require.NoError(t, err)

			assert.True(t, g.CanAddTower(26, 26, 32, 32))
		})
	})
	t.Run("Error", func(t *testing.T) {
		t.Run("InvalidBoundaries", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			assert.False(t, g.CanAddTower(2, 2, 2, 2))
		})
		t.Run("InvalidPosition", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 4, 1, 1, 2, 1)
			require.NoError(t, err)

			assert.True(t, g.CanAddTower(1, 1, 2, 2))
			err = g.AddTower("id", 1, 1, 2, 2)
			require.NoError(t, err)

			assert.False(t, g.CanAddTower(1, 1, 2, 2))
		})
		t.Run("ErrInvalidBlockingPath", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 3, 1, 1, 1, 1)
			require.NoError(t, err)

			err = g.AddTower("id", 0, 1, 1, 1)
			require.NoError(t, err)
			err = g.AddTower("id", 1, 1, 1, 1)
			require.NoError(t, err)
			assert.False(t, g.CanAddTower(2, 1, 1, 1))
		})
	})
}

func TestGraph_RemoveTower(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Basic", func(t *testing.T) {
			g, err := graph.New(0, 0, 3, 4, 1, 1, 2, 1)
			require.NoError(t, err)

			g.AddTower("id", 1, 1, 2, 2)

			ok := g.RemoveTower("id")
			assert.True(t, ok)

			ok = g.RemoveTower("id")
			assert.False(t, ok)

			x0y0 := g.GetNode(0, 0)
			x1y0 := g.GetNode(1, 0)
			x2y0 := g.GetNode(2, 0)
			x0y1 := g.GetNode(0, 1)
			x1y1 := g.GetNode(1, 1)
			x2y1 := g.GetNode(2, 1)
			x0y2 := g.GetNode(0, 2)
			x1y2 := g.GetNode(1, 2)
			x2y2 := g.GetNode(2, 2)
			x0y3 := g.GetNode(0, 3)
			x1y3 := g.GetNode(1, 3)
			x2y3 := g.GetNode(2, 3)

			nodes := []*graph.Node{x0y0, x1y0, x2y0, x0y1, x1y1, x2y1, x0y2, x1y2, x2y2, x0y3, x1y3, x2y3}

			for _, n := range nodes {
				require.NotNil(t, n)
				assert.False(t, n.HasTower())
			}
		})
	})
}
