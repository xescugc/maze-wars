package client

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/ltw/store"
)

// Game is the main struct that is the initializer
// of the main loop.
// It holds all the other Stores and the Map
type Game struct {
	Store *store.Store

	Camera *CameraStore
	HUD    *HUDStore

	Units  *Units
	Towers *Towers

	SessionID string
}

func (g *Game) Update() error {
	g.Camera.Update()
	g.HUD.Update()
	g.Units.Update()
	g.Towers.Update()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	// Draw will draw just a partial image of the map based on the viewport, so it does not render everything but just the
	// part that it's seen by the user
	// If we want to render everything and just move the viewport around we need o render the full image and change the
	// opt.GeoM.Transport to the Map.X/Y and change the Update function to do the opposite in terms of -+
	//
	// TODO: Maybe create a self Map entity with Update/Draw
	op := &ebiten.DrawImageOptions{}
	s := g.Camera.GetState().(CameraState)
	op.GeoM.Scale(s.Zoom, s.Zoom)
	inverseZoom := maxZoom - s.Zoom + zoomScale
	screen.DrawImage(g.Store.Map.GetState().(store.MapState).Image.(*ebiten.Image).SubImage(image.Rect(int(s.X), int(s.Y), int((s.X+s.W)*inverseZoom), int((s.Y+s.H)*inverseZoom))).(*ebiten.Image), op)

	g.Camera.Draw(screen)
	g.HUD.Draw(screen)
	g.Units.Draw(screen)
	g.Towers.Draw(screen)
}
