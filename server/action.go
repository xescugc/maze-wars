package server

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/utils"
	"nhooyr.io/websocket"
)

// ActionDispatcher is in charge of dispatching actions to the
// application dispatcher
type ActionDispatcher struct {
	dispatcher *flux.Dispatcher
	store      *Store
	logger     *slog.Logger
	ws         WSConnector
}

// NewActionDispatcher initializes the action dispatcher
// with the give dispatcher
func NewActionDispatcher(d *flux.Dispatcher, l *slog.Logger, s *Store, ws WSConnector) *ActionDispatcher {
	return &ActionDispatcher{
		dispatcher: d,
		store:      s,
		logger:     l,
		ws:         ws,
	}
}

// Dispatch is a helper to access to the internal dispatch directly with an action.
// This should only be used from the WS Handler to forward server actions directly
func (ac *ActionDispatcher) Dispatch(a *action.Action) {
	b := time.Now()
	defer utils.LogTime(ac.logger, b, "action dispatch", "action", a.Type)

	switch a.Type {
	case action.JoinWaitingRoom:
		ac.dispatcher.Dispatch(a)

		ac.startGame()
	default:
		ac.dispatcher.Dispatch(a)
	}
}

func (ac *ActionDispatcher) startGame() {
	wr := ac.store.Rooms.FindCurrentWaitingRoom()
	if wr == nil || (len(wr.Players) != wr.Size) {
		return
	}

	rstate := ac.store.Rooms.GetState().(RoomsState)
	sga := action.NewStartGame()

	ac.Dispatch(sga)
	ac.SyncState(ac.store.Rooms)

	for _, p := range rstate.Rooms[wr.Name].Players {
		err := ac.ws.Write(context.Background(), p.Conn, sga)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (ac *ActionDispatcher) IncomeTick(rooms *RoomsStore) {
	ita := action.NewIncomeTick()
	ac.Dispatch(ita)
}

func (ac *ActionDispatcher) WaitRoomCountdownTick() {
	wrcta := action.NewWaitRoomCountdownTick()
	ac.Dispatch(wrcta)

	ac.startGame()
}

func (ac *ActionDispatcher) UserSignUp(un string) {
	ac.Dispatch(action.NewUserSignUp(un))
}

func (ac *ActionDispatcher) UserSignIn(un, ra string, ws *websocket.Conn) {
	a := action.NewUserSignIn(un)
	a.UserSignIn.RemoteAddr = ra
	a.UserSignIn.Websocket = ws
	a.UserSignIn.Username = un
	ac.Dispatch(a)
}

func (ac *ActionDispatcher) UserSignOut(un string) {
	ac.Dispatch(action.NewUserSignOut(un))
}

func (ac *ActionDispatcher) SyncState(rooms *RoomsStore) {
	ac.Dispatch(action.NewTPS(time.Now()))
	rstate := rooms.GetState().(RoomsState)
	for _, r := range rstate.Rooms {
		if r.Name == rstate.CurrentWaitingRoom {
			continue
		}
		for id, pc := range r.Players {
			// Players
			players := make(map[string]*action.SyncStatePlayerPayload)
			ps := r.Game.Players.GetState().(store.PlayersState)
			for idp, p := range ps.Players {
				uspp := action.SyncStatePlayerPayload{
					ID:          p.ID,
					Name:        p.Name,
					Lives:       p.Lives,
					LineID:      p.LineID,
					Income:      p.Income,
					Gold:        p.Gold,
					Current:     p.Current,
					Winner:      p.Winner,
					UnitUpdates: make(map[string]action.SyncStatePlayerUnitUpdatePayload),
				}
				for t, uu := range p.UnitUpdates {
					uspp.UnitUpdates[t] = action.SyncStatePlayerUnitUpdatePayload(uu)
				}
				if id == idp {
					uspp.Current = true
				}
				players[idp] = &uspp
			}

			// Lines
			lines := make(map[int]*action.SyncStateLinePayload)
			for i, l := range r.Game.Store.Lines.GetState().(store.LinesState).Lines {
				// Towers
				towers := make(map[string]*action.SyncStateTowerPayload)
				for _, t := range l.Towers {
					ustp := action.SyncStateTowerPayload(*t)
					towers[t.ID] = &ustp
				}

				// Units
				units := make(map[string]*action.SyncStateUnitPayload)
				for _, u := range l.Units {
					usup := action.SyncStateUnitPayload(*u)
					units[u.ID] = &usup
				}
				lines[i] = &action.SyncStateLinePayload{
					Towers: towers,
					Units:  units,
				}
			}

			aus := action.NewSyncState(
				&action.SyncStatePlayersPayload{
					Players:     players,
					IncomeTimer: ps.IncomeTimer,
				},
				&action.SyncStateLinesPayload{
					Lines: lines,
				},
			)
			err := ac.ws.Write(context.Background(), pc.Conn, aus)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func (ac *ActionDispatcher) SyncUsers(users *UsersStore) {
	for _, u := range users.List() {
		// This will carry more information on the future
		// potentially more customized to the current user
		auu := action.NewSyncUsers(
			len(users.List()),
		)
		err := ac.ws.Write(context.Background(), u.Conn, auu)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (ac *ActionDispatcher) SyncWaitingRoom(rooms *RoomsStore) {
	rstate := rooms.GetState().(RoomsState)
	if rstate.CurrentWaitingRoom != "" {
		cwr := rstate.Rooms[rstate.CurrentWaitingRoom]
		swra := action.NewSyncWaitingRoom(
			len(cwr.Players),
			cwr.Size,
			cwr.Countdown,
		)
		for _, p := range cwr.Players {
			err := ac.ws.Write(context.Background(), p.Conn, swra)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
