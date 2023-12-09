package utils

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
)

type Object struct {
	X, Y, W, H float64
}

// IsColliding will check if Object c is colliding with Object o
func (o Object) IsColliding(c Object) bool {
	selfLeft := o.X
	selfRight := o.X + o.W
	selfTop := o.Y
	selfBottom := o.Y + o.H

	enemyLeft := c.X
	enemyRight := c.X + c.W
	enemyTop := c.Y
	enemyBottom := c.Y + c.H

	return selfRight > enemyLeft && selfLeft < enemyRight && selfBottom > enemyTop && selfTop < enemyBottom
}

// IsInside will check if Object c is inside (all of it) of Object o
func (o Object) IsInside(c Object) bool {
	selfLeft := o.X
	selfRight := o.X + o.W
	selfTop := o.Y
	selfBottom := o.Y + o.H

	enemyLeft := c.X
	enemyRight := c.X + c.W
	enemyTop := c.Y
	enemyBottom := c.Y + c.H

	return selfRight >= enemyRight && selfLeft <= enemyLeft && selfBottom >= enemyBottom && selfTop <= enemyTop
}

// PDistance will calculate the Pythagorean distance between the 2 objects
// based on X and Y position
func (o Object) PDistance(c Object) float64 {
	return math.Sqrt(
		math.Pow(o.X-c.X, 2) + math.Pow(o.Y-c.Y, 2),
	)
}

// MDistance will calculate the Manhattan distance between the 2 objects
// based on X and Y position
func (o Object) MDistance(c Object) float64 {
	//return 2 * (math.Abs(o.X-c.X) + math.Abs(o.Y-c.Y))
	return (math.Abs(o.X-c.X) + math.Abs(o.Y-c.Y))
}

type Step struct {
	Object

	Facing Direction
}

func (s Step) String() string {
	return fmt.Sprintf("X:%f, Y:%f, W:%f, H:%f, F:%s", s.X, s.Y, s.W, s.H, s.Facing.String())
}

// NeighborSteps returns all the possible steps around the o
func (o Object) NeighborSteps() []Step {
	return []Step{
		Step{
			Object: Object{
				Y: o.Y - 1,
				X: o.X,
				W: 1, H: 1,
			},
			Facing: Up,
		},

		Step{
			Object: Object{
				Y: o.Y + 1,
				X: o.X,
				W: 1, H: 1,
			},
			Facing: Down,
		},

		Step{
			Object: Object{
				X: o.X - 1,
				Y: o.Y,
				W: 1, H: 1,
			},
			Facing: Left,
		},

		Step{
			Object: Object{
				X: o.X + 1,
				Y: o.Y,
				W: 1, H: 1,
			},
			Facing: Right,
		},
	}
}

type MovingObject struct {
	Object

	Facing      Direction
	MovingCount int
}

func HashSteps(ss []Step) string {
	var buffer bytes.Buffer
	for _, s := range ss {
		buffer.WriteString(s.String())
	}
	return fmt.Sprintf("%x", (sha256.Sum256([]byte(buffer.String()))))
}
