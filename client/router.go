package client

import (
	"log/slog"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/utils"
)

const (
	SignUpRoute      = "sign_up"
	LobbyRoute       = "lobby"
	GameRoute        = "game"
	WaitingRoomRoute = "waiting_room"
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
		Route: SignUpRoute,
		//Route: GameRoute,
	})

	return rs
}

func (rs *RouterStore) Update() error {
	b := time.Now()
	defer utils.LogTime(rs.logger, b, "router update")

	rstate := rs.GetState().(RouterState)
	switch rstate.Route {
	case SignUpRoute:
		rs.signUp.Update()
	case LobbyRoute:
		rs.lobby.Update()
	case WaitingRoomRoute:
		rs.waitingRoom.Update()
	case GameRoute:
		rs.game.Update()
	}
	return nil
}

func (rs *RouterStore) Draw(screen *ebiten.Image) {
	b := time.Now()
	defer utils.LogTime(rs.logger, b, "router draw")

	rstate := rs.GetState().(RouterState)
	switch rstate.Route {
	case SignUpRoute:
		rs.signUp.Draw(screen)
	case LobbyRoute:
		rs.lobby.Draw(screen)
	case WaitingRoomRoute:
		rs.waitingRoom.Draw(screen)
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
	case action.StartGame:
		rstate.Route = GameRoute
	}

	return rstate
}
