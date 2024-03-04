package client

import (
	"bytes"
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/maze-wars/assets"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/unit"
	"github.com/xescugc/maze-wars/utils"
)

type Lines struct {
	game *Game

	tilesetLogicImage image.Image
	lifeBarProgress   image.Image
	lifeBarUnder      image.Image
}

var (
	directionToTile = map[utils.Direction]int{
		utils.Down:  0,
		utils.Up:    1,
		utils.Left:  2,
		utils.Right: 3,
	}
)

func NewLines(g *Game) (*Lines, error) {
	tli, _, err := image.Decode(bytes.NewReader(assets.TilesetLogic_png))
	if err != nil {
		return nil, err
	}

	lbpi, _, err := image.Decode(bytes.NewReader(assets.LifeBarMiniProgress_png))
	if err != nil {
		return nil, err
	}

	lbui, _, err := image.Decode(bytes.NewReader(assets.LifeBarMiniUnder_png))
	if err != nil {
		return nil, err
	}

	ls := &Lines{
		game:              g,
		tilesetLogicImage: ebiten.NewImageFromImage(tli).SubImage(image.Rect(4*16, 5*16, 4*16+16, 5*16+16)),
		lifeBarProgress:   ebiten.NewImageFromImage(lbpi),
		lifeBarUnder:      ebiten.NewImageFromImage(lbui),
	}

	return ls, nil
}

func (ls *Lines) Update() error {
	b := time.Now()
	defer utils.LogTime(ls.game.Logger, b, "lines update")

	cp := ls.game.Store.Players.FindCurrent()
	l := ls.game.Store.Lines.FindByID(cp.LineID)
	for _, t := range l.Towers {
		if t.PlayerID != cp.ID {
			continue
		}
		var (
			minDist     float64 = 0
			minDistUnit string
		)
		for _, u := range l.Units {
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

	for _, u := range l.Units {
		// Only dispatch events from your own line
		if u.Health == 0 {
			p := ls.game.Store.Players.FindByLineID(u.CurrentLineID)
			actionDispatcher.UnitKilled(p.ID, u.Type)
			actionDispatcher.RemoveUnit(u.ID)
			continue
		}
		if n := l.Graph.GetNodeOf(int(u.X), int(u.Y)); n != nil && n.IsDeathZone {
			p := ls.game.Store.Players.FindByLineID(u.CurrentLineID)
			nlid := ls.game.Store.Map.GetNextLineID(u.CurrentLineID)
			actionDispatcher.StealLive(p.ID, u.PlayerID)
			if nlid == u.PlayerLineID {
				actionDispatcher.RemoveUnit(u.ID)
			} else {
				actionDispatcher.ChangeUnitLine(u.ID)
			}
		}
	}
	return nil
}

func (ls *Lines) Draw(screen *ebiten.Image) {
	b := time.Now()
	defer utils.LogTime(ls.game.Logger, b, "lines draw")

	for _, l := range ls.game.Store.Lines.List() {
		for _, t := range l.Towers {
			ls.DrawTower(screen, ls.game.Camera, t)
		}
		for _, u := range l.Units {
			ls.DrawUnit(screen, ls.game.Camera, u)
		}
	}
}

func (ls *Lines) DrawTower(screen *ebiten.Image, c *CameraStore, t *store.Tower) {
	cs := c.GetState().(CameraState)
	hst := ls.game.HUD.GetState().(HUDState)
	if !t.IsColliding(cs.Object) {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(t.X-cs.X, t.Y-cs.Y)
	op.GeoM.Scale(cs.Zoom, cs.Zoom)
	screen.DrawImage(imagesCache.Get(t.FacetKey()), op)

	if t.ID == hst.TowerOpenMenuID {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(t.X-cs.X+8, t.Y-cs.Y+8)
		op.GeoM.Scale(cs.Zoom, cs.Zoom)
		screen.DrawImage(ls.tilesetLogicImage.(*ebiten.Image), op)
	}
}

func (ls *Lines) DrawUnit(screen *ebiten.Image, c *CameraStore, u *store.Unit) {
	cs := c.GetState().(CameraState)
	// This is to display the full unit calculated path as a line
	// used for testing visually the path
	//for _, s := range u.Path {
	//screen.Set(s.X-int(cs.X), s.Y-int(cs.Y), color.Black)
	//}
	if !u.IsColliding(cs.Object) {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(u.X-cs.X, u.Y-cs.Y)
	op.GeoM.Scale(cs.Zoom, cs.Zoom)
	sx := directionToTile[u.Facing] * int(u.W)
	i := (u.MovingCount / 5) % 4
	sy := i * int(u.H)
	screen.DrawImage(imagesCache.Get(u.SpriteKey()).SubImage(image.Rect(sx, sy, sx+int(u.W), sy+int(u.H))).(*ebiten.Image), op)

	// Only draw the Health bar if the unit has been hit
	h := unit.Units[u.Type].Health
	if unit.Units[u.Type].Health != u.Health {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(u.X-cs.X, u.Y-cs.Y-float64(ls.lifeBarUnder.Bounds().Dy()))
		screen.DrawImage(ls.lifeBarUnder.(*ebiten.Image), op)

		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(u.X-cs.X, u.Y-cs.Y-float64(ls.lifeBarProgress.Bounds().Dy()))
		screen.DrawImage(ls.lifeBarProgress.(*ebiten.Image).SubImage(image.Rect(0, 0, int(float64(ls.lifeBarProgress.Bounds().Dx())*(u.Health/h)), ls.lifeBarProgress.Bounds().Dy())).(*ebiten.Image), op)
	}
}
