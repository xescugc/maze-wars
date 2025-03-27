package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/coder/websocket"
	"github.com/getsentry/sentry-go"
	"github.com/gofrs/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/server/bot"
	"github.com/xescugc/maze-wars/unit"
)

type RoomsStore struct {
	*flux.ReduceStore[RoomsState, *action.Action]

	Store *Store

	logger  *slog.Logger
	mxRooms sync.RWMutex

	ws WSConnector

	discord *discordgo.Session
	options Options
}

type RoomsState struct {
	// Searching are the current rooms that are searching for players
	// the key of the map is a combination of the vs+ranked values
	// so it can have 8 possible keys
	Searching map[string]*Room

	// Waiting is when a Searching Room it's ready to start we enter the Waiting
	// where all the participants have to accept the game
	Waiting map[string]*Room

	// Rooms is when an Waiting room is ready to start
	Rooms map[string]*Room

	Users map[string]*User
}

const (
	RoomTypeBots    = "bots"
	RoomTypeLobbies = "lobbies"
	RoomTypePlayers = "players"
)

type Room struct {
	Name string

	Type string

	Players map[string]PlayerConn

	Connections map[string]string

	// This is used so when the Room ends we
	// cancel any connected process to it
	Context         context.Context
	ContextCancelFn context.CancelFunc
	Bots            map[string]*bot.Bot

	Size      int
	Countdown int

	Ranked bool

	Game *Game

	SearchingSince time.Time
	WaitingSince   time.Time

	// PlayersAccepted is the list of players username that accepted the game
	PlayersAccepted map[string]struct{}

	// StartedAt is the time in which the Game started
	StartedAt time.Time

	// LFGMessageID is the ID of the message, used so we can delete it if the game starts
	LFGMessageID string
}

func (r Room) LFGMessage() string {
	rk := "Unranked"
	if r.Ranked {
		rk = "Ranked"
	}
	return fmt.Sprintf("LFG https://yawp-games.itch.io/maze-wars:\n* Size: %d\n* %s\n", r.Size, rk)
}

type User struct {
	ID       string
	Username string
	ImageKey string

	Conn       *websocket.Conn
	RemoteAddr string

	CurrentRoomID  string
	CurrentLobbyID string
}

type PlayerConn struct {
	User  *User
	IsBot bool
}

func NewRoomsStore(d *flux.Dispatcher[*action.Action], s *Store, ws WSConnector, dgo *discordgo.Session, opt Options, l *slog.Logger) *RoomsStore {
	rs := &RoomsStore{
		Store:   s,
		logger:  l,
		ws:      ws,
		discord: dgo,
		options: opt,
	}

	rs.ReduceStore = flux.NewReduceStore(d, rs.Reduce, RoomsState{
		Rooms:     make(map[string]*Room),
		Users:     make(map[string]*User),
		Searching: make(map[string]*Room),
		Waiting:   make(map[string]*Room),
	})

	return rs
}

func (rs *RoomsStore) ListRooms() []*Room {
	rs.mxRooms.RLock()
	defer rs.mxRooms.RUnlock()

	state := rs.GetState()
	rooms := make([]*Room, 0, len(state.Rooms))
	for _, r := range state.Rooms {
		rooms = append(rooms, r)
	}
	return rooms
}

func (rs *RoomsStore) ListSearchingRooms() []*Room {
	rs.mxRooms.RLock()
	defer rs.mxRooms.RUnlock()

	state := rs.GetState()
	rooms := make([]*Room, 0, len(state.Rooms))
	for _, r := range state.Searching {
		rooms = append(rooms, r)
	}
	return rooms
}

func (rs *RoomsStore) ListWaitingRooms() []*Room {
	rs.mxRooms.RLock()
	defer rs.mxRooms.RUnlock()

	state := rs.GetState()
	rooms := make([]*Room, 0, len(state.Rooms))
	for _, r := range state.Waiting {
		rooms = append(rooms, r)
	}
	return rooms
}

