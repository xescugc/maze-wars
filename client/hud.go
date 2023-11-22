package client

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
	"github.com/xescugc/ltw/assets"
	"github.com/xescugc/ltw/inputer"
	"github.com/xescugc/ltw/store"
	"github.com/xescugc/ltw/tower"
	"github.com/xescugc/ltw/unit"
	"github.com/xescugc/ltw/utils"
)

// HUDStore is in charge of keeping track of all the elements
// on the player HUD that are static and always seen
type HUDStore struct {
	*flux.ReduceStore

	game *Game

	cyclopeFacesetImage image.Image
	tilesetHouseImage   image.Image
	houseIcon           image.Image

	input inputer.Inputer
}

// HUDState stores the HUD state
type HUDState struct {
	Facesets []facesetButton

	CyclopeButton utils.Object
	SoldierButton utils.Object
	HouseButton   utils.Object

	SelectedTower   *SelectedTower
	TowerOpenMenuID string

	LastCursorPosition utils.Object
	CheckedPath        bool
}

type facesetButton struct {
	Unit   *unit.Unit
	Object utils.Object
}

type SelectedTower struct {
	store.Tower

	Invalid bool
}

// NewHUDStore creates a new HUDStore with the Dispatcher d and the Game g
func NewHUDStore(d *flux.Dispatcher, i inputer.Inputer, g *Game) (*HUDStore, error) {
	us := make([]*unit.Unit, 0, 0)
	for _, u := range unit.Units {
		us = append(us, u)
	}
	sort.Slice(us, func(i, j int) bool {
		return us[i].Gold < us[j].Gold
	})

	cs := g.Camera.GetState().(CameraState)
	// We want to create rows of 5
	fs := make([]facesetButton, 0, 0)
	nrows := len(us) / 5

	// As all the Faceset are equal squares
	// we just need to take one
	fhw := float64(us[0].Faceset.Bounds().Dx())
	for i, u := range us {
		fs = append(fs, facesetButton{
			Unit: u,
			Object: utils.Object{
				X: cs.W - (fhw * float64(5-(i%5))),
				Y: cs.H - (fhw * float64(nrows-(i/5))),
				W: fhw,
				H: fhw,
			},
		})
	}

	thi, _, err := image.Decode(bytes.NewReader(assets.TilesetHouse_png))
	if err != nil {
		return nil, err
	}

	hi, _, err := image.Decode(bytes.NewReader(assets.TilesetElement_png))
	if err != nil {
		return nil, err
	}

	hs := &HUDStore{
		game: g,

		tilesetHouseImage: ebiten.NewImageFromImage(thi).SubImage(image.Rect(5*16, 17*16, 5*16+16*2, 17*16+16*2)),
		houseIcon:         ebiten.NewImageFromImage(hi).SubImage(image.Rect(12*16, 0*16, 12*16+16, 0*16+16)),

		input: i,
	}
	hs.ReduceStore = flux.NewReduceStore(d, hs.Reduce, HUDState{
		Facesets: fs,
		SoldierButton: utils.Object{
			X: 0,
			Y: float64(cs.H - 16*2),
			W: float64(16 * 2),
			H: float64(16 * 2),
		},
		HouseButton: utils.Object{
			X: float64(cs.W - 16),
			Y: 0,
			W: float64(16),
			H: float64(16),
		},
	})

	return hs, nil
}

