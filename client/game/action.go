package game

import (
	"log/slog"
	"time"

	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	cutils "github.com/xescugc/maze-wars/client/utils"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/utils"
)

// ActionDispatcher is in charge of dispatching actions to the
// application dispatcher
type ActionDispatcher struct {
	dispatcher *flux.Dispatcher
	store      *store.Store
	logger     *slog.Logger

	wsSend func(a *action.Action)
}

// NewActionDispatcher initializes the action dispatcher
// with the give dispatcher
func NewActionDispatcher(d *flux.Dispatcher, s *store.Store, wsSendFn func(a *action.Action), l *slog.Logger) *ActionDispatcher {
	return &ActionDispatcher{
		dispatcher: d,
		store:      s,
		logger:     l,
		wsSend:     wsSendFn,
	}
}

// Dispatch is a helper to access to the internal dispatch directly with an action.
// This should only be used from the WS Handler to forward server actions directly
func (ac *ActionDispatcher) Dispatch(a *action.Action) {
	b := time.Now()
	defer utils.LogTime(ac.logger, b, "action dispatched", "action", a.Type)

	ac.dispatcher.Dispatch(a)
}

// CursorMove dispatches an action of moving the Cursor
// to the new x,y coordinates
func (ac *ActionDispatcher) CursorMove(x, y int) {
	cma := action.NewCursorMove(x, y)
	ac.Dispatch(cma)
}

// SummonUnit summons the 'unit' from the player id 'pid' to the line
// 'plid' and with the current line id 'clid'
func (ac *ActionDispatcher) SummonUnit(unit, pid string, plid, clid int) {
	sua := action.NewSummonUnit(unit, pid, plid, clid)
	ac.wsSend(sua)
	//ac.Dispatch(sua)
}

// TPS is the call for every TPS event
func (ac *ActionDispatcher) TPS() {
	tpsa := action.NewTPS(time.Time{})
	ac.Dispatch(tpsa)
}

// RemoveUnit removes the unit with the id 'uid'
func (ac *ActionDispatcher) RemoveUnit(uid string) {
	rua := action.NewRemoveUnit(uid)
	ac.wsSend(rua)
	ac.Dispatch(rua)
}

// StealLive removes one live from the player with id 'fpid' and
// adds it to the player with id 'tpid'
func (ac *ActionDispatcher) StealLive(fpid, tpid string) {
	sla := action.NewStealLive(fpid, tpid)
	ac.wsSend(sla)
	ac.Dispatch(sla)
}

// CameraZoom zooms the camera the direction 'd'
func (ac *ActionDispatcher) CameraZoom(d int) {
	cza := action.NewCameraZoom(d)
	ac.Dispatch(cza)
}

// PlaceTower places the tower 't' on the position X and Y of the player pid
func (ac *ActionDispatcher) PlaceTower(t, pid string, x, y int) {
	// TODO: Add the LineID in the action
	p := ac.store.Players.FindByID(pid)
	if l := ac.store.Lines.FindByID(p.LineID); l != nil && l.Graph.CanAddTower(x, y, 32, 32) {
		pta := action.NewPlaceTower(t, pid, x, y)
		ac.wsSend(pta)
	}
	//ac.Dispatch(pta)
}

// RemoveTower removes the tower tid
func (ac *ActionDispatcher) RemoveTower(pid, tid, tt string) {
	rta := action.NewRemoveTower(pid, tid, tt)
	ac.wsSend(rta)
	//ac.Dispatch(rta)
}

// SelectTower selects the tower 't' on the position x, y
func (ac *ActionDispatcher) SelectTower(t string, x, y int) {
	sta := action.NewSelectTower(t, x, y)
	ac.Dispatch(sta)
}

// SelectTower selects the tower 't' on the position x, y
func (ac *ActionDispatcher) SelectedTowerInvalid(i bool) {
	sta := action.NewSelectedTowerInvalid(i)
	ac.Dispatch(sta)
}

// DeselectTower cleans the current selected tower
func (ac *ActionDispatcher) DeselectTower(t string) {
	dsta := action.NewDeselectTower(t)
	ac.Dispatch(dsta)
}

// IncomeTick a new tick for the income
func (ac *ActionDispatcher) IncomeTick() {
	it := action.NewIncomeTick()
	ac.Dispatch(it)
}

// TowerAttack issues a attack to the Unit with uid
func (ac *ActionDispatcher) TowerAttack(uid, tt string) {
	ta := action.NewTowerAttack(uid, tt)
	ac.wsSend(ta)
	ac.Dispatch(ta)
}

// UnitKilled adds gold to the user
func (ac *ActionDispatcher) UnitKilled(pid, ut string) {
	uk := action.NewUnitKilled(pid, ut)
	ac.wsSend(uk)
	ac.Dispatch(uk)
}

func (ac *ActionDispatcher) RemovePlayer(pid string) {
	rpa := action.NewRemovePlayer(pid)
	ac.wsSend(rpa)
	ac.Dispatch(rpa)
	ac.Dispatch(action.NewNavigateTo(cutils.LobbyRoute))
}

// OpenTowerMenu when a tower is clicked and the menu of
// the tower is displayed
func (ac *ActionDispatcher) OpenTowerMenu(tid string) {
	otm := action.NewOpenTowerMenu(tid)
	ac.Dispatch(otm)
}

// CloseTowerMenu when a tower menu needs to be closed
func (ac *ActionDispatcher) CloseTowerMenu() {
	ctm := action.NewCloseTowerMenu()
	ac.Dispatch(ctm)
}

// ChangeUnitLine will move the unit to the next line
func (ac *ActionDispatcher) ChangeUnitLine(uid string) {
	cula := action.NewChangeUnitLine(uid)
	ac.wsSend(cula)
	ac.Dispatch(cula)
}

func (ac *ActionDispatcher) ToggleStats() {
	tsa := action.NewToggleStats()

	ac.Dispatch(tsa)
}

// GoHome will move the camera to the current player home line
func (ac *ActionDispatcher) GoHome() {
	gha := action.NewGoHome()
	ac.Dispatch(gha)
}
