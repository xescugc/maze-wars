package client

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/maze-wars/store"
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
	return nil
}

func (m *Map) Draw(screen *ebiten.Image) {
	// Draw will draw just a partial image of the map based on the viewport, so it does not render everything but just the
	// part that it's seen by the user
	// If we want to render everything and just move the viewport around we need o render the full image and change the
	// opt.GeoM.Transport to the Map.X/Y and change the Update function to do the opposite in terms of -+
	//
	// TODO: Maybe create a self Map entity with Update/Draw
	op := &ebiten.DrawImageOptions{}
	s := m.game.Camera.GetState().(CameraState)
	op.GeoM.Scale(s.Zoom, s.Zoom)
	inverseZoom := maxZoom - s.Zoom + zoomScale
	mi := ebiten.NewImageFromImage(m.game.Store.Map.GetState().(store.MapState).Image)
	screen.DrawImage(mi.SubImage(image.Rect(int(s.X), int(s.Y), int((s.X+s.W)*inverseZoom), int((s.Y+s.H)*inverseZoom))).(*ebiten.Image), op)
}
