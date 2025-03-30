package graph

import (
	"fmt"

	"github.com/xescugc/maze-wars/utils"
)

type Step struct {
	// Node is used to calculate the Path of the Graph but should
	// not be used to get the position of the step (X,Y) as it can
	// be nil when the Path has the flag 'AtScale'
	Node *Node

	X, Y float64

	// When the Steps are for a Unit that has 'Attack' this
	// keeps the 'TowerID' of where this steps are directed
	TowerID string

	Facing utils.Direction
}

func (s Step) String() string {
	return fmt.Sprintf("X:%.2f, Y:%.2f, F:%s", s.X, s.Y, s.Facing.String())
}

//func HashSteps(ss []Step) string {
//var buffer bytes.Buffer
//for _, s := range ss {
//buffer.WriteString(s.String())
//}
//return fmt.Sprintf("%x", (sha256.Sum256([]byte(buffer.String()))))
//}
