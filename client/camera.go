package client

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
	"github.com/xescugc/ltw/store"
	"github.com/xescugc/ltw/utils"
)

// CameraStore is in charge of what it's seen
// on the screen, it also tracks the position
// of the cursor and the wheel scroll
type CameraStore struct {
	*flux.ReduceStore

	Store *store.Store

	cameraSpeed float64
}

// CameraState is the store data on the Camera
type CameraState struct {
	utils.Object
	Zoom float64
}

const (
	zoomScale = 0.5
	minZoom   = 0
	maxZoom   = 2
	leeway    = 50
)

// NewCameraStore creates a new CameraState linked to the Dispatcher d
// with the Game g and with width w and height h which is the size of
// the viewport
func NewCameraStore(d *flux.Dispatcher, s *store.Store, w, h int) *CameraStore {
	cs := &CameraStore{
		Store:       s,
		cameraSpeed: 10,
	}

	cs.ReduceStore = flux.NewReduceStore(d, cs.Reduce, CameraState{
		Object: utils.Object{
			X: 0, Y: 0,
			W: float64(w),
			H: float64(h),
		},
		Zoom: 1,
	})
	return cs
}

func (cs *CameraStore) Update() error {
	// TODO: https://github.com/xescugc/ltw/issues/4
	//s := cs.GetState().(CameraState)
	//if _, wy := ebiten.Wheel(); wy != 0 {
	//fmt.Println(s.Zoom)
	//if s.Zoom+(wy*zoomScale) <= maxZoom && s.Zoom+(wy*zoomScale) > minZoom {
	//actionDispatcher.CameraZoom(int(wy))
	//}
	//}

	return nil
}

func (cs *CameraStore) Draw(screen *ebiten.Image) {}

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
	case action.CursorMove:
		// If the X or Y exceed the current Height or Width then
		// it means the cursor is moving out of boundaries so we
		// increase the camera X/Y at a ratio of the cameraSpeed
		// so we move it around on the map
		if float64(act.CursorMove.Y) >= (cstate.H - leeway) {
			cstate.Y += cs.cameraSpeed
		} else if act.CursorMove.Y <= (0 + leeway) {
			cstate.Y -= cs.cameraSpeed
		}

		if float64(act.CursorMove.X) >= (cstate.W - leeway) {
			cstate.X += cs.cameraSpeed
		} else if act.CursorMove.X <= (0 + leeway) {
			cstate.X -= cs.cameraSpeed
		}

		// If any of the X or Y values exceeds the boundaries
		// of the actual map we set it to the maximum possible
		// values as we cannot go out of the map
		if cstate.X <= 0 {
			cstate.X = 0
		} else if cstate.X >= float64(cs.Store.Map.GetX()) {
			cstate.X = float64(cs.Store.Map.GetX())
		}
		if cstate.Y <= 0 {
			cstate.Y = 0
		} else if cstate.Y >= float64(cs.Store.Map.GetY()) {
			cstate.Y = float64(cs.Store.Map.GetY())
		}
	case action.CameraZoom:
		cstate.Zoom += float64(act.CameraZoom.Direction) * zoomScale
	case action.WindowResizing:
		cstate.W = float64(act.WindowResizing.Width)
		cstate.H = float64(act.WindowResizing.Height)
	case action.GoHome:
		cp := cs.Store.Players.FindCurrent()
		cstate.X, cstate.Y = cs.Store.Map.GetHomeCoordinates(cp.LineID)
	}

	return cstate
}
