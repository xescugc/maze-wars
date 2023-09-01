package main

import (
	_ "embed"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
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

type Step struct {
	Object

	Facing ebiten.Key
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
			Facing: ebiten.KeyW,
		},

		Step{
			Object: Object{
				Y: o.Y + 1,
				X: o.X,
				W: 1, H: 1,
			},
			Facing: ebiten.KeyS,
		},

		Step{
			Object: Object{
				X: o.X - 1,
				Y: o.Y,
				W: 1, H: 1,
			},
			Facing: ebiten.KeyA,
		},

		Step{
			Object: Object{
				X: o.X + 1,
				Y: o.Y,
				W: 1, H: 1,
			},
			Facing: ebiten.KeyD,
		},
	}
}

type Entity struct {
	Object
	Image image.Image
}

type MovingEntity struct {
	Entity

	Facing      ebiten.Key
	MovingCount int
}
