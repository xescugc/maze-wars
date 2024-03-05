package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	cutils "github.com/xescugc/maze-wars/client/utils"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/utils"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// ActionDispatcher is in charge of dispatching actions to the
// application dispatcher
type ActionDispatcher struct {
	dispatcher *flux.Dispatcher
	opt        Options
	store      *store.Store
	logger     *slog.Logger
}

// NewActionDispatcher initializes the action dispatcher
// with the give dispatcher
func NewActionDispatcher(d *flux.Dispatcher, s *store.Store, l *slog.Logger, opt Options) *ActionDispatcher {
	return &ActionDispatcher{
		dispatcher: d,
		opt:        opt,
		store:      s,
		logger:     l,
	}
}

// Dispatch is a helper to access to the internal dispatch directly with an action.
// This should only be used from the WS Handler to forward server actions directly
func (ac *ActionDispatcher) Dispatch(a *action.Action) {
	b := time.Now()
	defer utils.LogTime(ac.logger, b, "action dispatched", "action", a.Type)

	ac.dispatcher.Dispatch(a)
}

// WindowResizing new sizes of the window
func (ac *ActionDispatcher) WindowResizing(w, h int) {
	wr := action.NewWindowResizing(w, h)
	ac.Dispatch(wr)
}

// NavigateTo navigates to the given route
func (ac *ActionDispatcher) NavigateTo(route string) {
	nt := action.NewNavigateTo(route)
	ac.Dispatch(nt)
}

func (ac *ActionDispatcher) SignUpSubmit(un string) {
	httpu, _ := url.Parse(ac.opt.HostURL)
	httpu.Path = "/users"
	resp, err := http.Post(httpu.String(), "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{"username":"%s"}`, un))))
	if err != nil {
		ac.Dispatch(action.NewSignUpError(err.Error()))
		return
	}
	body := struct {
		Error string `json:"error"`
	}{}
	if resp.StatusCode != http.StatusCreated {
		err = json.NewDecoder(resp.Body).Decode(&body)
		if err != nil {
			ac.Dispatch(action.NewSignUpError(err.Error()))
			return
		}
		ac.Dispatch(action.NewSignUpError(body.Error))
		return
	}

	ac.Dispatch(action.NewSignUpError(""))

	ctx := context.Background()

	// We manually clone it to then change it for the WS
	chttpu := *httpu
	wsu := &chttpu
	wsu.Scheme = "ws"
	if httpu.Scheme == "https" {
		wsu.Scheme = "wss"
	}
	wsu.Path = "/ws"

	wsc, _, err = websocket.Dial(ctx, wsu.String(), nil)
	if err != nil {
		panic(fmt.Errorf("failed to dial the server %q: %w", wsu.String(), err))
	}

	wsc.SetReadLimit(-1)

	usia := action.NewUserSignIn(un)
	err = wsjson.Write(ctx, wsc, usia)
	if err != nil {
		panic(fmt.Errorf("failed to write JSON: %w", err))
	}

	ac.Dispatch(usia)

	go wsHandler(ctx)

	ac.Dispatch(action.NewNavigateTo(cutils.LobbyRoute))
}

// GoHome will move the camera to the current player home line
func (ac *ActionDispatcher) GoHome() {
	gha := action.NewGoHome()
	ac.Dispatch(gha)
}

func (ac *ActionDispatcher) JoinWaitingRoom(un string) {
	jwra := action.NewJoinWaitingRoom(un)
	wsSend(jwra)

	ac.Dispatch(action.NewNavigateTo(cutils.WaitingRoomRoute))
}

func (ac *ActionDispatcher) ExitWaitingRoom(un string) {
	ewra := action.NewExitWaitingRoom(un)
	wsSend(ewra)

	ac.Dispatch(action.NewNavigateTo(cutils.LobbyRoute))
}
