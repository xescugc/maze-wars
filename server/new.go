package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"

	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/server/assets"
	"github.com/xescugc/maze-wars/server/models"
	"github.com/xescugc/maze-wars/server/templates"
	"github.com/xescugc/maze-wars/unit"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
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

	// Game Websocket
	r.HandleFunc("/ws", wsHandler(s)).Methods(http.MethodGet)

	// Webpage
	r.HandleFunc("/play", playHandler(opt.Version)).Methods(http.MethodGet)
	r.HandleFunc("/download", downloadHandler(opt.Version)).Methods(http.MethodGet)
	r.HandleFunc("/game", gameHandler(opt.Version)).Methods(http.MethodGet)
	r.HandleFunc("/docs", docsHandler(opt.Version)).Methods(http.MethodGet)
	r.HandleFunc("/", homeHandler(opt.Version)).Methods(http.MethodGet)

	// Game Endpoints
	r.HandleFunc("/users", usersCreateHandler(s)).Methods(http.MethodPost).Headers("Content-Type", "application/json")
	r.HandleFunc("/version", versionHandler(opt.Version)).Methods(http.MethodPost).Headers("Content-Type", "application/json")

	r.HandleFunc("/lobbies", listLobbiesHandler(s)).Methods(http.MethodGet)

	hmux := http.NewServeMux()
	hmux.Handle("/", r)
	hmux.Handle("/css/", http.FileServer(http.FS(assets.Assets)))
	hmux.Handle("/js/", http.FileServer(http.FS(assets.Assets)))
	hmux.Handle("/wasm/", http.FileServer(http.FS(assets.Assets)))
	hmux.Handle("/images/", http.FileServer(http.FS(assets.Assets)))
	hmux.Handle("/metrics", promhttp.Handler())

	handler := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	}).Handle(hmux)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	})

	svr := &http.Server{
		Addr:    fmt.Sprintf(":%s", opt.Port),
		Handler: handlers.LoggingHandler(os.Stdout, c.Handler(handler)),
	}

	log.Printf("Staring server at %s\n", opt.Port)
	if err := svr.ListenAndServe(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

type templateData struct {
	Tab     string
	Version string
}

func homeHandler(v string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t, _ := templates.Templates["views/home/index.tmpl"]
		t.Execute(w, templateData{Tab: "home", Version: v})
	}
}

func playHandler(v string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t, _ := templates.Templates["views/game/play.tmpl"]
		t.Execute(w, templateData{Tab: "game", Version: v})
	}
}

func downloadHandler(v string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t, _ := templates.Templates["views/game/download.tmpl"]
		t.Execute(w, templateData{Tab: "download", Version: v})
	}
}

func gameHandler(v string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t, _ := templates.Templates["views/game/game.tmpl"]
		t.Execute(w, map[string]interface{}{"version": v})
	}
}

func docsHandler(v string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t, _ := templates.Templates["views/docs/index.tmpl"]
		t.Execute(w, templateData{Tab: "docs", Version: v})
	}
}

type usersCreateRequest struct {
	Username string `json:"username"`
	ImageKey string `json:"image_key"`
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
		if ucr.ImageKey == "" {
			ucr.ImageKey = unit.TypeStrings()[0]
		}

		if _, ok := s.Rooms.FindUserByUsername(ucr.Username); ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{Error: "User already exists"})
			return
		}

		if _, ok := s.Rooms.FindUserByRemoteAddress(r.RemoteAddr); ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{Error: "A session already exists from this computer"})
			return
		}

		actionDispatcher.UserSignUp(ucr.Username, ucr.ImageKey)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
	}
}

func listLobbiesHandler(s *Store) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		lobbies := s.Lobbies.List()
		respLobbies := models.LobbiesResponse{
			Lobbies: make([]models.LobbyResponse, 0, len(lobbies)),
		}

		for _, l := range lobbies {
			lr := models.LobbyResponse{
				ID:         l.ID,
				Name:       l.Name,
				MaxPlayers: l.MaxPlayers,
				Owner:      l.Owner,
				Players:    l.Players,
			}
			respLobbies.Lobbies = append(respLobbies.Lobbies, lr)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(respLobbies)
	}
}

type versionRequest struct {
	Version string `json:"version"`
}

func versionHandler(v string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var vr versionRequest

		err := json.NewDecoder(r.Body).Decode(&vr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
			return
		}

		if vr.Version != v {
			w.WriteHeader(http.StatusBadRequest)

			json.NewEncoder(w).Encode(errorResponse{Error: fmt.Sprintf("The client version (%q) is outdated, download the new version %q", vr.Version, v)})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func wsHandler(s *Store) func(http.ResponseWriter, *http.Request) {
	return func(hw http.ResponseWriter, hr *http.Request) {
		ws, err := websocket.Accept(hw, hr, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			hw.WriteHeader(http.StatusBadRequest)

			json.NewEncoder(hw).Encode(errorResponse{Error: fmt.Errorf("Failed to accept websocket connection: %w", err).Error()})
			return
		}
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
				u, _ := s.Rooms.FindUserByRemoteAddress(hr.RemoteAddr)
				fmt.Printf("Error when reading the WS message: %s\n", err)

				actionDispatcher.UserSignOut(u.Username)
				break
			}

			u, _ := s.Rooms.FindUserByRemoteAddress(hr.RemoteAddr)

			// If the User is in a Room we set it directly on the
			// action from the handler
			msg.Room = u.CurrentRoomID

			switch msg.Type {
			case action.UserSignIn:
				actionDispatcher.UserSignIn(*&msg.UserSignIn.Username, hr.RemoteAddr, ws)
			case action.RemovePlayer:
				actionDispatcher.Dispatch(&msg)
			default:
				actionDispatcher.Dispatch(&msg)
			}
		}
	}
}

func startLoop(ctx context.Context, s *Store) {
	defer func() {
		err := recover()

		if err != nil {
			sentry.CurrentHub().Recover(err)
			sentry.Flush(time.Second * 5)
			if Environment == "dev" {
				panic(err)
			}
		}
	}()
	secondTicker := time.NewTicker(time.Second)
	stateTicker := time.NewTicker(time.Second / 4)
	for {
		select {
		// TODO: All this should be calling actionDispatcher.Dispatch(Action)
		// so then I funnel all of them always through the dispatcher
		case <-stateTicker.C:
			actionDispatcher.Dispatch(&action.Action{Type: action.SyncState})
		case <-secondTicker.C:
			actionDispatcher.IncomeTick()
			actionDispatcher.Dispatch(&action.Action{Type: action.SyncLobbies})
			actionDispatcher.Dispatch(&action.Action{Type: action.SyncWaitingRooms})
		case <-ctx.Done():
			stateTicker.Stop()
			secondTicker.Stop()
			goto FINISH
		}
	}
FINISH:
}
