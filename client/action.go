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

// CursorMove dispatches an action of moving the Cursor
// to the new x,y coordinates
func (ac *ActionDispatcher) CursorMove(x, y int) {
	cma := action.NewCursorMove(x, y)
	ac.dispatcher.Dispatch(cma)
}

// SummonUnit summons the 'unit' from the player id 'pid' to the line
// 'plid' and with the current line id 'clid'
func (ac *ActionDispatcher) SummonUnit(unit string, pid, plid, clid int) {
	sua := action.NewSummonUnit(unit, pid, plid, clid)
	ac.dispatcher.Dispatch(sua)
}

// MoveUnit moves all the units
func (ac *ActionDispatcher) MoveUnit() {
	mua := action.NewMoveUnit()
	ac.dispatcher.Dispatch(mua)
}

// RemoveUnit removes the unit with the id 'uid'
func (ac *ActionDispatcher) RemoveUnit(uid int) {
	rua := action.NewRemoveUnit(uid)
	ac.dispatcher.Dispatch(rua)
}

// StealLive removes one live from the player with id 'fpid' and
// adds it to the player with id 'tpid'
func (ac *ActionDispatcher) StealLive(fpid, tpid int) {
	sla := action.NewStealLive(fpid, tpid)
	ac.dispatcher.Dispatch(sla)
}

// CameraZoom zooms the camera the direction 'd'
func (ac *ActionDispatcher) CameraZoom(d int) {
	cza := action.NewCameraZoom(d)
	ac.dispatcher.Dispatch(cza)
}

// PlaceTower places the tower 't' on the position X and Y on the line id 'lid'
func (ac *ActionDispatcher) PlaceTower(t string, x, y, lid int) {
	bta := action.NewPlaceTower(t, x, y, lid)
	ac.dispatcher.Dispatch(bta)
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
