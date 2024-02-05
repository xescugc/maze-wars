package server

import (
	"context"
	"log"
	"time"

	"github.com/sagikazarmark/slog-shim"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/store"
	"nhooyr.io/websocket/wsjson"
)

// ActionDispatcher is in charge of dispatching actions to the
// application dispatcher
type ActionDispatcher struct {
	dispatcher *flux.Dispatcher
	store      *Store
	logger     *slog.Logger
}

// NewActionDispatcher initializes the action dispatcher
// with the give dispatcher
func NewActionDispatcher(d *flux.Dispatcher, l *slog.Logger, s *Store) *ActionDispatcher {
	return &ActionDispatcher{
		dispatcher: d,
		store:      s,
		logger:     l,
	}
}

// Dispatch is a helper to access to the internal dispatch directly with an action.
// This should only be used from the WS Handler to forward server actions directly
func (ac *ActionDispatcher) Dispatch(a *action.Action) {
	b := time.Now()
	switch a.Type {
	case action.JoinWaitingRoom:
		ac.dispatcher.Dispatch(a)

		ac.startGame()
	default:
		ac.dispatcher.Dispatch(a)
	}
	d := time.Now().Sub(b)
	if d > time.Millisecond {
		ac.logger.Info("action dispatched", "type", a.Type, "time", d)
	}
	ac.logger.Debug("action dispatched", "type", a.Type, "time", d)
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
		err := wsjson.Write(context.Background(), p.Conn, sga)
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

func (ac *ActionDispatcher) TPS(rooms *RoomsStore) {
	tpsa := action.NewTPS()
	ac.Dispatch(tpsa)
}

func (ac *ActionDispatcher) UserSignUp(un string) {
	ac.Dispatch(action.NewUserSignUp(un))
}

func (ac *ActionDispatcher) UserSignIn(un string) {
	ac.Dispatch(action.NewUserSignIn(un))
}

func (ac *ActionDispatcher) UserSignOut(un string) {
	ac.Dispatch(action.NewUserSignOut(un))
}

func (ac *ActionDispatcher) SyncState(rooms *RoomsStore) {
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
				uspp := action.SyncStatePlayerPayload(*p)
				if id == idp {
					uspp.Current = true
				}
				players[idp] = &uspp
			}

			// Towers
			towers := make(map[string]*action.SyncStateTowerPayload)
			ts := r.Game.Towers.List()
			for _, t := range ts {
				ustp := action.SyncStateTowerPayload(*t)
				towers[t.ID] = &ustp
			}

			// Units
			units := make(map[string]*action.SyncStateUnitPayload)
			us := r.Game.Units.List()
			for _, u := range us {
				usup := action.SyncStateUnitPayload(*u)
				units[u.ID] = &usup
			}

			aus := action.NewSyncState(
				&action.SyncStatePlayersPayload{
					Players:     players,
					IncomeTimer: ps.IncomeTimer,
				},
				&action.SyncStateTowersPayload{
					Towers: towers,
				},
				&action.SyncStateUnitsPayload{
					Units: units,
				},
			)
			err := wsjson.Write(context.Background(), pc.Conn, aus)
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
		err := wsjson.Write(context.Background(), u.Conn, auu)
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
			err := wsjson.Write(context.Background(), p.Conn, swra)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
