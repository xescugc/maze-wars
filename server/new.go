package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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

func New(ad *ActionDispatcher, s *Store, opt Options) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	actionDispatcher = ad

	go startLoop(ctx, s)

	r := mux.NewRouter()

	r.HandleFunc("/ws", wsHandler(s)).Methods(http.MethodGet)

	r.HandleFunc("/play", playHandler).Methods(http.MethodGet)
	r.HandleFunc("/game", gameHandler).Methods(http.MethodGet)
	r.HandleFunc("/docs", docsHandler).Methods(http.MethodGet)
	r.HandleFunc("/", homeHandler).Methods(http.MethodGet)

	r.HandleFunc("/users", usersCreateHandler(s)).Methods(http.MethodPost).Headers("Content-Type", "application/json")

	hmux := http.NewServeMux()
	hmux.Handle("/", r)
	hmux.Handle("/css/", http.FileServer(http.FS(assets.Assets)))
	hmux.Handle("/js/", http.FileServer(http.FS(assets.Assets)))
	hmux.Handle("/wasm/", http.FileServer(http.FS(assets.Assets)))
	hmux.Handle("/images/", http.FileServer(http.FS(assets.Assets)))

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

type templateData struct {
	Tab string
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := templates.Templates["views/home/index.tmpl"]
	t.Execute(w, templateData{"home"})
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := templates.Templates["views/game/play.tmpl"]
	t.Execute(w, templateData{"game"})
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := templates.Templates["views/game/game.tmpl"]
	t.Execute(w, nil)
}

func docsHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := templates.Templates["views/docs/index.tmpl"]
	t.Execute(w, templateData{"docs"})
}

type usersCreateRequest struct {
	Username string `json:"username"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func usersCreateHandler(s *Store) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var ucr usersCreateRequest

		err := json.NewDecoder(r.Body).Decode(&ucr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
			return
		}

		if ucr.Username == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{Error: "Username cannot be empty"})
			return
		}

		if _, ok := s.Users.FindByUsername(ucr.Username); ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{Error: "User already exists"})
			return
		}

		if _, ok := s.Users.FindByRemoteAddress(r.RemoteAddr); ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{Error: "A session already exists from this computer"})
			return
		}

		actionDispatcher.UserSignUp(ucr.Username)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
	}
}

func wsHandler(s *Store) func(http.ResponseWriter, *http.Request) {
	return func(hw http.ResponseWriter, hr *http.Request) {
		ws, _ := websocket.Accept(hw, hr, nil)
		defer ws.CloseNow()

		for {
			var msg action.Action
			// If there is an error parsing the msg
			// we kick the user
			err := wsjson.Read(hr.Context(), ws, &msg)
			if err != nil {
				// We cannot move this 'u' call outside as the Read
				// block until a new message is received so it may have
				// a wrong value stored inside
				u, _ := s.Users.FindByRemoteAddress(hr.RemoteAddr)
				fmt.Printf("Error when reading the WS message: %s\n", err)

				actionDispatcher.UserSignOut(u.Username)
				break
			}

			u, _ := s.Users.FindByRemoteAddress(hr.RemoteAddr)

			// If the User is in a Room we set it directly on the
			// action from the handler
			msg.Room = u.CurrentRoomID

			switch msg.Type {
			case action.UserSignIn:
				// We need to append this extra information to the Action
				msg.UserSignIn.Websocket = ws
				msg.UserSignIn.RemoteAddr = hr.RemoteAddr
				actionDispatcher.Dispatch(&msg)
			default:
				actionDispatcher.Dispatch(&msg)
			}
		}
	}
}

func startLoop(ctx context.Context, s *Store) {
	stateTicker := time.NewTicker(time.Second / 4)
	incomeTicker := time.NewTicker(time.Second)
	// The default TPS on of Ebiten client if 60 so to
	// emulate that we trigger the move action every TPS
	moveTicker := time.NewTicker(time.Second / 60)
	usersTicker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-stateTicker.C:
			actionDispatcher.SyncState(s.Rooms)
		case <-incomeTicker.C:
			actionDispatcher.IncomeTick(s.Rooms)
			actionDispatcher.WaitRoomCountdownTick()
			actionDispatcher.SyncWaitingRoom(s.Rooms)
		case <-moveTicker.C:
			actionDispatcher.TPS(s.Rooms)
		case <-usersTicker.C:
			actionDispatcher.SyncUsers(s.Users)
		case <-ctx.Done():
			stateTicker.Stop()
			incomeTicker.Stop()
			moveTicker.Stop()
			usersTicker.Stop()
			goto FINISH
		}
	}
FINISH:
}
