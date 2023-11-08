package main

import (
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
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

// CursorMove dispatches an action of moving the Cursor
// to the new x,y coordinates
func (ac *ActionDispatcher) CursorMove(x, y int) {
	cma := action.NewCursorMove(x, y)
	ac.dispatcher.Dispatch(cma)
}

// SummonUnit summons the 'unit' from the player id 'pid' to the line
// 'plid' and with the current line id 'clid'
func (ac *ActionDispatcher) SummonUnit(unit, pid string, plid, clid int) {
	sua := action.NewSummonUnit(unit, pid, plid, clid)
	wsSend(sua)
	//ac.dispatcher.Dispatch(sua)
}

// MoveUnit moves all the units
func (ac *ActionDispatcher) MoveUnit() {
	mua := action.NewMoveUnit()
	ac.dispatcher.Dispatch(mua)
}

// RemoveUnit removes the unit with the id 'uid'
func (ac *ActionDispatcher) RemoveUnit(uid string) {
	rua := action.NewRemoveUnit(uid)
	wsSend(rua)
	ac.dispatcher.Dispatch(rua)
}

// StealLive removes one live from the player with id 'fpid' and
// adds it to the player with id 'tpid'
func (ac *ActionDispatcher) StealLive(fpid, tpid string) {
	sla := action.NewStealLive(fpid, tpid)
	wsSend(sla)
	ac.dispatcher.Dispatch(sla)
}

// CameraZoom zooms the camera the direction 'd'
func (ac *ActionDispatcher) CameraZoom(d int) {
	cza := action.NewCameraZoom(d)
	ac.dispatcher.Dispatch(cza)
}

// PlaceTower places the tower 't' on the position X and Y of the player pid
func (ac *ActionDispatcher) PlaceTower(t, pid string, x, y int) {
	pta := action.NewPlaceTower(t, pid, x, y)
	wsSend(pta)
	ac.dispatcher.Dispatch(pta)
}

// RemoveTower removes the tower tid
func (ac *ActionDispatcher) RemoveTower(pid, tid, tt string) {
	rta := action.NewRemoveTower(pid, tid, tt)
	wsSend(rta)
	ac.dispatcher.Dispatch(rta)
}

// SelectTower selects the tower 't' on the position x, y
func (ac *ActionDispatcher) SelectTower(t string, x, y int) {
	sta := action.NewSelectTower(t, x, y)
	ac.dispatcher.Dispatch(sta)
}

// SelectTower selects the tower 't' on the position x, y
func (ac *ActionDispatcher) SelectedTowerInvalid(i bool) {
	sta := action.NewSelectedTowerInvalid(i)
	ac.dispatcher.Dispatch(sta)
}

// DeelectTower cleans the current selected tower
func (ac *ActionDispatcher) DeselectTower(t string) {
	dsta := action.NewDeselectTower(t)
	ac.dispatcher.Dispatch(dsta)
}

// IncomeTick a new tick for the income
func (ac *ActionDispatcher) IncomeTick() {
	it := action.NewIncomeTick()
	ac.dispatcher.Dispatch(it)
}

// TowerAttack issues a attack to the Unit with uid
func (ac *ActionDispatcher) TowerAttack(uid, tt string) {
	ta := action.NewTowerAttack(uid, tt)
	wsSend(ta)
	ac.dispatcher.Dispatch(ta)
}

// UnitKilled adds gold to the user
func (ac *ActionDispatcher) UnitKilled(pid, ut string) {
	uk := action.NewUnitKilled(pid, ut)
	wsSend(uk)
	ac.dispatcher.Dispatch(uk)
}

// PlayerReady marks the player as ready to start the game
func (ac *ActionDispatcher) PlayerReady(pid string) {
	pr := action.NewPlayerReady(pid)
	wsSend(pr)
	ac.dispatcher.Dispatch(pr)
}

// WindowResizing new sizes of the window
func (ac *ActionDispatcher) WindowResizing(w, h int) {
	wr := action.NewWindowResizing(w, h)
	ac.dispatcher.Dispatch(wr)
}

// JoinRoom new sizes of the window
func (ac *ActionDispatcher) JoinRoom(room, name string) {
	jr := action.NewJoinRoom(room, name)
	wsSend(jr)
	ac.dispatcher.Dispatch(jr)
}

// NavigateTo navigates to the given route
func (ac *ActionDispatcher) NavigateTo(route string) {
	nt := action.NewNavigateTo(route)
	ac.dispatcher.Dispatch(nt)
}

// StartGame notifies that the game will start,
// used to update any store before that
func (ac *ActionDispatcher) StartGame() {
	sg := action.NewStartGame()
	ac.dispatcher.Dispatch(sg)
}

// OpenTowerMenu when a tower is clicked and the menu of
// the tower is displayed
func (ac *ActionDispatcher) OpenTowerMenu(tid string) {
	otm := action.NewOpenTowerMenu(tid)
	ac.dispatcher.Dispatch(otm)
}

// CloseTowerMenu when a tower menu needs to be closed
func (ac *ActionDispatcher) CloseTowerMenu() {
	ctm := action.NewCloseTowerMenu()
	ac.dispatcher.Dispatch(ctm)
}

// GoHome will move the camera to the current player home line
func (ac *ActionDispatcher) GoHome() {
	gha := action.NewGoHome()
	ac.dispatcher.Dispatch(gha)
}

// CheckedPath will set the value of the path checked
func (ac *ActionDispatcher) CheckedPath(cp bool) {
	cpa := action.NewCheckedPath(cp)
	ac.dispatcher.Dispatch(cpa)
}
