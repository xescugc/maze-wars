package game

import (
	"log/slog"
	"time"

	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/utils"
)

// ActionDispatcher is in charge of dispatching actions to the
// application dispatcher
type ActionDispatcher struct {
	dispatcher *flux.Dispatcher[*action.Action]
	store      *store.Store
	logger     *slog.Logger

	wsSend func(a *action.Action)
}

// NewActionDispatcher initializes the action dispatcher
// with the give dispatcher
func NewActionDispatcher(d *flux.Dispatcher[*action.Action], s *store.Store, wsSendFn func(a *action.Action), l *slog.Logger) *ActionDispatcher {
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
func (ac *ActionDispatcher) CursorMove(x, y int, imp bool) {
	cma := action.NewCursorMove(x, y, imp)
	ac.Dispatch(cma)
}

// SummonUnit summons the 'unit' from the player id 'pid' to the line
// 'plid' and with the current line id 'clid'
func (ac *ActionDispatcher) SummonUnit(unit, pid string, plid, clid int) {
	sua := action.NewSummonUnit(unit, pid, plid, clid)
	ac.wsSend(sua)
	//ac.Dispatch(sua)
}

func (ac *ActionDispatcher) UpdateUnit(pid, ut string) {
	uua := action.NewUpdateUnit(pid, ut)
	ac.wsSend(uua)
	ac.Dispatch(uua)
}

// TPS is the call for every TPS event
func (ac *ActionDispatcher) TPS() {
	tpsa := action.NewTPS(time.Time{})
	ac.Dispatch(tpsa)
}

// CameraZoom zooms the camera the direction 'd'
func (ac *ActionDispatcher) CameraZoom(d int) {
	cza := action.NewCameraZoom(d)
	ac.Dispatch(cza)
}

// PlaceTower places the tower 't' on the position X and Y of the player pid
func (ac *ActionDispatcher) PlaceTower(t, pid string, x, y int) {
	// TODO: Add the LineID in the action
	p := ac.store.Game.FindPlayerByID(pid)
	if l := ac.store.Game.FindLineByID(p.LineID); l != nil && l.Graph.CanAddTower(x, y, 32, 32) {
		pta := action.NewPlaceTower(t, pid, x, y)
		ac.wsSend(pta)
	}
	//ac.Dispatch(pta)
}

// RemoveTower removes the tower tid
func (ac *ActionDispatcher) RemoveTower(pid, tid string) {
	rta := action.NewRemoveTower(pid, tid)
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

func (ac *ActionDispatcher) RemovePlayer(pid string) {
	rpa := action.NewRemovePlayer(pid)
	ac.wsSend(rpa)
	ac.Dispatch(rpa)
	ac.Dispatch(action.NewNavigateTo(utils.RootRoute))
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

// OpenUnitMenu when a unit is clicked and the menu of
// the unit is displayed
func (ac *ActionDispatcher) OpenUnitMenu(uid string) {
	oum := action.NewOpenUnitMenu(uid)
	ac.Dispatch(oum)
}

// CloseUnitMenu when a tower menu needs to be closed
func (ac *ActionDispatcher) CloseUnitMenu() {
	cum := action.NewCloseUnitMenu()
	ac.Dispatch(cum)
}

//func (ac *ActionDispatcher) ToggleStats() {
//tsa := action.NewToggleStats()

//ac.Dispatch(tsa)
//}

// GoHome will move the camera to the current player home line
func (ac *ActionDispatcher) GoHome() {
	gha := action.NewGoHome()
	ac.Dispatch(gha)
}

// GoHome will move the camera to the current player home line
func (ac *ActionDispatcher) UpdateTower(pid, tid, tt string) {
	uta := action.NewUpdateTower(pid, tid, tt)
	ac.wsSend(uta)
	ac.Dispatch(uta)
}

// GoHome will move the camera to the current player home line
func (ac *ActionDispatcher) ShowScoreboard(d bool) {
	ssa := action.NewShowScoreboard(d)
	ac.Dispatch(ssa)
}

// GoHome will move the camera to the current player home line
func (ac *ActionDispatcher) AddError(err string) {
	ssa := action.NewAddError(err)
	ac.Dispatch(ssa)
}
