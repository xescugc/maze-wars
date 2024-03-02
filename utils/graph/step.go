package graph

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/xescugc/maze-wars/utils"
)

type Step struct {
	// Node is used to calculate the AStar of the Graph but should
	// not be used to get the position of the step (X,Y) as it can
	// be nil when the AStar has the flag 'AtScale'
	Node *Node

	X, Y int

	Facing utils.Direction
}

func (s Step) String() string {
	return fmt.Sprintf("X:%d, Y:%d, F:%s", s.X, s.Y, s.Facing.String())
}

func HashSteps(ss []Step) string {
	var buffer bytes.Buffer
	for _, s := range ss {
		buffer.WriteString(s.String())
	}
	return fmt.Sprintf("%x", (sha256.Sum256([]byte(buffer.String()))))
}