func (rs *RoomsStore) FindRoomByID(rid string) *Room {
	rs.mxRooms.RLock()
	defer rs.mxRooms.RUnlock()

	state := rs.GetState()
	r, ok := state.Rooms[rid]
	if !ok {
		return nil
	}
	return r
}

func (rs *RoomsStore) GetNextID(room string) int {
	r, _ := rs.GetState().Rooms[room]
	return len(r.Players)
}

func (rs *RoomsStore) FindUserByUsername(un string) (User, bool) {
	rs.mxRooms.RLock()
	defer rs.mxRooms.RUnlock()

	u, ok := rs.GetState().Users[un]
	if !ok {
		return User{}, false
	}
	return *u, true
}

func (rs *RoomsStore) FindUserByID(uid string) (User, bool) {
	rs.mxRooms.RLock()
	defer rs.mxRooms.RUnlock()

	for _, u := range rs.GetState().Users {
		if u.ID == uid {
			return *u, true
		}
	}
	return User{}, false
}

func (rs *RoomsStore) FindUserByRemoteAddress(ra string) (User, bool) {
	rs.mxRooms.RLock()
	defer rs.mxRooms.RUnlock()

	state := rs.GetState()

	return rs.findUserByRemoteAddress(state, ra)
}

func (rs *RoomsStore) findUserByRemoteAddress(state RoomsState, ra string) (User, bool) {
	for _, u := range state.Users {
		if u.RemoteAddr == ra {
			return *u, true
		}
	}
	return User{}, false
}

func (rs *RoomsStore) ListUsers() []*User {
	rs.mxRooms.RLock()
	defer rs.mxRooms.RUnlock()

	state := rs.GetState()
	users := make([]*User, 0, len(state.Users))
	for _, u := range state.Users {
		users = append(users, u)
	}
	return users
}

