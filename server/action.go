package server

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
	"github.com/xescugc/ltw/store"
)

// ActionDispatcher is in charge of dispatching actions to the
// application dispatcher
type ActionDispatcher struct {
	dispatcher *flux.Dispatcher
}

// NewActionDispatcher initializes the action dispatcher
// with the give dispatcher
func NewActionDispatcher(d *flux.Dispatcher) *ActionDispatcher {
	return &ActionDispatcher{
		dispatcher: d,
	}
}

// Dispatch is a helper to access to the internal dispatch directly with an action.
// This should only be used from the WS Handler to forward server actions directly
func (ac *ActionDispatcher) Dispatch(a *action.Action) {
	ac.dispatcher.Dispatch(a)
}

func (ac *ActionDispatcher) AddPlayer(sid, id, name string, lid int, ws *websocket.Conn) {
	npa := action.NewAddPlayer(sid, id, name, lid, ws)
	ac.dispatcher.Dispatch(npa)
}

func (ac *ActionDispatcher) RemovePlayer(rn, sid string) {
	rpa := action.NewRemovePlayer(rn, sid)
	rpa.Room = rn
	ac.dispatcher.Dispatch(rpa)
}

func (ac *ActionDispatcher) IncomeTick(rooms *RoomsStore) {
	ita := action.NewIncomeTick()
	ac.dispatcher.Dispatch(ita)
}

func (ac *ActionDispatcher) MoveUnit(rooms *RoomsStore) {
	mua := action.NewMoveUnit()
	ac.dispatcher.Dispatch(mua)
}

func (ac *ActionDispatcher) UpdateState(rooms *RoomsStore) {
	for _, r := range rooms.GetState().(RoomsState).Rooms {
		for id, con := range r.Players {
			// Players
			players := make(map[string]*action.UpdateStatePlayerPayload)
			ps := r.Game.Players.GetState().(store.PlayersState)
			for idp, p := range ps.Players {
				uspp := action.UpdateStatePlayerPayload(*p)
				if id == idp {
					uspp.Current = true
				}
				players[idp] = &uspp
			}

			// Towers
			towers := make(map[string]*action.UpdateStateTowerPayload)
			ts := r.Game.Towers.GetTowers()
			for _, t := range ts {
				ustp := action.UpdateStateTowerPayload(*t)
				towers[t.ID] = &ustp
			}

			// Units
			units := make(map[string]*action.UpdateStateUnitPayload)
			us := r.Game.Units.GetUnits()
			for _, u := range us {
				usup := action.UpdateStateUnitPayload(*u)
				units[u.ID] = &usup
			}

			aus := action.NewUpdateState(
				&action.UpdateStatePlayersPayload{
					Players:     players,
					IncomeTimer: ps.IncomeTimer,
				},
				&action.UpdateStateTowersPayload{
					Towers: towers,
				},
				&action.UpdateStateUnitsPayload{
					Units: units,
				},
			)
			err := con.WriteJSON(aus)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
