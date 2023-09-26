package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
)

var upgrader = websocket.Upgrader{}
var (
	port string

	// actionDispatcher is the main dispatcher of the application
	// all the actions have to be registered to it
	actionDispatcher *ActionDispatcher
)

func init() {
	flag.StringVar(&port, "port", ":5555", "The port of the application with the ':'")
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dispatcher := flux.NewDispatcher()

	actionDispatcher = NewActionDispatcher(dispatcher)

	rooms := NewRoomsStore(dispatcher)

	go startRoomsLoop(ctx, rooms)
	http.HandleFunc("/ws", wsHandler(rooms))
	log.Printf("Staring server at %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func wsHandler(rooms *RoomsStore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, _ := upgrader.Upgrade(w, r, nil)
		defer ws.Close()

		for {
			var msg action.Action
			// If there is an error parsing the msg
			// we kick the user
			err := ws.ReadJSON(&msg)
			if err != nil {
				fmt.Printf("Error when reading the WS message: %w\n", err)
				// TODO: Remove the player from Room and Players
				ws.Close()
				break
			}

			switch msg.Type {
			case action.JoinRoom:
				// Fist we action JoinRoom
				actionDispatcher.Dispatch(&msg)

				// Then we have to add the Player
				sid := uuid.Must(uuid.NewV4())
				nextID := rooms.GetNextID(msg.JoinRoom.Room)
				aap := action.NewAddPlayer(msg.JoinRoom.Room, sid.String(), msg.JoinRoom.Name, nextID, ws)
				aap.Room = msg.JoinRoom.Room
				actionDispatcher.Dispatch(aap)
			default:
				actionDispatcher.Dispatch(&msg)
			}
		}
	}
}

func startRoomsLoop(ctx context.Context, rooms *RoomsStore) {
	stateTicker := time.NewTicker(time.Second / 4)
	incomeTicker := time.NewTicker(time.Second)
	// The default TPS on of Ebiten client if 60 so to
	// emulate that we trigger the move action every TPS
	moveTicker := time.NewTicker(time.Second / 60)
	for {
		select {
		case <-stateTicker.C:
			// TODO: Send state
			actionDispatcher.UpdateState(rooms)
		case <-incomeTicker.C:
			actionDispatcher.IncomeTick(rooms)
		case <-moveTicker.C:
			actionDispatcher.MoveUnit(rooms)
		case <-ctx.Done():
			stateTicker.Stop()
			incomeTicker.Stop()
			goto FINISH
		}
	}
FINISH:
}