func (rs *RoomsStore) Reduce(state RoomsState, act *action.Action) RoomsState {
	switch act.Type {
	case action.UserSignUp:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		id := uuid.Must(uuid.NewV4())
		state.Users[act.UserSignUp.Username] = &User{
			ID:       id.String(),
			Username: act.UserSignUp.Username,
			ImageKey: act.UserSignUp.ImageKey,
		}

		currentNumberOfPlayers.Inc()
	case action.UserSignIn:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		if u, ok := state.Users[act.UserSignIn.Username]; ok {
			u.Conn = act.UserSignIn.Websocket
			u.RemoteAddr = act.UserSignIn.RemoteAddr
		}
	case action.UserSignOut:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		u, ok := state.Users[act.UserSignOut.Username]
		if ok && u.CurrentRoomID != "" {
			rs.removePlayer(&state, u.ID, u.CurrentRoomID)
		}

		delete(state.Users, act.UserSignOut.Username)
		currentNumberOfPlayers.Dec()
	case action.JoinLobby:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		if act.JoinLobby.IsBot {
			break
		}
		state.Users[act.JoinLobby.Username].CurrentLobbyID = act.JoinLobby.LobbyID
	case action.LeaveLobby:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		if un, ok := state.Users[act.LeaveLobby.Username]; ok {
			un.CurrentLobbyID = ""
		}
	case action.DeleteLobby:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		// TODO: Potentially make it better if this could have access to the
		// lobby and just target the users that we know need this.
		// It has access but it would need a WaitFor in order so synchronize
		for _, u := range state.Users {
			if u.CurrentLobbyID == act.DeleteLobby.LobbyID {
				u.CurrentLobbyID = ""
			}
		}
	case action.CreateLobby:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		state.Users[act.CreateLobby.Owner].CurrentLobbyID = act.CreateLobby.LobbyID
	case action.StartLobby:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		l := rs.Store.Lobbies.FindByID(act.StartLobby.LobbyID)
		r := &Room{
			Name:        l.ID,
			Type:        RoomTypeLobbies,
			Players:     make(map[string]PlayerConn),
			Connections: make(map[string]string),
			Bots:        make(map[string]*bot.Bot),

			Size:      l.MaxPlayers,
			Countdown: 10,
		}

		for p, ib := range l.Players {
			if ib {
				r.Players[p] = PlayerConn{
					IsBot: true,
				}
				continue
			}
			us, _ := state.Users[p]
			r.Players[us.ID] = PlayerConn{
				User: us,
			}
			r.Connections[us.RemoteAddr] = us.ID
		}

		state.Rooms[l.ID] = r

		rs.startRoom(state, l.ID)
	case action.FindGame:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		size := 2
		if !act.FindGame.Vs1 {
			size = 6
		}

		// If it has the VsBots true it'll never match as we'll never set it
		// so with VsBots we'll always create a new Room but before setting it
		// we'll do the logic
		sID := fmt.Sprintf("vs1=%t-ranked=%tbot=%t", act.FindGame.Vs1, act.FindGame.Ranked, act.FindGame.VsBots)

		ty := RoomTypePlayers
		if act.FindGame.VsBots {
			ty = RoomTypeBots
		}
		sr, ok := state.Searching[sID]
		if !ok {
			rid := uuid.Must(uuid.NewV4())
			sr = &Room{
				Name:        rid.String(),
				Type:        ty,
				Players:     make(map[string]PlayerConn),
				Connections: make(map[string]string),
				Bots:        make(map[string]*bot.Bot),

				Size: size,

				Ranked: act.FindGame.Ranked,

				SearchingSince: time.Now(),

				PlayersAccepted: make(map[string]struct{}),
			}

			if Environment == "prod" {
				msg, err := rs.discord.ChannelMessageSend(rs.options.DiscordChannelID, sr.LFGMessage())
				if err != nil {
					rs.logger.Error("Failed to send message to Discord", "error", err.Error())
				} else {
					sr.LFGMessageID = msg.ID
				}
			}
		}

		us, _ := state.Users[act.FindGame.Username]
		sr.Players[us.ID] = PlayerConn{
			User: us,
		}
		sr.Connections[us.RemoteAddr] = us.ID

		// If the number of players is reached then we move the Room
		// to the waiting state
		if len(sr.Players) == sr.Size {
			sr.SearchingSince = time.Time{}
			sr.WaitingSince = time.Now()
			state.Waiting[sr.Name] = sr

			delete(state.Searching, sID)
		} else {
			if act.FindGame.VsBots {
				// If we play against bots we'll add the Bots, and make them all Accept
				// so the player is prompted with the accept modal
				sr.SearchingSince = time.Time{}
				sr.WaitingSince = time.Now()
				for i := range make([]int, sr.Size-1) {
					bn := fmt.Sprintf("Bot-%d", i+1)
					sr.Players[bn] = PlayerConn{
						IsBot: true,
					}
					sr.PlayersAccepted[bn] = struct{}{}
				}
				state.Waiting[sr.Name] = sr
			} else {
				state.Searching[sID] = sr
			}
		}
	case action.ExitSearchingGame:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		// TODO: Potentially add this to the RemovePlayer
		us, _ := state.Users[act.ExitSearchingGame.Username]
		for sID, sr := range state.Searching {
			if _, ok := sr.Players[us.ID]; ok {
				delete(sr.Players, us.ID)
				delete(sr.Connections, us.RemoteAddr)
				if len(sr.Players) == 0 {
					delete(state.Searching, sID)
					if Environment == "prod" {
						err := rs.discord.ChannelMessageDelete(rs.options.DiscordChannelID, sr.LFGMessageID)
						if err != nil {
							rs.logger.Error("Failed to delete message from Discord", "error", err.Error())
						}
					}
				}
			}
		}
	case action.AcceptWaitingGame:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		us, _ := state.Users[act.AcceptWaitingGame.Username]
		for wrID, wr := range state.Waiting {
			if _, ok := wr.Players[us.ID]; ok {
				wr.PlayersAccepted[us.Username] = struct{}{}
			}
			if len(wr.PlayersAccepted) == len(wr.Players) {
				state.Rooms[wrID] = wr
				wr.WaitingSince = time.Time{}
				wr.PlayersAccepted = nil

				delete(state.Waiting, wrID)
				if Environment == "prod" {
					err := rs.discord.ChannelMessageDelete(rs.options.DiscordChannelID, wr.LFGMessageID)
					if err != nil {
						rs.logger.Error("Failed to delete message from Discord", "error", err.Error())
					}
				}
				rs.startRoom(state, wrID)
			}
		}
	case action.CancelWaitingGame:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		us, _ := state.Users[act.CancelWaitingGame.Username]
		for wrID, wr := range state.Waiting {
			if _, ok := wr.Players[us.ID]; ok {

				for _, pc := range wr.Players {
					if pc.IsBot {
						continue
					}
					err := rs.ws.Write(context.Background(), pc.User.Conn, act)
					if err != nil {
						log.Fatal(err)
					}
				}
				delete(state.Waiting, wrID)
				if Environment == "prod" {
					err := rs.discord.ChannelMessageDelete(rs.options.DiscordChannelID, wr.LFGMessageID)
					if err != nil {
						rs.logger.Error("Failed to delete message from Discord", "error", err.Error())
					}
				}
			}
		}

	case action.RemovePlayer:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		for _, u := range state.Users {
			if u.ID == act.RemovePlayer.ID {
				u.CurrentRoomID = ""
				break
			}
		}

		rs.removePlayer(&state, act.RemovePlayer.ID, act.Room)
	default:
		rs.mxRooms.Lock()
		defer rs.mxRooms.Unlock()

		// If no room means that is a broadcast
		if act.Room == "" {
			for _, r := range state.Rooms {
				if r.Game == nil {
					return state
				}
				go func() {
					localHub := sentry.CurrentHub().Clone()
					defer func() {
						err := recover()

						if err != nil {
							localHub.Recover(err)
							localHub.Flush(time.Second * 5)
							if Environment == "dev" {
								panic(err)
							}
						}
					}()

					r.Game.Dispatch(act)
				}()
			}
		} else {
			if r, ok := state.Rooms[act.Room]; ok {
				go func() {
					localHub := sentry.CurrentHub().Clone()
					defer func() {
						err := recover()

						if err != nil {
							localHub.Recover(err)
							localHub.Flush(time.Second * 5)
							if Environment == "dev" {
								panic(err)
							}
						}
					}()

					r.Game.Dispatch(act)
				}()
			}
		}
	}

	return state
}

