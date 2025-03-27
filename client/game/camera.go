package game

import (
	"log/slog"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/utils"
)

// CameraStore is in charge of what it's seen
// on the screen, it also tracks the position
// of the cursor and the wheel scroll
type CameraStore struct {
	*flux.ReduceStore[CameraState, *action.Action]

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
func NewCameraStore(d *flux.Dispatcher[*action.Action], s *store.Store, l *slog.Logger, w, h int) *CameraStore {
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

func (cs *CameraStore) Reduce(state CameraState, act *action.Action) CameraState {
	switch act.Type {
	case action.CursorMove:
		// We update the last seen cursor position to not resend unnecessary events
		state.LastCursorPosition.X = float64(act.CursorMove.X)
		state.LastCursorPosition.Y = float64(act.CursorMove.Y)
	case action.TPS:
		// If the X or Y exceed the current Height or Width then
		// it means the cursor is moving out of boundaries so we
		// increase the camera X/Y at a ratio of the cameraSpeed
		// so we move it around on the map
		if int(state.LastCursorPosition.Y) >= (state.H - leeway) {
			state.Y += float64(cs.cameraSpeed)
		} else if state.LastCursorPosition.Y <= (0 + leeway) {
			state.Y -= float64(cs.cameraSpeed)
		}

		if int(state.LastCursorPosition.X) >= (state.W - leeway) {
			state.X += float64(cs.cameraSpeed)
		} else if int(state.LastCursorPosition.X) <= (0 + leeway) {
			state.X -= float64(cs.cameraSpeed)
		}

		// If any of the X or Y values exceeds the boundaries
		// of the actual map we set it to the maximum possible
		// values as we cannot go out of the map
		if state.X <= 0 {
			state.X = 0
		} else if int(state.X) >= cs.Store.Map.GetX() {
			state.X = float64(cs.Store.Map.GetX())
		}
		if state.Y <= 0 {
			state.Y = 0
		} else if int(state.Y) >= cs.Store.Map.GetY() {
			state.Y = float64(cs.Store.Map.GetY())
		}
	//case action.CameraZoom:
	//state.Zoom += act.CameraZoom.Direction * zoomScale
	case action.WindowResizing:
		state.W = act.WindowResizing.Width
		state.H = act.WindowResizing.Height
	case action.GoHome:
		cp := cs.Store.Game.FindCurrentPlayer()
		x, y := cs.Store.Map.GetHomeCoordinates(cp.LineID)
		x -= (state.W / 2) - ((18 * 16) / 2)
		state.X, state.Y = float64(x), float64(y)
	}

	return state
}
