package server

import (
	"context"
	"log"
	"log/slog"
	"sort"
	"time"

	"github.com/coder/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/unit"
	"github.com/xescugc/maze-wars/utils"
)

// ActionDispatcher is in charge of dispatching actions to the
// application dispatcher
type ActionDispatcher struct {
	dispatcher *flux.Dispatcher[*action.Action]
	store      *Store
	logger     *slog.Logger
	ws         WSConnector
}

const (
	noRoomID = ""

	noVs = "no-vs"
	vs1  = "vs1"
	vs6  = "vs6"
)

// NewActionDispatcher initializes the action dispatcher
// with the give dispatcher
func NewActionDispatcher(d *flux.Dispatcher[*action.Action], l *slog.Logger, s *Store, ws WSConnector) *ActionDispatcher {
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
	case action.DeleteLobby:
		l := ac.store.Lobbies.FindByID(a.DeleteLobby.LobbyID)
		ac.dispatcher.Dispatch(a)
		ac.notifyPlayersLobbyDeleted(l.Players)
	case action.UserSignOut:
		// TODO: Kill the bot
		u, _ := ac.store.Rooms.FindUserByUsername(a.UserSignOut.Username)
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
		ac.startGame(noVs, a.StartLobby.LobbyID)
		ac.dispatcher.Dispatch(action.NewDeleteLobby(a.StartLobby.LobbyID))
	case action.SyncState:
		ac.syncState()
	case action.SyncLobbies:
		ac.syncLobbies()
	case action.SyncWaitingRooms:
		ac.syncWaitingRooms()
	default:
		ac.dispatcher.Dispatch(a)
	}

	numberOfActions.With(prometheus.Labels{"type": a.Type.String()}).Observe(float64(time.Now().Sub(b)))
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
		u, ok := ac.store.Rooms.FindUserByUsername(un)
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

func (ac *ActionDispatcher) startGame(vs, roid string) {
	var rid = roid
	// If no rid is passed then the WR is the one chosen and
	// will only start if it has the number of players
	if rid == "" {
		var r *Room
		if r == nil || (len(r.Players) != r.Size) {
			return
		}
		rid = r.Name
	}

	state := ac.store.Rooms.GetState()
	r := state.Rooms[rid]

	ac.syncState()

	for pid, p := range r.Players {
		// We do not need to communicate with the bots
		if p.IsBot {
			state.Rooms[rid].Bots[pid].Start()
			continue
		}
		sga := action.NewStartGame(ac.store.Rooms.SyncState(r, pid))
		err := ac.ws.Write(context.Background(), p.User.Conn, sga)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (ac *ActionDispatcher) IncomeTick() {
	ita := action.NewIncomeTick()
	ac.Dispatch(ita)
}

func (ac *ActionDispatcher) UserSignUp(un, ik string) {
	ac.Dispatch(action.NewUserSignUp(un, ik))
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

func (ac *ActionDispatcher) syncState() {
	ac.Dispatch(action.NewTPS(time.Now()))
	rms := ac.store.Rooms.ListRooms()
	for _, r := range rms {
		for id, pc := range r.Players {
			// We do not want to communicate state to a bot
			if pc.IsBot {
				continue
			}
			// Players
			players := make(map[string]*action.SyncStatePlayerPayload)
			lplayers := r.Game.Store.Game.ListPlayers()
			for _, p := range lplayers {
				ap := p
				uspp := action.SyncStatePlayerPayload{
					ID:          ap.ID,
					Name:        ap.Name,
					ImageKey:    ap.ImageKey,
					Lives:       ap.Lives,
					LineID:      ap.LineID,
					Income:      ap.Income,
					Gold:        ap.Gold,
					IsBot:       ap.IsBot,
					Current:     ap.Current,
					Winner:      ap.Winner,
					Capacity:    ap.Capacity,
					UnitUpdates: make(map[string]action.SyncStatePlayerUnitUpdatePayload),
				}
				// TODO: Make it concurrently safe
				for t, uu := range ap.UnitUpdates {
					uspp.UnitUpdates[t] = action.SyncStatePlayerUnitUpdatePayload(uu)
				}
				if id == ap.ID {
					uspp.Current = true
				}
				players[ap.ID] = &uspp
			}

			// Lines
			lines := make(map[int]*action.SyncStateLinePayload)
			llines := r.Game.Store.Game.ListLines()
			for _, l := range llines {
				al := l
				// Towers
				towers := make(map[string]*action.SyncStateTowerPayload)
				for _, t := range al.Towers {
					at := t
					payload := action.SyncStateTowerPayload(*at)
					towers[at.ID] = &payload
				}

				// Units
				units := make(map[string]*action.SyncStateUnitPayload)
				for _, u := range al.Units {
					au := u
					payload := action.SyncStateUnitPayload(*au)
					units[au.ID] = &payload
				}

				// Projectiles
				projectiles := make(map[string]*action.SyncStateProjectilePayload)
				for _, p := range al.Projectiles {
					ap := p
					payload := action.SyncStateProjectilePayload(*ap)
					projectiles[ap.ID] = &payload
				}

				lines[al.ID] = &action.SyncStateLinePayload{
					ID:          al.ID,
					Towers:      towers,
					Units:       units,
					Projectiles: projectiles,
				}
			}

			aus := action.NewSyncState(
				&action.SyncStatePlayersPayload{
					Players:     players,
					IncomeTimer: r.Game.Store.Game.GetIncomeTimer(),
				},
				&action.SyncStateLinesPayload{
					Lines: lines,
				},
				r.StartedAt,
			)
			err := ac.ws.Write(context.Background(), pc.User.Conn, aus)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// syncLobbies will just sync the info of each lobby to the players on it
func (ac *ActionDispatcher) syncLobbies() {
	lbs := ac.store.Lobbies.List()
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
			u, ok := ac.store.Rooms.FindUserByUsername(p)
			if ok {
				err := ac.ws.Write(context.Background(), u.Conn, ula)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

// syncWaitingRooms updates the waiting room status
func (ac *ActionDispatcher) syncWaitingRooms() {
	for _, w := range ac.store.Rooms.ListWaitingRooms() {
		var players = make([]action.SyncWaitingRoomPlayersPayload, 0, 0)
		for pid, p := range w.Players {
			u, ok := ac.store.Rooms.FindUserByID(pid)
			if !ok {
				if !p.IsBot {
					continue
				}
				u.Username = pid
				u.ImageKey = unit.TypeStrings()[0]
			}
			_, ok = w.PlayersAccepted[u.Username]
			players = append(players, action.SyncWaitingRoomPlayersPayload{
				Username: u.Username,
				Accepted: ok,
				ImageKey: u.ImageKey,
			})
		}
		sort.Slice(players, func(i, j int) bool {
			return players[i].Username > players[j].Username
		})
		swra := action.NewSyncWaitingRoom(w.Size, w.Ranked, players, w.WaitingSince)
		for _, pc := range w.Players {
			// If is bot we skip it
			if pc.IsBot {
				continue
			}
			err := ac.ws.Write(context.Background(), pc.User.Conn, swra)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
