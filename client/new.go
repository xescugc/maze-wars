package client

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/ltw/action"
	"github.com/xescugc/ltw/assets"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	// actionDispatcher is the main dispatcher of the application
	// all the actions have to be registered to it
	actionDispatcher *ActionDispatcher

	wsc *websocket.Conn

	normalFont font.Face
	smallFont  font.Face

	// TODO: Remove this global when we can specify
	// the room from the client
	room string
)

func init() {
	rand.Seed(time.Now().UnixNano())

	// Initialize Font
	tt, err := opentype.Parse(assets.NormalFont_ttf)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	normalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingVertical,
	})
	smallFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    16,
		DPI:     dpi,
		Hinting: font.HintingVertical,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func New(ad *ActionDispatcher, rs *RouterStore, opt Options) error {
	ebiten.SetWindowTitle("LTW")
	ebiten.SetWindowSize(opt.ScreenW*2, opt.ScreenH*2)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	actionDispatcher = ad
	room = opt.Room

	// Establish connection
	u := url.URL{Scheme: "ws", Host: opt.HostURL, Path: "/ws"}

	var err error
	wsc, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to dial the server %q: %w", u, err)
	}
	defer wsc.Close()

	go wsHandler()

	err = wsc.WriteJSON(action.NewJoinRoom(opt.Room, opt.Name))
	if err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	err = ebiten.RunGame(rs)
	if err != nil {
		return fmt.Errorf("failed to RunGame: %w", err)
	}

	return nil
}

func wsHandler() {
	for {
		var act *action.Action
		err := wsc.ReadJSON(&act)
		if err != nil {
			// TODO remove from the Room
			log.Fatal(err)
		}

		actionDispatcher.Dispatch(act)
	}
}

func wsSend(a *action.Action) {
	a.Room = room
	err := wsc.WriteJSON(a)
	if err != nil {
		log.Fatal(err)
	}
}
