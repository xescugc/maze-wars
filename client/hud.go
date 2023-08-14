package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
)

//go:embed assets/cyclope/Faceset.png
var CyclopeFaceset_png []byte

//go:embed assets/TilesetHouse.png
var TilesetHouse_png []byte

// HUDStore is in charge of keeping track of all the elements
// on the player HUD that are static and always seen
type HUDStore struct {
	*flux.ReduceStore

	game *Game

	cyclopeFacesetImage image.Image
	tilesetHouseImage   image.Image
}

// HUDState stores the HUD state
type HUDState struct {
	CyclopeButton Object
	SoldierButton Object

	SelectedTower *SelectedTower
}

type SelectedTower struct {
	Tower

	Invalid bool
}

// NewHUDStore creates a new HUDStore with the Dispatcher d and the Game g
func NewHUDStore(d *flux.Dispatcher, g *Game) (*HUDStore, error) {
	fi, _, err := image.Decode(bytes.NewReader(CyclopeFaceset_png))
	if err != nil {
		return nil, err
	}

	thi, _, err := image.Decode(bytes.NewReader(TilesetHouse_png))
	if err != nil {
		return nil, err
	}

	hs := &HUDStore{
		game: g,

		cyclopeFacesetImage: ebiten.NewImageFromImage(fi),
		tilesetHouseImage:   ebiten.NewImageFromImage(thi).SubImage(image.Rect(5*16, 17*16, 5*16+16*2, 17*16+16*2)),
	}
	hs.ReduceStore = flux.NewReduceStore(d, hs.Reduce, HUDState{
		CyclopeButton: Object{
			X: float64(g.Screen.GetWidth() - hs.cyclopeFacesetImage.Bounds().Dx()),
			Y: float64(g.Screen.GetHeight() - hs.cyclopeFacesetImage.Bounds().Dy()),
			W: float64(hs.cyclopeFacesetImage.Bounds().Dx()),
			H: float64(hs.cyclopeFacesetImage.Bounds().Dy()),
		},
		SoldierButton: Object{
			X: 0,
			Y: float64(g.Screen.GetHeight() - 16*2),
			W: float64(16 * 2),
			H: float64(16 * 2),
		},
	})

	return hs, nil
}

func (hs *HUDStore) Update() error {
	x, y := ebiten.CursorPosition()
	actionDispatcher.CursorMove(x, y)
	hst := hs.GetState().(HUDState)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		obj := Object{
			X: float64(x),
			Y: float64(y),
			W: 1, H: 1,
		}
		// Check what the user has just clicked
		if hst.CyclopeButton.IsColliding(obj) {
			cp := hs.game.Players.GetCurrentPlayer()
			actionDispatcher.SummonUnit("cyclope", cp.ID, cp.LineID, hs.game.Map.GetNextLineID(cp.LineID))
			return nil
		}
		if hst.SoldierButton.IsColliding(obj) {
			actionDispatcher.SelectTower("soldier", x, y)
			return nil
		}

		if hst.SelectedTower != nil && !hst.SelectedTower.Invalid {
			cs := hs.game.Camera.GetState().(CameraState)
			actionDispatcher.PlaceTower(hst.SelectedTower.Type, int(hst.SelectedTower.X+cs.X), int(hst.SelectedTower.Y+cs.Y), hst.SelectedTower.LineID)
		}
	}
	if hst.SelectedTower != nil {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			actionDispatcher.DeselectTower(hst.SelectedTower.Type)
		} else {
			ts := hs.game.Towers.GetState().(TowersState)
			var invalid bool
			for _, t := range ts.Towers {
				cs := hs.game.Camera.GetState().(CameraState)
				// The t.Entity has the X and Y relative to the map
				// and the hst.SelectedTower has them relative to the
				// screen so we need to port the t.Entity to the same
				// relative values
				neo := t.Entity.Object
				neo.X -= cs.X
				neo.Y -= cs.Y
				if hst.SelectedTower.IsColliding(neo) {
					actionDispatcher.SelectedTowerInvalid(true)
					invalid = true
					break
				}
			}
			if !invalid && hst.SelectedTower.Invalid {
				actionDispatcher.SelectedTowerInvalid(false)
			}
		}
	}

	return nil
}

func (hs *HUDStore) Draw(screen *ebiten.Image) {
	hst := hs.GetState().(HUDState)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(hst.CyclopeButton.X, hst.CyclopeButton.Y)
	screen.DrawImage(hs.cyclopeFacesetImage.(*ebiten.Image), op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(hst.SoldierButton.X, hst.SoldierButton.Y)
	if hst.SelectedTower != nil && hst.SelectedTower.Type == "soldier" {
		// Once the tower is selected we gray it out
		op.ColorM.Scale(0.5, 0.5, 0.5, 0.5)
	}
	screen.DrawImage(hs.tilesetHouseImage.(*ebiten.Image), op)

	if hst.SelectedTower != nil {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(hst.SelectedTower.X, hst.SelectedTower.Y)

		if hst.SelectedTower != nil && hst.SelectedTower.Invalid {
			op.ColorM.Scale(2, 0.5, 0.5, 0.9)
		}

		screen.DrawImage(hst.SelectedTower.Image.(*ebiten.Image), op)
	}

	cp := hs.game.Players.GetCurrentPlayer()
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Lives: %d", cp.Lives))
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
	case action.CameraZoom:
		hs.GetDispatcher().WaitFor(hs.game.Screen.GetDispatcherToken())
		hstate.CyclopeButton = Object{
			X: float64(hs.game.Screen.GetWidth() - hs.cyclopeFacesetImage.Bounds().Dx()),
			Y: float64(hs.game.Screen.GetHeight() - hs.cyclopeFacesetImage.Bounds().Dy()),
			W: float64(hs.cyclopeFacesetImage.Bounds().Dx()),
			H: float64(hs.cyclopeFacesetImage.Bounds().Dy()),
		}
		hstate.SoldierButton = Object{
			X: 0,
			Y: float64(hs.game.Screen.GetHeight() - 16*2),
			W: float64(16 * 2),
			H: float64(16 * 2),
		}
	case action.SelectTower:
		hs.GetDispatcher().WaitFor(hs.game.Players.GetDispatcherToken())
		cp := hs.game.Players.GetCurrentPlayer()
		// TODO: Insead of hardcoding the image and W, H we should
		// use the Type on the action to select the right image
		hstate.SelectedTower = &SelectedTower{
			Tower: Tower{
				Entity: Entity{
					Object: Object{
						X: float64(act.SelectTower.X) - (hstate.SoldierButton.W / 2),
						Y: float64(act.SelectTower.Y) - (hstate.SoldierButton.H / 2),
						W: hstate.SoldierButton.W,
						H: hstate.SoldierButton.H,
					},
					Image: hs.tilesetHouseImage,
				},
				Type:   act.SelectTower.Type,
				LineID: cp.LineID,
			},
		}
	case action.CursorMove:
		if hstate.SelectedTower != nil {
			hstate.SelectedTower.X = float64(act.CursorMove.X) - (hstate.SoldierButton.W / 2)
			hstate.SelectedTower.Y = float64(act.CursorMove.Y) - (hstate.SoldierButton.H / 2)
		}
	case action.PlaceTower, action.DeselectTower:
		hstate.SelectedTower = nil
	case action.SelectedTowerInvalid:
		if hstate.SelectedTower != nil {
			hstate.SelectedTower.Invalid = act.SelectedTowerInvalid.Invalid
		}
	default:
	}

	return hstate
}