package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/server/assets"
	"github.com/xescugc/maze-wars/server/templates"
)

var (
	// actionDispatcher is the main dispatcher of the application
	// all the actions have to be registered to it
	actionDispatcher *ActionDispatcher
)

func New(ad *ActionDispatcher, rooms *RoomsStore, opt Options) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	actionDispatcher = ad

	go startRoomsLoop(ctx, rooms)

	r := mux.NewRouter()

	r.HandleFunc("/ws", wsHandler(rooms)).Methods(http.MethodGet)

	r.HandleFunc("/rooms", roomsCreateHandler).Methods(http.MethodPost)
	r.HandleFunc("/rooms/new", roomsNewHandler).Methods(http.MethodGet)
	r.HandleFunc("/rooms/{room}", roomsShowHandler).Methods(http.MethodGet)
	r.HandleFunc("/game", gameHandler).Methods(http.MethodGet)
	r.HandleFunc("/", homeHandler).Methods(http.MethodGet)

	hmux := http.NewServeMux()
	hmux.Handle("/", r)
	hmux.Handle("/css/", http.FileServer(http.FS(assets.Assets)))
	hmux.Handle("/js/", http.FileServer(http.FS(assets.Assets)))
	hmux.Handle("/wasm/", http.FileServer(http.FS(assets.Assets)))

	svr := &http.Server{
		Addr:    fmt.Sprintf(":%s", opt.Port),
		Handler: handlers.LoggingHandler(os.Stdout, hmux),
	}

	log.Printf("Staring server at %s\n", opt.Port)
	if err := svr.ListenAndServe(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := templates.Templates["views/home/index.tmpl"]
	t.Execute(w, nil)
}

func roomsShowHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := templates.Templates["views/rooms/show.tmpl"]
	t.Execute(w, nil)
}

func roomsNewHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := templates.Templates["views/rooms/new.tmpl"]
	t.Execute(w, nil)
}

func roomsCreateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.FormValue("name")
	room := r.FormValue("room")

	w.Header().Set("Location", fmt.Sprintf("/rooms/%s?name=%s", room, name))
	w.WriteHeader(http.StatusSeeOther)
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := templates.Templates["views/game/game.tmpl"]
	t.Execute(w, nil)
}

func wsHandler(rooms *RoomsStore) func(http.ResponseWriter, *http.Request) {
	return func(hw http.ResponseWriter, hr *http.Request) {
		ws, _ := websocket.Accept(hw, hr, nil)
		defer ws.CloseNow()

		for {
			var msg action.Action
			// If there is an error parsing the msg
			// we kick the user
			err := wsjson.Read(hr.Context(), ws, &msg)
			if err != nil {
				fmt.Printf("Error when reading the WS message: %s\n", err)

				for rn, r := range rooms.GetState().(RoomsState).Rooms {
					if uid, ok := r.Connections[hr.RemoteAddr]; ok {
						actionDispatcher.RemovePlayer(rn, uid)
						break
					}
				}
				break
			}

			switch msg.Type {
			case action.JoinRoom:
				// Fist we action JoinRoom
				actionDispatcher.Dispatch(&msg)

				// Then we have to add the Player
				sid := uuid.Must(uuid.NewV4())
				nextID := rooms.GetNextID(msg.JoinRoom.Room)
				aap := action.NewAddPlayer(msg.JoinRoom.Room, sid.String(), msg.JoinRoom.Name, nextID, ws, hr.RemoteAddr)
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
			actionDispatcher.TPS(rooms)
		case <-ctx.Done():
			stateTicker.Stop()
			incomeTicker.Stop()
			goto FINISH
		}
	}
FINISH:
}
