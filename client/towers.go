package client

import (
	"bytes"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/ltw/assets"
	"github.com/xescugc/ltw/store"
	"github.com/xescugc/ltw/tower"
)

type Towers struct {
	game *Game

	tilesetLogicImage image.Image
}

func NewTowers(g *Game) (*Towers, error) {
	tli, _, err := image.Decode(bytes.NewReader(assets.TilesetLogic_png))
	if err != nil {
		return nil, err
	}

	ts := &Towers{
		game:              g,
		tilesetLogicImage: ebiten.NewImageFromImage(tli).SubImage(image.Rect(4*16, 5*16, 4*16+16, 5*16+16)),
	}

	return ts, nil
}

func (ts *Towers) Update() error {
	uts := ts.game.Store.Units.List()
	tws := ts.game.Store.Towers.List()
	cp := ts.game.Store.Players.FindCurrent()
	for _, t := range tws {
		if t.PlayerID != cp.ID {
			continue
		}
		// If there are any units then we check if we can attack them
		if len(uts) != 0 {
			var (
				minDist     float64 = 0
				minDistUnit string
			)
			for _, u := range uts {
				if u.CurrentLineID != cp.LineID {
					continue
				}
				d := t.PDistance(u.Object)
				if minDist == 0 {
					minDist = d
				}
				if d <= tower.Towers[t.Type].Range && d <= minDist {
					minDist = d
					minDistUnit = u.ID
				}
			}
			if minDistUnit != "" {
				actionDispatcher.TowerAttack(minDistUnit, t.Type)
			}
		}
	}
	return nil
}

func (ts *Towers) Draw(screen *ebiten.Image) {
	for _, t := range ts.game.Store.Towers.List() {
		ts.DrawTower(screen, ts.game.Camera, t)
	}
}

func (ts *Towers) DrawTower(screen *ebiten.Image, c *CameraStore, t *store.Tower) {
	cs := c.GetState().(CameraState)
	hst := ts.game.HUD.GetState().(HUDState)
	if !t.IsColliding(cs.Object) {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(t.X-cs.X, t.Y-cs.Y)
	op.GeoM.Scale(cs.Zoom, cs.Zoom)
	screen.DrawImage(ebiten.NewImageFromImage(t.Faceset()), op)

	if t.ID == hst.TowerOpenMenuID {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(t.X-cs.X+8, t.Y-cs.Y+8)
		op.GeoM.Scale(cs.Zoom, cs.Zoom)
		screen.DrawImage(ts.tilesetLogicImage.(*ebiten.Image), op)
	}
}
