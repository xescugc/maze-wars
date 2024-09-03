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
	"github.com/xescugc/maze-wars/server/models"
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
	// If we are going to the LobbiesRoute we need to
	// preload the data
	if route == utils.LobbiesRoute {
		ac.RefreshLobbies()
	}
	ac.Dispatch(nt)
}

func (ac *ActionDispatcher) RefreshLobbies() {
	httpu, _ := url.Parse(ac.opt.HostURL)
	httpu.Path = "/lobbies"
	resp, err := http.Get(httpu.String())
	if err != nil {
		ac.logger.Error(err.Error())
		return
	}
	if resp.StatusCode != http.StatusOK {
		return
	}
	mlbs := &models.LobbiesResponse{}
	err = json.NewDecoder(resp.Body).Decode(&mlbs)
	if err != nil {
		ac.logger.Error(err.Error())
		return
	}

	lbs := make([]*action.LobbyPayload, 0, len(mlbs.Lobbies))
	for _, ml := range mlbs.Lobbies {
		allp := action.LobbyPayload(ml)
		lbs = append(lbs, &allp)
	}
	ac.Dispatch(action.NewAddLobbies(&action.AddLobbiesPayload{Lobbies: lbs}))
}

func (ac *ActionDispatcher) CheckVersion() {
	httpu, _ := url.Parse(ac.opt.HostURL)
	httpu.Path = "/version"
	resp, err := http.Post(httpu.String(), "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{"version":"%s"}`, ac.opt.Version))))
	if err != nil {
		ac.Dispatch(action.NewVersionError(err.Error()))
		return
	}
	body := struct {
		Error string `json:"error"`
	}{}
	if resp.StatusCode != http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&body)
		if err != nil {
			ac.Dispatch(action.NewVersionError(err.Error()))
			return
		}
		ac.Dispatch(action.NewVersionError(body.Error))
		return
	}
	return
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

	ac.Dispatch(action.NewNavigateTo(utils.RootRoute))
}

// GoHome will move the camera to the current player home line
func (ac *ActionDispatcher) GoHome() {
	gha := action.NewGoHome()
	ac.Dispatch(gha)
}

func (ac *ActionDispatcher) JoinVs6WaitingRoom(un string) {
	jwra := action.NewJoinVs6WaitingRoom(un)
	wsSend(jwra)

	ac.Dispatch(action.NewNavigateTo(utils.Vs6WaitingRoomRoute))
}

func (ac *ActionDispatcher) ExitVs6WaitingRoom(un string) {
	ewra := action.NewExitVs6WaitingRoom(un)
	wsSend(ewra)

	ac.Dispatch(action.NewNavigateTo(utils.RootRoute))
}

func (ac *ActionDispatcher) JoinVs1WaitingRoom(un string) {
	jwra := action.NewJoinVs1WaitingRoom(un)
	wsSend(jwra)

	ac.Dispatch(action.NewNavigateTo(utils.Vs1WaitingRoomRoute))
}

func (ac *ActionDispatcher) ExitVs1WaitingRoom(un string) {
	ewra := action.NewExitVs1WaitingRoom(un)
	wsSend(ewra)

	ac.Dispatch(action.NewNavigateTo(utils.RootRoute))
}

func (ac *ActionDispatcher) CreateLobby(lid, o, ln string, lmp int) {
	cla := action.NewCreateLobby(lid, o, ln, lmp)
	wsSend(cla)
	ac.Dispatch(cla)
}

func (ac *ActionDispatcher) SelectLobby(lid string) {
	sla := action.NewSelectLobby(lid)
	ac.Dispatch(sla)
}

func (ac *ActionDispatcher) DeleteLobby(lid string) {
	dla := action.NewDeleteLobby(lid)
	wsSend(dla)
	ac.Dispatch(dla)
}

func (ac *ActionDispatcher) LeaveLobby(lid, un string) {
	lla := action.NewLeaveLobby(lid, un)
	wsSend(lla)
	ac.Dispatch(lla)
}

func (ac *ActionDispatcher) JoinLobby(lid, un string, ib bool) {
	jla := action.NewJoinLobby(lid, un, ib)
	wsSend(jla)
	ac.Dispatch(jla)
}

func (ac *ActionDispatcher) StartLobby(lid string) {
	sla := action.NewStartLobby(lid)
	wsSend(sla)
}

func (ac *ActionDispatcher) SetupGame(d bool) {
	sga := action.NewSetupGame(d)
	ac.Dispatch(sga)
}

func (ac *ActionDispatcher) FindGame(un string, vs, ranked, vsBots bool) {
	fga := action.NewFindGame(un, vs, ranked, vsBots)
	wsSend(fga)
	ac.Dispatch(fga)
}

func (ac *ActionDispatcher) ExitSearchingGame(un string) {
	esga := action.NewExitSearchingGame(un)
	wsSend(esga)
	ac.Dispatch(esga)
}

func (ac *ActionDispatcher) AcceptWaitingGame(un string) {
	awga := action.NewAcceptWaitingGame(un)
	wsSend(awga)
}

func (ac *ActionDispatcher) SeenLobbies() {
	lsa := action.NewSeenLobbies()
	ac.Dispatch(lsa)
}

func (ac *ActionDispatcher) CancelWaitingGame(un string) {
	cwga := action.NewCancelWaitingGame(un)
	wsSend(cwga)
}
