package inputer

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// go:generate mockgen -destination=../mock/inputer.go -package mock github.com/xescugc/maze-wars/inputer Inputer
type Inputer interface {
	CursorPosition() (int, int)

	IsMouseButtonJustPressed(button ebiten.MouseButton) bool
	IsKeyJustPressed(key ebiten.Key) bool
}

type Ebiten struct{}

func NewEbiten() *Ebiten {
	return &Ebiten{}
}

func (e *Ebiten) CursorPosition() (int, int) {
	return ebiten.CursorPosition()
}

func (e *Ebiten) IsMouseButtonJustPressed(button ebiten.MouseButton) bool {
	return inpututil.IsMouseButtonJustPressed(button)
}

func (e *Ebiten) IsKeyJustPressed(key ebiten.Key) bool {
	return inpututil.IsKeyJustPressed(key)
}
