package client

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/assets"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var (
	// actionDispatcher is the main dispatcher of the application
	// all the actions have to be registered to it
	actionDispatcher *ActionDispatcher

	wsc *websocket.Conn

	normalFont font.Face
	smallFont  font.Face
)

func init() {
	rand.Seed(time.Now().UnixNano())

	// Initialize Font
	tt, err := opentype.Parse(assets.Kongtext_ttf)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	normalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	smallFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    16,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

}

func New(ctx context.Context, ad *ActionDispatcher, rs *RouterStore, opt Options) error {
	ebiten.SetWindowTitle("Maze Wars")
	ebiten.SetWindowSize(opt.ScreenW*2, opt.ScreenH*2)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	actionDispatcher = ad

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
			// TODO remove from the Room
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
