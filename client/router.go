package client

import (
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/client/game"
	"github.com/xescugc/maze-wars/utils"
)

type RouterStore struct {
	*flux.ReduceStore

	game   *Game
	root   *RootStore
	signUp *SignUpStore

	logger *slog.Logger
}

type RouterState struct {
	Route string
}

func NewRouterStore(d *flux.Dispatcher, su *SignUpStore, ros *RootStore, g *Game, l *slog.Logger) *RouterStore {
	rs := &RouterStore{
		game:   g,
		root:   ros,
		signUp: su,

		logger: l,
	}

	rs.ReduceStore = flux.NewReduceStore(d, rs.Reduce, RouterState{
		Route: utils.SignUpRoute,
	})

	return rs
}

func (rs *RouterStore) Update() error {
	b := time.Now()
	defer utils.LogTime(rs.logger, b, "router update")

	// See https://golang.org/ref/spec#Handling_panics.
	// This will recover from runtime panics and then panic again after
	// reporting to Sentry.
	defer func() {
		err := recover()

		if err != nil {
			hub := sentry.CurrentHub().Clone()
			hub.Recover(err)
			hub.Flush(time.Second * 5)
			if Environment == "dev" {
				panic(err)
			}
		}
	}()

	rstate := rs.GetState().(RouterState)
	switch rstate.Route {
	case utils.SignUpRoute:
		rs.signUp.Update()
	case utils.RootRoute, utils.LobbiesRoute, utils.LearnRoute, utils.HomeRoute, utils.NewLobbyRoute, utils.ShowLobbyRoute:
		rs.root.Update()
	case utils.GameRoute:
		rs.game.Update()
	}
	return nil
}

func (rs *RouterStore) Draw(screen *ebiten.Image) {
	b := time.Now()
	defer utils.LogTime(rs.logger, b, "router draw")

	// See https://golang.org/ref/spec#Handling_panics.
	// This will recover from runtime panics and then panic again after
	// reporting to Sentry.
	defer func() {
		err := recover()

		if err != nil {
			hub := sentry.CurrentHub().Clone()
			hub.Recover(err)
			hub.Flush(time.Second * 5)
			if Environment == "dev" {
				panic(err)
			}
		}
	}()

	rstate := rs.GetState().(RouterState)
	switch rstate.Route {
	case utils.SignUpRoute:
		rs.signUp.Draw(screen)
	case utils.RootRoute, utils.LobbiesRoute, utils.LearnRoute, utils.HomeRoute, utils.NewLobbyRoute, utils.ShowLobbyRoute:
		rs.root.Draw(screen)
	case utils.GameRoute:
		rs.game.Draw(screen)
	}
}

func (rs *RouterStore) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	cs := rs.game.Game.Camera.GetState().(game.CameraState)
	if cs.W != outsideWidth || cs.H != outsideHeight {
		actionDispatcher.WindowResizing(outsideWidth, outsideHeight)
	}
	return outsideWidth, outsideHeight
}

func (rs *RouterStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	rstate, ok := state.(RouterState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.NavigateTo:
		rstate.Route = act.NavigateTo.Route
	case action.StartGame:
		rstate.Route = utils.GameRoute
	}

	return rstate
}
