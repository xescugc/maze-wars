package utils

import (
	"math"
)

type Object struct {
	X, Y, W, H int
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
		math.Pow(float64(o.X-c.X), 2) + math.Pow(float64(o.Y-c.Y), 2),
	)
}

type MovingObject struct {
	Object

	Facing      Direction
	MovingCount int
}
