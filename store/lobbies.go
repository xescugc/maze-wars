package store

import (
	"sort"
	"sync"

	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	cutils "github.com/xescugc/maze-wars/client/utils"
)

type Lobbies struct {
	*flux.ReduceStore

	mxLobbies sync.RWMutex
}

type LobbiesState struct {
	Lobbies map[string]*Lobby
}

type Lobby struct {
	ID         string
	Name       string
	MaxPlayers int
	// Players holds the usernames
	// including the owner one
	Players map[string]struct{}
	// The username of the owner
	Owner   string
	Current bool
}

func (l Lobby) PlayersSlice() []string {
	pls := make([]string, 0, len(l.Players))
	for p := range l.Players {
		pls = append(pls, p)
	}
	sort.Strings(pls)
	return pls
}

func NewLobbies(d *flux.Dispatcher) *Lobbies {
	l := &Lobbies{}
	l.ReduceStore = flux.NewReduceStore(d, l.Reduce, LobbiesState{
		Lobbies: make(map[string]*Lobby),
	})

	return l
}

func (ls *Lobbies) List() []*Lobby {
	ls.mxLobbies.RLock()
	defer ls.mxLobbies.RUnlock()

	slobbies := ls.GetState().(LobbiesState)
	lobbies := make([]*Lobby, 0, len(slobbies.Lobbies))
	for _, l := range slobbies.Lobbies {
		lobbies = append(lobbies, l)
	}
	return lobbies
}

func (ls *Lobbies) FindCurrent() *Lobby {
	ls.mxLobbies.RLock()
	defer ls.mxLobbies.RUnlock()

	slobbies := ls.GetState().(LobbiesState)
	for _, l := range slobbies.Lobbies {
		if l.Current {
			return l
		}
	}
	return nil
}

func (ls *Lobbies) FindByID(id string) *Lobby {
	ls.mxLobbies.RLock()
	defer ls.mxLobbies.RUnlock()

	slobbies := ls.GetState().(LobbiesState)
	l, _ := slobbies.Lobbies[id]
	return l
}

func (ls *Lobbies) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	lstate, ok := state.(LobbiesState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.CreateLobby:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		lstate.Lobbies[act.CreateLobby.LobbyID] = &Lobby{
			ID:         act.CreateLobby.LobbyID,
			Name:       act.CreateLobby.LobbyName,
			MaxPlayers: act.CreateLobby.LobbyMaxPlayers,
			Owner:      act.CreateLobby.Owner,
			Players:    map[string]struct{}{act.CreateLobby.Owner: struct{}{}},
		}
	case action.DeleteLobby:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		delete(lstate.Lobbies, act.DeleteLobby.LobbyID)
	case action.JoinLobby:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		lstate.Lobbies[act.JoinLobby.LobbyID].Players[act.JoinLobby.Username] = struct{}{}
	case action.LeaveLobby:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		delete(lstate.Lobbies[act.LeaveLobby.LobbyID].Players, act.LeaveLobby.Username)
	case action.AddLobbies:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		clbs := make(map[string]struct{})
		for id := range lstate.Lobbies {
			clbs[id] = struct{}{}
		}
		for _, al := range act.AddLobbies.Lobbies {
			l := &Lobby{
				ID:         al.ID,
				Name:       al.Name,
				MaxPlayers: al.MaxPlayers,
				Owner:      al.Owner,
				Players:    make(map[string]struct{}),
			}

			for _, p := range al.Players {
				l.Players[p] = struct{}{}
			}
			lstate.Lobbies[l.ID] = l
			delete(clbs, al.ID)
		}
		for id := range clbs {
			delete(lstate.Lobbies, id)
		}
	case action.NavigateTo:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		// If we navigate out of the Show page we should
		// unset the current
		if act.NavigateTo.Route != cutils.ShowLobbyRoute {
			for _, l := range lstate.Lobbies {
				l.Current = false
			}
		}
	case action.SelectLobby:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		lstate.Lobbies[act.SelectLobby.LobbyID].Current = true
	case action.UpdateLobby:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		ulp := act.UpdateLobby.Lobby

		l := lstate.Lobbies[ulp.ID]

		ul := &Lobby{
			ID:         ulp.ID,
			Name:       ulp.Name,
			MaxPlayers: ulp.MaxPlayers,
			Players:    make(map[string]struct{}),
			Owner:      ulp.Owner,
			Current:    l.Current,
		}
		for _, p := range ulp.Players {
			ul.Players[p] = struct{}{}
		}
		lstate.Lobbies[ul.ID] = ul
	case action.UserSignOut:
		ls.mxLobbies.Lock()
		defer ls.mxLobbies.Unlock()

		for _, l := range lstate.Lobbies {
			if l.Owner == act.UserSignOut.Username {
				delete(lstate.Lobbies, l.ID)
			} else {
				delete(l.Players, act.UserSignOut.Username)
			}
		}
	}

	return state
}
