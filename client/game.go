package client

import (
	"log/slog"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/client/game"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/utils"
)

// Game is the main struct that is the initializer
// of the main loop.
// It holds all the other Stores and the Map
type Game struct {
	Game *game.Game

	Logger *slog.Logger
}

func NewGame(s *store.Store, d *flux.Dispatcher[*action.Action], l *slog.Logger) *Game {
	gl := l.WithGroup("game")
	return &Game{
		Game:   game.New(s, game.NewActionDispatcher(d, s, wsSend, gl), gl),
		Logger: l,
	}
}

func (g *Game) Update() error {
	b := time.Now()
	defer utils.LogTime(g.Logger, b, "game update")

	g.Game.Update()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	b := time.Now()
	defer utils.LogTime(g.Logger, b, "game draw")

	g.Game.Draw(screen)
}
