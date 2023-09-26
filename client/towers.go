package main

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/ltw/assets"
	"github.com/xescugc/ltw/store"
)

var (
	towerImages         = make(map[string]image.Image)
	towerRange  float64 = 16 * 2
	towerDamage         = 1
	towerGold           = 10
)

func init() {
	si, _, err := image.Decode(bytes.NewReader(assets.TilesetHouse_png))
	if err != nil {
		log.Fatal(err)
	}

	towerImages["soldier"] = ebiten.NewImageFromImage(si).SubImage(image.Rect(5*16, 17*16, 5*16+16*2, 17*16+16*2))
}

type Towers struct {
	game *Game
}

func NewTowers(g *Game) *Towers {
	ts := &Towers{
		game: g,
	}

	return ts
}

func (ts *Towers) Update() error {
	uts := ts.game.Store.Units.GetState().(store.UnitsState).Units
	tws := ts.game.Store.Towers.GetState().(store.TowersState).Towers
	if len(uts) != 0 {
		for _, t := range tws {
			var (
				minDist     float64 = 0
				minDistUnit string
			)
			for uid, u := range uts {
				d := t.Distance(u.Object)
				if minDist == 0 {
					minDist = d
				}
				if d <= towerRange && d < minDist {
					minDist = d
					minDistUnit = uid
				}
			}
			if minDistUnit != "" {
				actionDispatcher.TowerAttack(minDistUnit)
			}
		}
	}
	return nil
}

func (ts *Towers) Draw(screen *ebiten.Image) {
	for _, t := range ts.game.Store.Towers.GetState().(store.TowersState).Towers {
		ts.DrawTower(screen, ts.game.Camera, t)
	}
}

func (ts *Towers) DrawTower(screen *ebiten.Image, c *CameraStore, t *store.Tower) {
	cs := c.GetState().(CameraState)
	if !t.IsColliding(cs.Object) {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(t.X-cs.X, t.Y-cs.Y)
	op.GeoM.Scale(cs.Zoom, cs.Zoom)
	screen.DrawImage(t.Image().(*ebiten.Image), op)
}
