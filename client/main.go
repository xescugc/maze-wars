package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
)

var (
	actionDispatcher *ActionDispatcher
)

func main() {
	screenW := 288
	screenH := 240

	ebiten.SetWindowTitle("LTW")
	ebiten.SetWindowSize(screenW*2, screenH*2)
	dispatcher := flux.NewDispatcher()

	actionDispatcher = NewActionDispatcher(dispatcher)

	m, err := NewMap()
	if err != nil {
		log.Fatal(err)
	}

	g := &Game{
		Screen: NewScreenStore(dispatcher, screenW, screenH),
		Map:    m,
	}
	g.Camera = NewCameraStore(dispatcher, g, screenW, screenH)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
