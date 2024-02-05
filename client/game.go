package client

import (
	"log/slog"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/utils"
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

	Logger *slog.Logger
}

func NewGame(s *store.Store, l *slog.Logger) *Game {
	return &Game{
		Store:  s,
		Logger: l,
	}
}

func (g *Game) Update() error {
	b := time.Now()
	defer utils.LogTime(g.Logger, b, "game update")

	g.Map.Update()
	g.Camera.Update()
	g.Units.Update()
	g.Towers.Update()
	g.HUD.Update()

	actionDispatcher.TPS()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	b := time.Now()
	defer utils.LogTime(g.Logger, b, "game draw")

	g.Map.Draw(screen)
	g.Camera.Draw(screen)
	g.HUD.Draw(screen)
	g.Units.Draw(screen)
	g.Towers.Draw(screen)
}
