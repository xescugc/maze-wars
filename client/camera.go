package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
)

type CameraStore struct {
	*flux.ReduceStore

	game *Game

	cameraSpeed float64
}

type CameraState struct {
	Object
}

func NewCameraStore(d *flux.Dispatcher, g *Game, w, h int) *CameraStore {
	cs := &CameraStore{
		game:        g,
		cameraSpeed: 10,
	}

	cs.ReduceStore = flux.NewReduceStore(d, cs.Reduce, CameraState{
		Object: Object{
			X: 0, Y: 0,
			W: float64(w),
			H: float64(h),
		},
	})
	return cs
}

func (cs *CameraStore) Update() error {
	x, y := ebiten.CursorPosition()
	actionDispatcher.CameraMove(x, y)

	return nil
}

// Draw will draw just a partial image of the map based on the viewport, so it does not render everything but just the
// part that it's seen by the user
// If we want to render everything and just move the viewport around we need o render the full image and change the
// opt.GeoM.Transport to the Map.X/Y and change the Update function to do the opposite in terms of -+
func (cs *CameraStore) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	s := cs.GetState().(CameraState)
	screen.DrawImage(cs.game.Map.Image.(*ebiten.Image).SubImage(image.Rect(int(s.X), int(s.Y), int(s.X+s.W), int(s.Y+s.H))).(*ebiten.Image), op)
}

func (cs *CameraStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	cstate, ok := state.(CameraState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.CameraMove:
		cs.GetDispatcher().WaitFor(cs.game.Screen.GetDispatcherToken())

		ss := cs.game.Screen.GetState().(ScreenState)

		if act.CameraMove.Y >= ss.H {
			cstate.Y += cs.cameraSpeed
		} else if act.CameraMove.Y <= 0 {
			cstate.Y -= cs.cameraSpeed
		}

		if act.CameraMove.X >= ss.W {
			cstate.X += cs.cameraSpeed
		} else if act.CameraMove.X <= 0 {
			cstate.X -= cs.cameraSpeed
		}

		if cstate.X <= 0 {
			cstate.X = 0
		} else if cstate.X >= float64(cs.game.Map.GetX()) {
			cstate.X = float64(cs.game.Map.GetX())
		}

		if cstate.Y <= 0 {
			cstate.Y = 0
		} else if cstate.Y >= float64(cs.game.Map.GetY()) {
			cstate.Y = float64(cs.game.Map.GetY())
		}
	}

	return cstate
}
