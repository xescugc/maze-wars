package client

import (
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/maze-wars/store"
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
	mi := ebiten.NewImageFromImage(m.game.Store.Map.GetState().(store.MapState).Image)
	screen.DrawImage(mi.SubImage(image.Rect(int(s.X), int(s.Y), int((s.X+s.W)*inverseZoom), int((s.Y+s.H)*inverseZoom))).(*ebiten.Image), op)

	cs := m.game.Camera.GetState().(CameraState)
	cp := m.game.Store.Players.FindCurrent()
	x, y := m.game.Store.Map.GetHomeCoordinates(cp.LineID)
	// Color TOP and Bottom
	for i := x - 4; i <= x+(18*16)+3; i++ {
		// We draw 3 lines so it's kind of **bold**
		// and it's easier to see
		screen.Set(int(i-cs.X), int(y-cs.Y-4), green)
		screen.Set(int(i-cs.X), int(y-cs.Y-3), green)
		screen.Set(int(i-cs.X), int(y-cs.Y-2), green)

		screen.Set(int(i-cs.X), int((y+86*16)-cs.Y+3), green)
		screen.Set(int(i-cs.X), int((y+86*16)-cs.Y+2), green)
		screen.Set(int(i-cs.X), int((y+86*16)-cs.Y+1), green)
	}

	// Color Left and Right
	for i := y - 1; i <= y+(86*16); i++ {
		screen.Set(int(x-cs.X-4), int(i-cs.Y), green)
		screen.Set(int(x-cs.X-3), int(i-cs.Y), green)
		screen.Set(int(x-cs.X-2), int(i-cs.Y), green)

		screen.Set(int((x+18*16)-cs.X+3), int(i-cs.Y), green)
		screen.Set(int((x+18*16)-cs.X+2), int(i-cs.Y), green)
		screen.Set(int((x+18*16)-cs.X+1), int(i-cs.Y), green)
	}
}
