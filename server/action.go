package server

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
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

const noRoomID = ""

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

		ac.startGame(noRoomID)
	case action.DeleteLobby:
		l := ac.store.Lobbies.FindByID(a.DeleteLobby.LobbyID)
		ac.dispatcher.Dispatch(a)
		ac.notifyPlayersLobbyDeleted(l.Players)
	case action.UserSignOut:
		// TODO: Kill the bot
		u, _ := ac.store.Users.FindByUsername(a.UserSignOut.Username)
		l := ac.store.Lobbies.FindByID(u.CurrentLobbyID)
		ac.dispatcher.Dispatch(a)
		if l != nil {
			nl := ac.store.Lobbies.FindByID(u.CurrentLobbyID)
			// It has been deleted
			if nl == nil {
				ac.notifyPlayersLobbyDeleted(l.Players)
			}
		}
	case action.StartLobby:
		// Order of things:
		// * I create the Room from the Lobby
		// * I start the room
		// * I delete the lobby
		ac.dispatcher.Dispatch(a)
		ac.startGame(a.StartLobby.LobbyID)
		ac.dispatcher.Dispatch(action.NewDeleteLobby(a.StartLobby.LobbyID))
	default:
		ac.dispatcher.Dispatch(a)
	}
}

func (ac *ActionDispatcher) notifyPlayersLobbyDeleted(uns map[string]bool) {
	lbs := ac.store.Lobbies.List()
	albs := make([]*action.LobbyPayload, 0, len(lbs))
	for _, l := range lbs {
		albs = append(albs, &action.LobbyPayload{
			ID:         l.ID,
			Name:       l.Name,
			MaxPlayers: l.MaxPlayers,
			Players:    l.Players,
			Owner:      l.Owner,
		})
	}
	ala := action.NewAddLobbies(&action.AddLobbiesPayload{Lobbies: albs})
	nto := action.NewNavigateTo(utils.LobbiesRoute)

	for un, ib := range uns {
		// We skip bots
		if ib {
			continue
		}
		u, ok := ac.store.Users.FindByUsername(un)
		if !ok {
			continue
		}

		err := ac.ws.Write(context.Background(), u.Conn, ala)
		if err != nil {
			log.Fatal(err)
		}
		err = ac.ws.Write(context.Background(), u.Conn, nto)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (ac *ActionDispatcher) startGame(roid string) {
	var rid = roid
	// If no rid is passed then the WR is the one chosen and
	// will only start if it has the number of players
	if rid == "" {
		r := ac.store.Rooms.FindCurrentWaitingRoom()
		if r == nil || (len(r.Players) != r.Size) {
			return
		}
		rid = r.Name
	}

	rstate := ac.store.Rooms.GetState().(RoomsState)
	sga := action.NewStartGame()
	sra := action.NewStartRoom(rid)

	ac.Dispatch(sra)
	ac.SyncState(ac.store.Rooms)

	for pid, p := range rstate.Rooms[rid].Players {
		// We do not need to communicate with the bots
		if p.IsBot {
			rstate.Rooms[rid].Bots[pid].Start()
			continue
		}
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

	ac.startGame(noRoomID)
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
	rms := rooms.List()
	cwr := rooms.FindCurrentWaitingRoom()
	for _, r := range rms {
		if cwr != nil && r.Name == cwr.Name {
			continue
		}
		for id, pc := range r.Players {
			// We do not want to communicate state to a bot
			if pc.IsBot {
				continue
			}
			// Players
			players := make(map[string]*action.SyncStatePlayerPayload)
			lplayers := r.Game.Store.Lines.ListPlayers()
			for _, p := range lplayers {
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
				if id == p.ID {
					uspp.Current = true
				}
				players[p.ID] = &uspp
			}

			// Lines
			lines := make(map[int]*action.SyncStateLinePayload)
			llines := r.Game.Store.Lines.ListLines()
			for i, l := range llines {
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
					IncomeTimer: r.Game.Store.Lines.GetIncomeTimer(),
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

// SyncLobbies will just sync the info of each lobby to the players on it
func (ac *ActionDispatcher) SyncLobbies(s *Store) {
	lbs := s.Lobbies.List()
	for _, l := range lbs {
		al := action.LobbyPayload{
			ID:         l.ID,
			Name:       l.Name,
			MaxPlayers: l.MaxPlayers,
			Players:    l.Players,
			Owner:      l.Owner,
		}
		ula := action.NewUpdateLobby(al)
		for p, ib := range l.Players {
			// If is bot we skip it
			if ib {
				continue
			}
			u, ok := s.Users.FindByUsername(p)
			if ok {
				err := ac.ws.Write(context.Background(), u.Conn, ula)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
