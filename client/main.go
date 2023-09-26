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
	"github.com/xescugc/ltw/store"
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
)

func init() {
	flag.BoolVar(&verbose, "verbose", false, "Will log all the actions")
	flag.StringVar(&wsHost, "ws-host", ":5555", "The host of the server, the format is 'host:port'")
	flag.StringVar(&room, "room", "room", "The room to connect to")
	flag.StringVar(&name, "name", "john doe", "The name of the player")

	rand.Seed(time.Now().UnixNano())
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

	m, err := store.NewMap()
	if err != nil {
		log.Fatal(err)
	}

	store := store.NewStore(dispatcher)
	g := &Game{
		Store:   store,
		Players: NewPlayers(store),
		Map:     m,
	}

	// Establish connection
	u := url.URL{Scheme: "ws", Host: wsHost, Path: "/ws"}

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
	g.Camera = NewCameraStore(dispatcher, g, screenW, screenH)
	g.Units = NewUnits(g)
	g.Towers = NewTowers(g)
	g.HUD, err = NewHUDStore(dispatcher, g)
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(g); err != nil {
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
