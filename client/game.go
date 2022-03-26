package main

import "github.com/hajimehoshi/ebiten/v2"

type Game struct {
	Camera *CameraStore
	Screen *ScreenStore
	Map    *Map
}

func (g *Game) Update() error {
	g.Camera.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Camera.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	ss := g.Screen.GetState().(ScreenState)
	return ss.W, ss.H
}