func (rs *RoomsStore) removePlayer(state *RoomsState, pid, room string) {
	if room != "" {
		pc := state.Rooms[room].Players[pid]
		delete(state.Rooms[room].Players, pid)
		delete(state.Rooms[room].Connections, pc.User.RemoteAddr)

		state.Rooms[room].Game.Dispatch(action.NewRemovePlayer(pid))

		if pc.IsBot {
			state.Rooms[room].Bots[pid].Stop()
			delete(state.Rooms[room].Bots, pid)
		}

		r := state.Rooms[room]
		if len(r.Players) == 0 {
			delete(state.Rooms, room)
			currentNumberOfGames.Dec()
			if Environment == "prod" {
				err := rs.discord.ChannelMessageDelete(rs.options.DiscordChannelID, r.LFGMessageID)
				if err != nil {
					rs.logger.Error("Failed to delete message from Discord", "error", err.Error())
				}
			}
			return
		}
		var humanFound bool
		for _, pc := range r.Players {
			if !pc.IsBot {
				humanFound = true
				break
			}
		}
		// If no human was found left alive we just remove the room
		if !humanFound {
			for _, b := range r.Bots {
				b.Stop()
			}
			delete(state.Rooms, room)
			currentNumberOfGames.Dec()
			if Environment == "prod" {
				err := rs.discord.ChannelMessageDelete(rs.options.DiscordChannelID, r.LFGMessageID)
				if err != nil {
					rs.logger.Error("Failed to delete message from Discord", "error", err.Error())
				}
			}
		}
	} else {
		// TODO: Fix this
		// Search for the Waiting rooms
		//for sID, sr := range rs.Waiting {
		//if _, ok := sr.Players[pid]; ok {
		//delete(sr.Players, pid)
		//delete(sr.Connections, pc.RemoteAddr)
		//if len(sr.Players) == 0 {
		//delete(rs.Waiting, sID)
		//}
		//}
		//}
	}
}

