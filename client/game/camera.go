package game

import (
	"log/slog"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/utils"
)

// CameraStore is in charge of what it's seen
// on the screen, it also tracks the position
// of the cursor and the wheel scroll
type CameraStore struct {
	*flux.ReduceStore

	Store *store.Store

	logger *slog.Logger

	cameraSpeed int
}

// CameraState is the store data on the Camera
type CameraState struct {
	utils.Object
	Zoom               float64
	LastCursorPosition utils.Object
}

const (
	zoomScale = 0.5
	minZoom   = 0
	maxZoom   = 2
	leeway    = 25
)

// NewCameraStore creates a new CameraState linked to the Dispatcher d
// with the Game g and with width w and height h which is the size of
// the viewport
func NewCameraStore(d *flux.Dispatcher, s *store.Store, l *slog.Logger, w, h int) *CameraStore {
	cs := &CameraStore{
		Store:       s,
		logger:      l,
		cameraSpeed: 10,
	}

	cs.ReduceStore = flux.NewReduceStore(d, cs.Reduce, CameraState{
		Object: utils.Object{
			X: 0, Y: 0,
			W: w,
			H: h,
		},
		Zoom: 1,
	})
	return cs
}

func (cs *CameraStore) Update() error {
	b := time.Now()
	defer utils.LogTime(cs.logger, b, "camera update")

	// TODO: https://github.com/xescugc/maze-wars/issues/4
	//s := cs.GetState().(CameraState)
	//if _, wy := ebiten.Wheel(); wy != 0 {
	//fmt.Println(s.Zoom)
	//if s.Zoom+(wy*zoomScale) <= maxZoom && s.Zoom+(wy*zoomScale) > minZoom {
	//actionDispatcher.CameraZoom(int(wy))
	//}
	//}

	return nil
}

func (cs *CameraStore) Draw(screen *ebiten.Image) {
	b := time.Now()
	defer utils.LogTime(cs.logger, b, "camera draw")
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
	case action.CursorMove:
		// We update the last seen cursor position to not resend unnecessary events
		cstate.LastCursorPosition.X = float64(act.CursorMove.X)
		cstate.LastCursorPosition.Y = float64(act.CursorMove.Y)
	case action.TPS:
		// If the X or Y exceed the current Height or Width then
		// it means the cursor is moving out of boundaries so we
		// increase the camera X/Y at a ratio of the cameraSpeed
		// so we move it around on the map
		if int(cstate.LastCursorPosition.Y) >= (cstate.H - leeway) {
			cstate.Y += float64(cs.cameraSpeed)
		} else if cstate.LastCursorPosition.Y <= (0 + leeway) {
			cstate.Y -= float64(cs.cameraSpeed)
		}

		if int(cstate.LastCursorPosition.X) >= (cstate.W - leeway) {
			cstate.X += float64(cs.cameraSpeed)
		} else if int(cstate.LastCursorPosition.X) <= (0 + leeway) {
			cstate.X -= float64(cs.cameraSpeed)
		}

		// If any of the X or Y values exceeds the boundaries
		// of the actual map we set it to the maximum possible
		// values as we cannot go out of the map
		if cstate.X <= 0 {
			cstate.X = 0
		} else if int(cstate.X) >= cs.Store.Map.GetX() {
			cstate.X = float64(cs.Store.Map.GetX())
		}
		if cstate.Y <= 0 {
			cstate.Y = 0
		} else if int(cstate.Y) >= cs.Store.Map.GetY() {
			cstate.Y = float64(cs.Store.Map.GetY())
		}
	//case action.CameraZoom:
	//cstate.Zoom += act.CameraZoom.Direction * zoomScale
	case action.WindowResizing:
		cstate.W = act.WindowResizing.Width
		cstate.H = act.WindowResizing.Height
	case action.GoHome:
		cp := cs.Store.Lines.FindCurrentPlayer()
		x, y := cs.Store.Map.GetHomeCoordinates(cp.LineID)
		x -= (cstate.W / 2) - ((18 * 16) / 2)
		cstate.X, cstate.Y = float64(x), float64(y)
	}

	return cstate
}
