package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/ltw/store"
	"github.com/xescugc/ltw/tower"
)

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
				if d <= tower.Towers[t.Type].Range && d < minDist {
					minDist = d
					minDistUnit = uid
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
