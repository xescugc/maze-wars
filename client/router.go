package client

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
)

const (
	LobbyRoute = "lobby"
	GameRoute  = "game"
)

type RouterStore struct {
	*flux.ReduceStore

	game  *Game
	lobby *LobbyStore
}

type RouterState struct {
	Route string
}

func NewRouterStore(d *flux.Dispatcher, g *Game, l *LobbyStore) *RouterStore {
	rs := &RouterStore{
		game:  g,
		lobby: l,
	}

	rs.ReduceStore = flux.NewReduceStore(d, rs.Reduce, RouterState{
		Route: LobbyRoute,
		//Route: GameRoute,
	})

	return rs
}

func (rs *RouterStore) Update() error {
	rstate := rs.GetState().(RouterState)
	switch rstate.Route {
	case LobbyRoute:
		rs.lobby.Update()
	case GameRoute:
		rs.game.Update()
	}
	return nil
}

func (rs *RouterStore) Draw(screen *ebiten.Image) {
	rstate := rs.GetState().(RouterState)
	switch rstate.Route {
	case LobbyRoute:
		rs.lobby.Draw(screen)
	case GameRoute:
		rs.game.Draw(screen)
	}
}

func (rs *RouterStore) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	cs := rs.game.Camera.GetState().(CameraState)
	if cs.W != float64(outsideWidth) || cs.H != float64(outsideHeight) {
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
	}

	return rstate
}
