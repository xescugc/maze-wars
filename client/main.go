package main

import (
	"flag"
	"log"
	"math/rand"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
	"github.com/xescugc/ltw/assets"
	"github.com/xescugc/ltw/store"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	// actionDispatcher is the main dispatcher of the application
	// all the actions have to be registered to it
	actionDispatcher *ActionDispatcher

	// verbose is used to check weather or not we have to display
	// more logs
	verbose bool
	wsHost  string
	room    string
	name    string

	wsc *websocket.Conn

	normalFont font.Face
)

func init() {
	flag.BoolVar(&verbose, "verbose", false, "Will log all the actions")
	flag.StringVar(&wsHost, "ws-host", ":5555", "The host of the server, the format is 'host:port'")
	flag.StringVar(&room, "room", "room", "The room to connect to")
	flag.StringVar(&name, "name", "john doe", "The name of the player")

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
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.Parse()

	screenW := 288
	screenH := 240

	ebiten.SetWindowTitle("LTW")
	ebiten.SetWindowSize(screenW*2, screenH*2)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	dispatcher := flux.NewDispatcher()

	actionDispatcher = NewActionDispatcher(dispatcher)

	if verbose {
		NewLoggerStore(dispatcher)
	}

	s := store.NewStore(dispatcher)
	m := store.NewMap(dispatcher, s)
	g := &Game{
		Store: s,
		Map:   m,
	}

	// Establish connection
	u := url.URL{Scheme: "ws", Host: wsHost, Path: "/ws"}

	var err error
	wsc, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer wsc.Close()

	go wsHandler()

	err = wsc.WriteJSON(action.NewJoinRoom(room, name))
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Change this to pass the specific store needed instead of all the game object
	cs := NewCameraStore(dispatcher, m, screenW, screenH)
	g.Camera = cs
	g.Units = NewUnits(g)
	g.Towers, err = NewTowers(g)
	if err != nil {
		log.Fatal(err)
	}
	g.HUD, err = NewHUDStore(dispatcher, g)
	if err != nil {
		log.Fatal(err)
	}

	l, err := NewLobby(dispatcher, s, cs)
	if err != nil {
		log.Fatal(err)
	}
	rs := NewRouterStore(dispatcher, g, l)

	if err := ebiten.RunGame(rs); err != nil {
		log.Fatal(err)
	}
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
