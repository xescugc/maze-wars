package graph

import (
	"fmt"
	"math"
)

// Node is a node in the graph that stores the int value at that node
// along with a map to the vertices it is connected to via edges.
type Node struct {
	// TODO: Maybe a way to improve the AttackUnit calculation
	// HasUnit bool

	// ID concatenation of X and Y
	ID string

	// X, Y are the positions of the Node
	X, Y int

	// TowerID is the ID of the Tower placed
	// on this Node
	TowerID string

	HasPath bool

	IsSpawnZone    bool
	IsBuildingZone bool
	IsDeathZone    bool

	TopNeighbor    *Node
	BottomNeighbor *Node
	RightNeighbor  *Node
	LeftNeighbor   *Node

	Neighbors     []*Node
	NeighborSteps []Step

	NextStep *Step
}

// GenerateID will generate the ID concatenating X and Y
func NewNode(x, y int) *Node {
	n := &Node{
		X: x, Y: y,
		ID: fmt.Sprintf("%d%d", x, y),
	}
	return n
}

// MDistance calculates the Manhattan distance between n and t
func (n *Node) MDistance(t *Node) int {
	return abs(n.X-t.X) + abs(n.Y-t.Y)
}

// PDistance calculates the Pythagoras distance between n and t
func (n *Node) PDistance(t *Node) int {
	return int(math.Sqrt(
		math.Pow(float64(n.X-t.X), 2) + math.Pow(float64(n.Y-t.Y), 2),
	))
}

// HasTower checks if the Node has a TowerID
func (n *Node) HasTower() bool { return n.TowerID != "" }

// abs converts the x into an absolute positive number
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
