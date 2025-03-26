package store

import (
	"sync"

	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/maze-wars/action"
)

type Lobbies struct {
	*flux.ReduceStore[LobbiesState, *action.Action]

	mxLobbies sync.RWMutex
}

type LobbiesState struct {
	Lobbies map[string]*Lobby
	// Marks if the Lobbies have been Seen
	// it's used to render or not the values
	Seen bool
}

type Lobby struct {
	ID         string
	Name       string
	MaxPlayers int
	// Players holds the usernames
	// including the owner one.
	// And the bool value represents
	// if it's a "bot" or not
	Players map[string]bool
	// The username of the owner
	Owner   string
	Current bool
}

func NewLobbies(d *flux.Dispatcher[*action.Action]) *Lobbies {
	l := &Lobbies{}
	l.ReduceStore = flux.NewReduceStore(d, l.Reduce, LobbiesState{
		Lobbies: make(map[string]*Lobby),
	})

	return l
}

func (ls *Lobbies) List() []*Lobby {
	ls.mxLobbies.RLock()
	defer ls.mxLobbies.RUnlock()

	state := ls.GetState()
	lobbies := make([]*Lobby, 0, len(state.Lobbies))
	for _, l := range state.Lobbies {
		lobbies = append(lobbies, l)
	}
	return lobbies
}

func (ls *Lobbies) Seen() bool {
	ls.mxLobbies.RLock()
	defer ls.mxLobbies.RUnlock()

	state := ls.GetState()
	return state.Seen
}

func (ls *Lobbies) FindCurrent() *Lobby {
	ls.mxLobbies.RLock()
	defer ls.mxLobbies.RUnlock()

	state := ls.GetState()
	for _, l := range state.Lobbies {
		if l.Current {
			return l
		}
	}
	return nil
}

func (ls *Lobbies) FindByID(id string) *Lobby {
	ls.mxLobbies.RLock()
	defer ls.mxLobbies.RUnlock()

	state := ls.GetState()
	l, _ := state.Lobbies[id]
	return l
}

func (ls *Lobbies) Reduce(state LobbiesState, act *action.Action) LobbiesState {
	switch act.Type {
	case action.CreateLobby:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		state.Lobbies[act.CreateLobby.LobbyID] = &Lobby{
			ID:         act.CreateLobby.LobbyID,
			Name:       act.CreateLobby.LobbyName,
			MaxPlayers: act.CreateLobby.LobbyMaxPlayers,
			Owner:      act.CreateLobby.Owner,
			Players:    map[string]bool{act.CreateLobby.Owner: false},
		}
	case action.DeleteLobby:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		delete(state.Lobbies, act.DeleteLobby.LobbyID)
	case action.JoinLobby:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		state.Lobbies[act.JoinLobby.LobbyID].Players[act.JoinLobby.Username] = act.JoinLobby.IsBot
	case action.LeaveLobby:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		delete(state.Lobbies[act.LeaveLobby.LobbyID].Players, act.LeaveLobby.Username)
	case action.AddLobbies:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		clbs := make(map[string]struct{})
		for id := range state.Lobbies {
			clbs[id] = struct{}{}
		}
		for _, al := range act.AddLobbies.Lobbies {
			l := &Lobby{
				ID:         al.ID,
				Name:       al.Name,
				MaxPlayers: al.MaxPlayers,
				Owner:      al.Owner,
				Players:    al.Players,
			}

			state.Lobbies[l.ID] = l
			delete(clbs, al.ID)
		}
		for id := range clbs {
			delete(state.Lobbies, id)
		}
		state.Seen = false
	case action.SeenLobbies:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		state.Seen = true
	case action.SelectLobby:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		state.Lobbies[act.SelectLobby.LobbyID].Current = true
	case action.UpdateLobby:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		ulp := act.UpdateLobby.Lobby

		l := state.Lobbies[ulp.ID]

		ul := &Lobby{
			ID:         ulp.ID,
			Name:       ulp.Name,
			MaxPlayers: ulp.MaxPlayers,
			Players:    ulp.Players,
			Owner:      ulp.Owner,
			Current:    l.Current,
		}
		state.Lobbies[ul.ID] = ul
	case action.UserSignOut:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		for _, l := range state.Lobbies {
			if l.Owner == act.UserSignOut.Username {
				delete(state.Lobbies, l.ID)
			} else {
				delete(l.Players, act.UserSignOut.Username)
			}
		}
	}

	return state
}