func (rs RoomsStore) startRoom(state RoomsState, rid string) {
	rd := flux.NewDispatcher[*action.Action]()
	g := NewGame(rd, rs.logger)
	cr := state.Rooms[rid]
	cr.StartedAt = time.Now()
	ctx := context.Background()
	cr.Context, cr.ContextCancelFn = context.WithCancel(ctx)
	cr.Game = g
	pcount := 0
	for pid, pc := range cr.Players {
		if pc.IsBot {
			g.Dispatch(action.NewAddPlayer(pid, pid, unit.TypeStrings()[0], pcount, pc.IsBot))
			cr.Bots[pid] = bot.New(cr.Context, rd, g.Store, pid)
		} else {
			//u, _ := rs.findUserByRemoteAddress(state, pc.RemoteAddr)
			//uu := state.Users[u.Username]
			//uu.CurrentRoomID = rid
			pc.User.CurrentRoomID = rid
			g.Dispatch(action.NewAddPlayer(pid, pc.User.Username, pc.User.ImageKey, pcount, pc.IsBot))
		}
		pcount++
	}
	ssp := rs.SyncState(cr, "")
	sga := action.NewStartGame(ssp)

	g.Dispatch(sga)
	for pid, pc := range cr.Players {
		if pc.IsBot {
			cr.Bots[pid].Start()
			continue
		} else {
			ssp := rs.SyncState(cr, pc.User.ID)
			sga := action.NewStartGame(ssp)
			err := rs.ws.Write(context.Background(), pc.User.Conn, sga)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	currentNumberOfGames.Inc()
	numberOfGames.With(prometheus.Labels{"type": cr.Type}).Inc()
}

func (rs RoomsStore) SyncState(r *Room, pid string) action.SyncStatePayload {
	// Players
	players := make(map[string]*action.SyncStatePlayerPayload)
	lstate := r.Game.Store.Game.GetState()
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
			Current:     ap.Current,
			Winner:      ap.Winner,
			UnitUpdates: make(map[string]action.SyncStatePlayerUnitUpdatePayload),
		}
		// TODO: Make it concurrently safe
		for t, uu := range ap.UnitUpdates {
			uspp.UnitUpdates[t] = action.SyncStatePlayerUnitUpdatePayload(uu)
		}
		if pid == ap.ID {
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
			ustp := action.SyncStateTowerPayload(*at)
			towers[at.ID] = &ustp
		}

		// Units
		units := make(map[string]*action.SyncStateUnitPayload)
		for _, u := range al.Units {
			au := u
			usup := action.SyncStateUnitPayload(*au)
			units[au.ID] = &usup
		}
		lines[al.ID] = &action.SyncStateLinePayload{
			ID:     al.ID,
			Towers: towers,
			Units:  units,
		}
	}

	return action.SyncStatePayload{
		&action.SyncStatePlayersPayload{
			Players:     players,
			IncomeTimer: r.Game.Store.Game.GetIncomeTimer(),
		},
		&action.SyncStateLinesPayload{
			Lines: lines,
		},
		r.StartedAt,
		lstate.Error,
		lstate.ErrorAt,
	}
}
