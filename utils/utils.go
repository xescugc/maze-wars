package utils

const MaxCapacity = 200

type Object struct {
	X, Y float64
	W, H int
}

// IsColliding will check if Object c is colliding with Object o
func (o Object) IsColliding(c Object) bool {
	selfLeft := o.X
	selfRight := o.X + float64(o.W)
	selfTop := o.Y
	selfBottom := o.Y + float64(o.H)

	enemyLeft := c.X
	enemyRight := c.X + float64(c.W)
	enemyTop := c.Y
	enemyBottom := c.Y + float64(c.H)

	return selfRight > enemyLeft && selfLeft < enemyRight && selfBottom > enemyTop && selfTop < enemyBottom
}

// IsInside will check if Object c is inside (all of it) of Object o
func (o Object) IsInside(c Object) bool {
	selfLeft := o.X
	selfRight := o.X + float64(o.W)
	selfTop := o.Y
	selfBottom := o.Y + float64(o.H)

	enemyLeft := c.X
	enemyRight := c.X + float64(c.W)
	enemyTop := c.Y
	enemyBottom := c.Y + float64(c.H)

	return selfRight >= enemyRight && selfLeft <= enemyLeft && selfBottom >= enemyBottom && selfTop <= enemyTop
}

func (o Object) IsCollidingCircle(c Object, r float64) bool {
	var cx, cy float64

	if c.X < o.X {
		cx = o.X
	} else if c.X > o.X+float64(o.W) {
		cx = o.X + float64(o.W)
	} else {
		cx = c.X
	}

	if c.Y < o.Y {
		cy = o.Y
	} else if c.Y > o.Y+float64(o.H) {
		cy = o.Y + float64(o.H)
	} else {
		cy = c.Y
	}

	dx := c.X - cx
	dy := c.Y - cy

	return dx*dx+dy*dy < r*r
}

type MovingObject struct {
	Object

	Facing      Direction
	MovingCount int
}