func (hs *HUDStore) Update() error {
	cs := hs.game.Camera.GetState().(CameraState)
	hst := hs.GetState().(HUDState)
	x, y := hs.input.CursorPosition()
	cp := hs.game.Store.Players.GetCurrentPlayer()
	tws := hs.game.Store.Towers.GetTowers()
	// Only send a CursorMove when the curso has actually moved
	if hst.LastCursorPosition.X != float64(x) || hst.LastCursorPosition.Y != float64(y) {
		actionDispatcher.CursorMove(x, y)
	}
	// If the Current player is dead or has no more lives there are no
	// mo actions that can be done
	// TODO Be able to move the camera when won or lose
	if cp.Lives == 0 || cp.Winner {
		return nil
	}
	if hs.input.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		click := utils.Object{
			X: float64(x),
			Y: float64(y),
			W: 1, H: 1,
		}
		clickAbsolute := utils.Object{
			X: float64(x) + cs.X,
			Y: float64(y) + cs.Y,
			W: 1, H: 1,
		}
		// Check what the user has just clicked
		for _, f := range hst.Facesets {
			if cp.Gold >= f.Unit.Gold && f.Object.IsColliding(click) {
				actionDispatcher.SummonUnit(f.Unit.Type.String(), cp.ID, cp.LineID, hs.game.Store.Map.GetNextLineID(cp.LineID))
				return nil
			}
		}
		if cp.Gold >= tower.Towers[tower.Soldier.String()].Gold && hst.SoldierButton.IsColliding(click) {
			actionDispatcher.SelectTower(tower.Soldier.String(), x, y)
			return nil
		}
		if hst.HouseButton.IsColliding(click) {
			actionDispatcher.GoHome()
			return nil
		}

		if hst.SelectedTower != nil && !hst.SelectedTower.Invalid {
			// We double check that placing the tower would not block the path
			utws := make([]utils.Object, 0, 0)
			for _, t := range tws {
				// If the tower does not belong to the current user then we can skip
				// as it's outside the Players Building Zone
				if t.PlayerID != cp.ID {
					continue
				}
				utws = append(utws, t.Object)
			}
			var fakex, fakey float64 = hs.game.Store.Map.GetRandomSpawnCoordinatesForLineID(cp.LineID)
			utws = append(utws, utils.Object{
				X: hst.SelectedTower.X + cs.X,
				Y: hst.SelectedTower.Y + cs.Y,
				H: hst.SelectedTower.H, W: hst.SelectedTower.W,
			})
			steps := hs.game.Store.Units.Astar(hs.game.Store.Map, cp.LineID, utils.MovingObject{
				Object: utils.Object{
					X: fakex,
					Y: fakey,
					W: 1, H: 1,
				},
			}, utws)
			if len(steps) != 0 {
				actionDispatcher.PlaceTower(hst.SelectedTower.Type, cp.ID, int(hst.SelectedTower.X+cs.X), int(hst.SelectedTower.Y+cs.Y))
			}
			return nil
		}
		for _, t := range tws {
			if clickAbsolute.IsColliding(t.Object) && cp.ID == t.PlayerID {
				if hst.TowerOpenMenuID != "" {
					// When the user clicks 2 times on the same tower we remove it
					if t.ID == hst.TowerOpenMenuID {
						actionDispatcher.RemoveTower(cp.ID, t.ID, t.Type)
						actionDispatcher.CloseTowerMenu()
						return nil
					}
				} else {
					actionDispatcher.OpenTowerMenu(t.ID)
					return nil
				}
			}
		}
		// If we are here no Tower was clicked but a click action was done,
		// so we check if the TowerOpenMenuID is set to unset it as this was
		// a click-off
		if hst.TowerOpenMenuID != "" {
			actionDispatcher.CloseTowerMenu()
		}
	}

	if cp.Gold >= tower.Towers[tower.Soldier.String()].Gold && hs.input.IsKeyJustPressed(ebiten.KeyT) {
		actionDispatcher.SelectTower(tower.Soldier.String(), x, y)
		return nil
	}
	if hst.TowerOpenMenuID != "" {
		if hs.input.IsKeyJustPressed(ebiten.KeyEscape) {
			actionDispatcher.CloseTowerMenu()
		}
	}
	if hst.SelectedTower != nil {
		if hs.input.IsMouseButtonJustPressed(ebiten.MouseButtonRight) || hs.input.IsKeyJustPressed(ebiten.KeyEscape) {
			actionDispatcher.DeselectTower(hst.SelectedTower.Type)
		} else {
			var invalid bool
			utws := make([]utils.Object, 0, 0)
			for _, t := range tws {
				// If the tower does not belong to the current user then we can skip
				// as it's outside the Players Building Zone
				if t.PlayerID != cp.ID {
					continue
				}
				utws = append(utws, t.Object)
				// The t.Object has the X and Y relative to the map
				// and the hst.SelectedTower has them relative to the
				// screen so we need to port the t.Object to the same
				// relative values
				neo := t.Object
				neo.X -= cs.X
				neo.Y -= cs.Y
				if hst.SelectedTower.IsColliding(neo) {
					invalid = true
					break
				}
			}
			neo := hst.SelectedTower.Object
			neo.X += cs.X
			neo.Y += cs.Y
			if !hs.game.Store.Map.IsInValidBuildingZone(neo, hst.SelectedTower.LineID) {
				invalid = true
			}

			// Only check if the line is blocked when is still valid position and it has not moved.
			// TODO: We can improve this by storing this result (if blocking or not) so we only validate
			// this once and not when the mouse is static with a selected tower
			if !invalid && (hst.LastCursorPosition.X == float64(x) && hst.LastCursorPosition.Y == float64(y) && !hst.CheckedPath) {
				var fakex, fakey float64 = hs.game.Store.Map.GetRandomSpawnCoordinatesForLineID(cp.LineID)
				utws = append(utws, utils.Object{
					X: hst.SelectedTower.X + cs.X,
					Y: hst.SelectedTower.Y + cs.Y,
					H: hst.SelectedTower.H, W: hst.SelectedTower.W,
				})
				steps := hs.game.Store.Units.Astar(hs.game.Store.Map, cp.LineID, utils.MovingObject{
					Object: utils.Object{
						X: fakex,
						Y: fakey,
						W: 1, H: 1,
					},
				}, utws)
				if len(steps) == 0 {
					invalid = true
				}
				actionDispatcher.CheckedPath(true)
			}
			if invalid != hst.SelectedTower.Invalid {
				actionDispatcher.SelectedTowerInvalid(invalid)
			}
		}
	}

	return nil
}

