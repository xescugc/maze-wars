package game

import (
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	cutils "github.com/xescugc/maze-wars/client/utils"
	"github.com/xescugc/maze-wars/utils"
)

type Map struct {
	game *Game
}

func NewMap(g *Game) *Map {
	m := &Map{
		game: g,
	}

	return m
}

func (m *Map) Update() error {
	b := time.Now()
	defer utils.LogTime(m.game.Logger, b, "map update")

	return nil
}

func (m *Map) Draw(screen *ebiten.Image) {
	b := time.Now()
	defer utils.LogTime(m.game.Logger, b, "map draw")

	// Draw will draw just a partial image of the map based on the viewport, so it does not render everything but just the
	// part that it's seen by the user
	// If we want to render everything and just move the viewport around we need o render the full image and change the
	// opt.GeoM.Transport to the Map.X/Y and change the Update function to do the opposite in terms of -+

	op := &ebiten.DrawImageOptions{}
	s := m.game.Camera.GetState().(CameraState)
	op.GeoM.Scale(s.Zoom, s.Zoom)
	inverseZoom := maxZoom - s.Zoom + zoomScale
	screen.DrawImage(imagesCache.Get(m.game.Store.Map.GetImageKey()).SubImage(image.Rect(s.X, s.Y, int(float64((s.X+s.W))*inverseZoom), int(float64((s.Y+s.H))*inverseZoom))).(*ebiten.Image), op)

	cs := m.game.Camera.GetState().(CameraState)
	cp := m.game.Store.Players.FindCurrent()
	x, y := m.game.Store.Map.GetHomeCoordinates(cp.LineID)
	// Color TOP and Bottom
	for i := x - 4; i <= x+(18*16)+3; i++ {
		// We draw 3 lines so it's kind of **bold**
		// and it's easier to see
		screen.Set(i-cs.X, y-cs.Y-4, cutils.Green)
		screen.Set(i-cs.X, y-cs.Y-3, cutils.Green)
		screen.Set(i-cs.X, y-cs.Y-2, cutils.Green)

		screen.Set(i-cs.X, (y+86*16)-cs.Y+3, cutils.Green)
		screen.Set(i-cs.X, (y+86*16)-cs.Y+2, cutils.Green)
		screen.Set(i-cs.X, (y+86*16)-cs.Y+1, cutils.Green)
	}

	// Color Left and Right
	for i := y - 1; i <= y+(86*16); i++ {
		screen.Set(x-cs.X-4, i-cs.Y, cutils.Green)
		screen.Set(x-cs.X-3, i-cs.Y, cutils.Green)
		screen.Set(x-cs.X-2, i-cs.Y, cutils.Green)

		screen.Set((x+18*16)-cs.X+3, i-cs.Y, cutils.Green)
		screen.Set((x+18*16)-cs.X+2, i-cs.Y, cutils.Green)
		screen.Set((x+18*16)-cs.X+1, i-cs.Y, cutils.Green)
	}
}
