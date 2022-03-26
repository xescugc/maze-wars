package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Enemy struct {
	Object
	Facing      ebiten.Key
	MovingCount int
}

var (
	facingToTile = map[ebiten.Key]int{
		ebiten.KeyS: 0,
		ebiten.KeyW: 1,
		ebiten.KeyA: 2,
		ebiten.KeyD: 3,
	}
)

func NewEnemy() *Enemy {
	var w, h float64 = 16, 16
	return &Enemy{
		Object: Object{
			X: 16,
			Y: 16,
			W: w,
			H: h,
		},
		Facing: ebiten.KeyS,
	}

}

func (e *Enemy) Update() error {
	e.MovingCount += 1
	e.Y += 1
	return nil
}

//func (e *Enemy) Draw(screen *ebiten.Image, c *Camera) {
//if !e.IsColliding(c.Object) {
//return
//}
//op := &ebiten.DrawImageOptions{}
//op.GeoM.Translate(e.X-c.X, e.Y-c.Y)
//sx := facingToTile[e.Facing] * int(e.W)
//i := (e.MovingCount / 5) % 4
//sy := i * int(e.H)
//screen.DrawImage(reptileImg.(*ebiten.Image).SubImage(image.Rect(sx, sy, sx+int(e.W), sy+int(e.H))).(*ebiten.Image), op)
//}
