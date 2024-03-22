package client

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/maze-wars/action"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var (
	// actionDispatcher is the main dispatcher of the application
	// all the actions have to be registered to it
	actionDispatcher *ActionDispatcher

	wsc *websocket.Conn
)

func init() {
	rand.Seed(time.Now().UnixNano())

}

func New(ctx context.Context, ad *ActionDispatcher, rs *RouterStore, opt Options) error {
	ebiten.SetWindowTitle("Maze Wars")
	ebiten.SetWindowSize(opt.ScreenW*2, opt.ScreenH*2)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	actionDispatcher = ad

	actionDispatcher.CheckVersion()

	err := ebiten.RunGame(rs)
	if err != nil {
		return fmt.Errorf("failed to RunGame: %w", err)
	}

	if wsc != nil {
		wsc.CloseNow()
	}

	return nil
}

func wsHandler(ctx context.Context) {
	for {
		var act *action.Action
		err := wsjson.Read(ctx, wsc, &act)
		if err != nil {
			log.Fatal(err)
		}

		actionDispatcher.Dispatch(act)
		// If the action is StartGame then we
		// focus the user on it's own line
		if act.Type == action.StartGame {
			actionDispatcher.GoHome()
		}
	}
}

func wsSend(a *action.Action) {
	err := wsjson.Write(context.Background(), wsc, a)
	if err != nil {
		log.Fatal(err)
	}
}
