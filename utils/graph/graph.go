package graph

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/xescugc/maze-wars/unit/environment"
	"github.com/xescugc/maze-wars/utils"
)

var (
	ErrInvalidPosition     = errors.New("this position is invalid")
	ErrInvalidBoundaries   = errors.New("this position is exceeds the boundaries")
	ErrInvalidBlockingPath = errors.New("this position is blocking the path")
	ErrInvalidZoneHeights  = errors.New("the heights of the zones does not add up to the expected height")
)

const (
	atScale    = true
	isAttacker = true
)

// Graph represents a set of Nodes in a X, Y coordinates
type Graph struct {
	// Nodes represents all the Nodes in X,Y position
	Nodes map[int]map[int]*Node

	// Is the position of what the 0,0 would be, so any
	// calculations are scaled to that new X,Y
	OffsetX, OffsetY int

	// The width and Height of the Graph when it was initialized
	W, H int

	SpawnZoneH    int
	BuildingZoneH int
	DeathZoneH    int

	// What will be the real conversion 1:Scale
	// so if the scale is 16 it means that 1 Node is
	// 16 pixels
	Scale int

	// DeathNode is the Node that will be the default
	// node in which the units will try to move to
	DeathNode *Node
}

// New crates a Graph with
// * ox, oy as Offsets from where the graph is
// * width w and height h
// * with scale s
// * szh is the Height of the Spawn Zone
// * bzh is the Height of the Building Zone
// * dzh is the Height of the Death Zone
// All the 3 Zones Height have to add to the
// global Height of the graph (h) and the numbers
// are not relative, meaning the 'bzh' height starts
// after the 'szh' height and the 'dzh' after the 'bzh',
// so szh + bzh + dzh = h
func New(ox, oy, w, h, s, szh, bzh, dzh int) (*Graph, error) {
	if szh+bzh+dzh != h {
		return nil, ErrInvalidZoneHeights
	}
	// The Scale has to be at least 1
	// it cannot be lower
	if s <= 0 {
		s = 1
	}

	g := &Graph{
		Nodes: make(map[int]map[int]*Node),

		OffsetX: ox, OffsetY: oy,
		W: w, H: h,

		SpawnZoneH:    szh,
		BuildingZoneH: bzh,
		DeathZoneH:    dzh,

		Scale: s,
	}

	// Initialize all the Nodes
	for wi := 0; wi < w; wi++ {
		nx := ox + (s * wi)
		g.Nodes[nx] = make(map[int]*Node)
		for hi := 0; hi < h; hi++ {
			ny := oy + (s * hi)
			n := NewNode(nx, ny)

			if hi < szh {
				n.IsSpawnZone = true
			} else if hi < szh+bzh {
				n.IsBuildingZone = true
			} else if hi < szh+bzh+dzh {
				n.IsDeathZone = true
				if g.DeathNode == nil && wi == w/2 {
					g.DeathNode = n
				}
			}

			g.Nodes[nx][ny] = n
		}
	}

	// Set the Neighbors of all the Nodes
	for wi := 0; wi < w; wi++ {
		nx := ox + (s * wi)
		for hi := 0; hi < h; hi++ {
			ny := oy + (s * hi)
			n := g.GetNode(nx, ny)

			// Set the specific Neighbors
			n.TopNeighbor = g.GetNode(nx, oy+(s*(hi-1)))
			n.BottomNeighbor = g.GetNode(nx, oy+(s*(hi+1)))
			n.LeftNeighbor = g.GetNode(ox+(s*(wi-1)), ny)
			n.RightNeighbor = g.GetNode(ox+(s*(wi+1)), ny)

			// Set the list of Neighbors
			if n.TopNeighbor != nil {
				n.Neighbors = append(n.Neighbors, n.TopNeighbor)
				n.NeighborSteps = append(n.NeighborSteps, Step{Node: n.TopNeighbor, Facing: utils.Up})
			}
			if n.BottomNeighbor != nil {
				n.Neighbors = append(n.Neighbors, n.BottomNeighbor)
				n.NeighborSteps = append(n.NeighborSteps, Step{Node: n.BottomNeighbor, Facing: utils.Down})
			}
			if n.LeftNeighbor != nil {
				n.Neighbors = append(n.Neighbors, n.LeftNeighbor)
				n.NeighborSteps = append(n.NeighborSteps, Step{Node: n.LeftNeighbor, Facing: utils.Left})
			}
			if n.RightNeighbor != nil {
				n.Neighbors = append(n.Neighbors, n.RightNeighbor)
				n.NeighborSteps = append(n.NeighborSteps, Step{Node: n.RightNeighbor, Facing: utils.Right})
			}
		}
	}

	return g, nil
}

