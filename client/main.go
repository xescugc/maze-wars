package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
)

var (
	// actionDispatcher is the main dispatcher of the application
	// all the actions have to be registered to it
	actionDispatcher *ActionDispatcher

	// verbose is used to check weather or not we have to display
	// more logs
	verbose bool
)

func init() {
	flag.BoolVar(&verbose, "verbose", false, "Will log all the actions")
	rand.Seed(time.Now().UnixNano())
}

func main() {
	flag.Parse()

	screenW := 288
	screenH := 240

	ebiten.SetWindowTitle("LTW")
	ebiten.SetWindowSize(screenW*2, screenH*2)
	dispatcher := flux.NewDispatcher()

	actionDispatcher = NewActionDispatcher(dispatcher)

	if verbose {
		NewLoggerStore(dispatcher)
	}

	m, err := NewMap()
	if err != nil {
		log.Fatal(err)
	}

	g := &Game{
		Screen:  NewScreenStore(dispatcher, screenW, screenH),
		Players: NewPlayersStore(dispatcher),
		Map:     m,
	}
	// TODO: Change this to pass the specific store needed instead of all the game object
	g.Camera = NewCameraStore(dispatcher, g, screenW, screenH)
	g.Units = NewUnitsStore(dispatcher, g)
	g.Towers = NewTowersStore(dispatcher, g)
	g.HUD, err = NewHUDStore(dispatcher, g)
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
