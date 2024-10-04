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

	// Clone the current hub so that modifications of the scope are visible only
	// within this function.
	hub := sentry.CurrentHub().Clone()

	// See https://golang.org/ref/spec#Handling_panics.
	// This will recover from runtime panics and then panic again after
	// reporting to Sentry.
	defer func() {
		if x := recover(); x != nil {
			// Create an event and enqueue it for reporting.
			hub.Recover(x)
			// Because the goroutine running this code is going to crash the
			// program, call Flush to send the event to Sentry before it is too
			// late. Set the timeout to an appropriate value depending on your
			// program. The value is the maximum time to wait before giving up
			// and dropping the event.
			hub.Flush(2 * time.Second)
			// Note that if multiple goroutines panic, possibly only the first
			// one to call Flush will succeed in sending the event. If you want
			// to capture multiple panics and still crash the program
			// afterwards, you need to coordinate error reporting and
			// termination differently.
			panic(x)
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

	// Clone the current hub so that modifications of the scope are visible only
	// within this function.
	hub := sentry.CurrentHub().Clone()

	// See https://golang.org/ref/spec#Handling_panics.
	// This will recover from runtime panics and then panic again after
	// reporting to Sentry.
	defer func() {
		if x := recover(); x != nil {
			// Create an event and enqueue it for reporting.
			hub.Recover(x)
			// Because the goroutine running this code is going to crash the
			// program, call Flush to send the event to Sentry before it is too
			// late. Set the timeout to an appropriate value depending on your
			// program. The value is the maximum time to wait before giving up
			// and dropping the event.
			hub.Flush(2 * time.Second)
			// Note that if multiple goroutines panic, possibly only the first
			// one to call Flush will succeed in sending the event. If you want
			// to capture multiple panics and still crash the program
			// afterwards, you need to coordinate error reporting and
			// termination differently.
			panic(x)
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