// GetNode returns the node on the coordinates x,y
// if not present then return nil
// The x, y are with Offset and Scale
func (g *Graph) GetNode(x, y int) *Node {
	xn, ok := g.Nodes[x]
	if !ok {
		return nil
	}
	return xn[y]
}

// GetNodeOf will get the Node in which x,y belong to if any
// by trying to find the closes x,y to the g.Scale
func (g *Graph) GetNodeOf(x, y int) *Node {
	x = utils.PreviousMultiple(x, g.Scale)
	y = utils.PreviousMultiple(y, g.Scale)
	return g.GetNode(x, y)
}

// GetRandomSpawnNode returns a random node on the Spawn zone
func (g *Graph) GetRandomSpawnNode() *Node {
	p := rand.Intn(g.W * g.SpawnZoneH)
	x := p % g.W
	y := p % g.SpawnZoneH
	return g.GetNode(g.OffsetX+(x*g.Scale), g.OffsetY+(y*g.Scale))
}

// AddTower adds a tower to the desired X,Y location
// with W and H
func (g *Graph) AddTower(id string, x, y, w, h int) error {
	nodes, err := g.canAddTower(x, y, w, h)
	if err != nil {
		return fmt.Errorf("failed to add tower: %w", err)
	}

	for _, n := range nodes {
		n.TowerID = id
	}

	// When a new tower is added we remove all the
	// cached paths by removing the Node.NextStep
	for _, ny := range g.Nodes {
		for _, n := range ny {
			n.NextStep = nil
		}
	}

	return nil
}

func (g *Graph) canAddTower(x, y, w, h int) ([]*Node, error) {
	w = w / g.Scale
	h = h / g.Scale

	// Validates that the affected nodes exist
	nodes := make([]*Node, 0, w*h)
	for wi := 0; wi < w; wi++ {
		nx := x + (g.Scale * wi)
		for hi := 0; hi < h; hi++ {
			ny := y + (g.Scale * hi)
			n := g.GetNode(nx, ny)
			if n == nil {
				return nil, ErrInvalidBoundaries
			}
			nodes = append(nodes, n)
		}
	}

	// Validates that the nodes do not have a tower already
	for _, n := range nodes {
		if n.HasTower() {
			return nil, ErrInvalidPosition
		} else if n.IsSpawnZone {
			return nil, ErrInvalidPosition
		} else if n.IsDeathZone {
			return nil, ErrInvalidPosition
		}
	}

	// In order for it to work we'll momentarily add the
	// Tower to the graph and remove it afterwards so we
	// can run AStar
	for _, n := range nodes {
		n.TowerID = "test"
	}
	defer func() {
		for _, n := range nodes {
			n.TowerID = ""
		}
	}()

	// Validates that adding the tower will not block the path
	// from top to bottom
	steps, _ := g.AStar(float64(g.OffsetX), float64(g.OffsetY), basicTPS, utils.Down, g.DeathNode.X, g.DeathNode.Y, environment.Terrestrial, !isAttacker, !atScale)
	if len(steps) == 0 {
		return nil, ErrInvalidBlockingPath
	}

	return nodes, nil
}

func (g *Graph) CanAddTower(x, y, w, h int) bool {
	_, err := g.canAddTower(x, y, w, h)
	return err == nil
}

func (g *Graph) RemoveTower(id string) bool {
	var found bool

	for _, xn := range g.Nodes {
		for _, n := range xn {
			if n.TowerID == id {
				found = true
				n.TowerID = ""
			}
		}
	}

	if found {
		for _, xn := range g.Nodes {
			for _, n := range xn {
				n.NextStep = nil
			}
		}
	}

	return found
}