func (hs *HUDStore) Draw(screen *ebiten.Image) {
	hst := hs.GetState().(HUDState)
	cs := hs.game.Camera.GetState().(CameraState)
	cp := hs.game.Store.Players.GetCurrentPlayer()

	if cp.Lives == 0 {
		text.Draw(screen, "YOU LOST", smallFont, int(cs.W/2), int(cs.H/2), color.White)
	}

	if cp.Winner {
		text.Draw(screen, "YOU WON!", smallFont, int(cs.W/2), int(cs.H/2), color.White)
	}

	for _, f := range hst.Facesets {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(f.Object.X, f.Object.Y)
		if cp.Gold < f.Unit.Gold {
			op.ColorM.Scale(2, 0.5, 0.5, 0.9)
		}
		screen.DrawImage(f.Unit.Faceset.(*ebiten.Image), op)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(hst.SoldierButton.X, hst.SoldierButton.Y)
	if cp.Gold < tower.Towers[tower.Soldier.String()].Gold {
		op.ColorM.Scale(2, 0.5, 0.5, 0.9)
	} else if hst.SelectedTower != nil && hst.SelectedTower.Type == tower.Soldier.String() {
		// Once the tower is selected we gray it out
		op.ColorM.Scale(0.5, 0.5, 0.5, 0.5)
	}
	screen.DrawImage(hs.tilesetHouseImage.(*ebiten.Image), op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(hst.HouseButton.X, hst.HouseButton.Y)
	screen.DrawImage(hs.houseIcon.(*ebiten.Image), op)

	if hst.SelectedTower != nil {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(hst.SelectedTower.X/cs.Zoom, hst.SelectedTower.Y/cs.Zoom)
		op.GeoM.Scale(cs.Zoom, cs.Zoom)

		if hst.SelectedTower != nil && hst.SelectedTower.Invalid {
			op.ColorM.Scale(2, 0.5, 0.5, 0.9)
		}

		screen.DrawImage(hst.SelectedTower.Image().(*ebiten.Image), op)
	}

	psit := hs.game.Store.Players.GetState().(store.PlayersState).IncomeTimer
	players := hs.game.Store.Players.GetPlayers()
	text.Draw(screen, fmt.Sprintf("Income Timer: %ds", psit), smallFont, 0, 0, color.White)
	var pcount = 1
	var sortedPlayers = make([]*store.Player, 0, 0)
	for _, p := range players {
		sortedPlayers = append(sortedPlayers, p)
	}
	sort.Slice(sortedPlayers, func(i, j int) bool { return sortedPlayers[i].LineID < sortedPlayers[j].LineID })
	for _, p := range sortedPlayers {
		text.Draw(screen, fmt.Sprintf("Name: %s, Lives: %d, Gold: %d, Income: %d", p.Name, p.Lives, p.Gold, p.Income), smallFont, 0, 15*pcount, color.White)
		pcount++
	}
}

func (hs *HUDStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	hstate, ok := state.(HUDState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.WindowResizing:
		hs.GetDispatcher().WaitFor(hs.game.Camera.GetDispatcherToken())
		cs := hs.game.Camera.GetState().(CameraState)

		us := make([]*unit.Unit, 0, 0)
		for _, u := range unit.Units {
			us = append(us, u)
		}
		sort.Slice(us, func(i, j int) bool {
			return us[i].Gold < us[j].Gold
		})

		// We want to create rows of 5
		fs := make([]facesetButton, 0, 0)
		nrows := len(us) / 5

		// As all the Faceset are equal squares
		// we just need to take one
		fhw := float64(us[0].Faceset.Bounds().Dx())
		for i, u := range us {
			fs = append(fs, facesetButton{
				Unit: u,
				Object: utils.Object{
					X: cs.W - (fhw * float64(5-(i%5))),
					Y: cs.H - (fhw * float64(nrows-(i/5))),
					W: fhw,
					H: fhw,
				},
			})
		}
		hstate.Facesets = fs
		hstate.SoldierButton = utils.Object{
			X: 0,
			Y: float64(cs.H - 16*2),
			W: float64(16 * 2),
			H: float64(16 * 2),
		}
		hstate.HouseButton = utils.Object{
			X: float64(cs.W - 16),
			Y: 0,
			W: float64(16),
			H: float64(16),
		}
	case action.SelectTower:
		hs.GetDispatcher().WaitFor(hs.game.Store.Players.GetDispatcherToken())
		cp := hs.game.Store.Players.GetCurrentPlayer()
		// TODO: Insead of hardcoding the image and W, H we should
		// use the Type on the action to select the right image
		hstate.SelectedTower = &SelectedTower{
			Tower: store.Tower{
				Object: utils.Object{
					X: float64(act.SelectTower.X) - (hstate.SoldierButton.W / 2),
					Y: float64(act.SelectTower.Y) - (hstate.SoldierButton.H / 2),
					W: hstate.SoldierButton.W,
					H: hstate.SoldierButton.H,
				},
				Type:   act.SelectTower.Type,
				LineID: cp.LineID,
			},
		}
	case action.CursorMove:
		// We update the last seen cursor position to not resend unnecessary events
		hstate.LastCursorPosition.X = float64(act.CursorMove.X)
		hstate.LastCursorPosition.Y = float64(act.CursorMove.Y)

		if hstate.SelectedTower != nil {
			// We find the closes multiple in case the cursor moves too fast, between FPS reloads,
			// and lands in a position not 'multiple' which means the position of the SelectedTower
			// is not updated and the result is the cursor far away from the Drawing of the SelectedTower
			// as it has stayed on the previous position
			var multiple int = 8
			if act.CursorMove.X%multiple == 0 {
				hstate.SelectedTower.X = float64(act.CursorMove.X) - (hstate.SoldierButton.W / 2)
			} else if math.Abs(float64(act.CursorMove.X)-hstate.SelectedTower.X) > float64(multiple) {
				hstate.SelectedTower.X = float64(closestMultiple(act.CursorMove.X, multiple)) - (hstate.SoldierButton.W / 2)
			}
			if act.CursorMove.Y%multiple == 0 {
				hstate.SelectedTower.Y = float64(act.CursorMove.Y) - (hstate.SoldierButton.H / 2)
			} else if math.Abs(float64(act.CursorMove.Y)-hstate.SelectedTower.Y) > float64(multiple) {
				hstate.SelectedTower.Y = float64(closestMultiple(act.CursorMove.Y, multiple)) - (hstate.SoldierButton.H / 2)
			}
		}
		// If it has moved we set the CheckedPath as not checked as it's only checked
		// when the Cursor has not moved
		hstate.CheckedPath = false
	case action.PlaceTower, action.DeselectTower:
		hstate.SelectedTower = nil
	case action.SelectedTowerInvalid:
		if hstate.SelectedTower != nil {
			hstate.SelectedTower.Invalid = act.SelectedTowerInvalid.Invalid
		}
	case action.OpenTowerMenu:
		hstate.TowerOpenMenuID = act.OpenTowerMenu.TowerID
	case action.CloseTowerMenu:
		hstate.TowerOpenMenuID = ""
	case action.CheckedPath:
		hstate.CheckedPath = act.CheckedPath.Checked
	default:
	}

	return hstate
}

// closestMultiple finds the coses multiple of 'b' for the number 'a'
func closestMultiple(a, b int) int {
	a = a + b/2
	a = a - (a % b)
	return a
}
