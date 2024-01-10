package client

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/store"
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

	Map *Map
}

func (g *Game) Update() error {
	g.Map.Update()
	g.Camera.Update()
	g.HUD.Update()
	g.Units.Update()
	g.Towers.Update()

	if len(g.Store.Players.List()) == 0 {
		actionDispatcher.Dispatch(action.NewAddPlayer("1", "test1", 0))
		actionDispatcher.Dispatch(action.NewAddPlayer("2", "test2", 1))
		actionDispatcher.Dispatch(action.NewAddPlayer("3", "test3", 2))
		actionDispatcher.Dispatch(action.NewAddPlayer("4", "test4", 3))
		actionDispatcher.Dispatch(action.NewAddPlayer("5", "test5", 4))
		actionDispatcher.Dispatch(action.NewAddPlayer("6", "test6", 5))
	}
	actionDispatcher.TPS()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Map.Draw(screen)
	g.Camera.Draw(screen)
	g.HUD.Draw(screen)
	g.Units.Draw(screen)
	g.Towers.Draw(screen)
}
