package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Game is the main struct that is the initializer
// of the main loop.
// It holds all the other Stores and the Map
type Game struct {
	Camera  *CameraStore
	HUD     *HUDStore
	Players *PlayersStore
	Units   *UnitsStore
	Towers  *TowersStore

	Map *Map
}

func (g *Game) Update() error {
	g.Camera.Update()
	g.HUD.Update()
	g.Units.Update()
	g.Towers.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Camera.Draw(screen)
	g.HUD.Draw(screen)
	g.Units.Draw(screen)
	g.Towers.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	cs := g.Camera.GetState().(CameraState)
	if cs.W != float64(outsideWidth) || cs.H != float64(outsideHeight) {
		actionDispatcher.WindowResizing(outsideWidth, outsideHeight)
	}
	return outsideWidth, outsideHeight
}
