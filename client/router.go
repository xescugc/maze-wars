package client

import (
	"log/slog"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/client/game"
	cutils "github.com/xescugc/maze-wars/client/utils"
	"github.com/xescugc/maze-wars/utils"
)

type RouterStore struct {
	*flux.ReduceStore

	game        *Game
	lobby       *LobbyStore
	signUp      *SignUpStore
	waitingRoom *WaitingRoomStore
	logger      *slog.Logger
}

type RouterState struct {
	Route string
}

func NewRouterStore(d *flux.Dispatcher, su *SignUpStore, ls *LobbyStore, wr *WaitingRoomStore, g *Game, l *slog.Logger) *RouterStore {
	rs := &RouterStore{
		game:        g,
		lobby:       ls,
		signUp:      su,
		waitingRoom: wr,
		logger:      l,
	}

	rs.ReduceStore = flux.NewReduceStore(d, rs.Reduce, RouterState{
		Route: cutils.SignUpRoute,
		//Route: GameRoute,
	})

	return rs
}

func (rs *RouterStore) Update() error {
	b := time.Now()
	defer utils.LogTime(rs.logger, b, "router update")

	rstate := rs.GetState().(RouterState)
	switch rstate.Route {
	case cutils.SignUpRoute:
		rs.signUp.Update()
	case cutils.LobbyRoute:
		rs.lobby.Update()
	case cutils.WaitingRoomRoute:
		rs.waitingRoom.Update()
	case cutils.GameRoute:
		rs.game.Update()
	}
	return nil
}

func (rs *RouterStore) Draw(screen *ebiten.Image) {
	b := time.Now()
	defer utils.LogTime(rs.logger, b, "router draw")

	rstate := rs.GetState().(RouterState)
	switch rstate.Route {
	case cutils.SignUpRoute:
		rs.signUp.Draw(screen)
	case cutils.LobbyRoute:
		rs.lobby.Draw(screen)
	case cutils.WaitingRoomRoute:
		rs.waitingRoom.Draw(screen)
	case cutils.GameRoute:
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
		rstate.Route = cutils.GameRoute
	}

	return rstate
}
