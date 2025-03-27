package game

import (
	"bytes"
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/xescugc/maze-wars/assets"
	cutils "github.com/xescugc/maze-wars/client/utils"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/unit/ability"
	"github.com/xescugc/maze-wars/unit/buff"
	"github.com/xescugc/maze-wars/utils"
)

type Lines struct {
	game *Game

	tilesetLogicImage image.Image
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

	ls := &Lines{
		game:              g,
		tilesetLogicImage: ebiten.NewImageFromImage(tli).SubImage(image.Rect(4*16, 5*16, 4*16+16, 5*16+16)),
	}

	return ls, nil
}

func (ls *Lines) Update() error {
	b := time.Now()
	defer utils.LogTime(ls.game.Logger, b, "lines update")

	return nil
}

func (ls *Lines) Draw(screen *ebiten.Image) {
	b := time.Now()
	defer utils.LogTime(ls.game.Logger, b, "lines draw")

	hst := ls.game.HUD.GetState()
	for _, l := range ls.game.Store.Game.ListLines() {
		for _, t := range l.Towers {
			ls.DrawTower(screen, ls.game.Camera, t)
		}
		for _, u := range l.ListSortedUnits() {
			ls.DrawUnit(screen, ls.game.Camera, u)
		}
		for _, t := range l.Towers {
			ls.DrawTowerHelath(screen, ls.game.Camera, t)
		}
		if hst.OpenTowerMenu != nil {
			t, ok := l.Towers[hst.OpenTowerMenu.ID]
			if ok {
				ls.DrawTowerSelected(screen, ls.game.Camera, t)
			}
		} else if hst.OpenUnitMenu != nil {
			u, ok := l.Units[hst.OpenUnitMenu.ID]
			if ok {
				ls.DrawUnitSelected(screen, ls.game.Camera, u)
			}
		}
		for _, p := range l.Projectiles {
			// TODO: Why is this happening?
			if p.ImageKey == "" {
				continue
			}
			ls.DrawProjectile(screen, ls.game.Camera, p)
		}
	}
}

func (ls *Lines) DrawTower(screen *ebiten.Image, c *CameraStore, t *store.Tower) {
	cs := c.GetState()
	if !t.IsColliding(cs.Object) {
		return
	}
	x := float64(t.X - cs.X)
	y := float64(t.Y - cs.Y)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	op.GeoM.Scale(cs.Zoom, cs.Zoom)
	screen.DrawImage(cutils.Images.Get(t.IdleKey()), op)

	//x := float64(p.X - cs.X)
	//y := float64(p.Y - cs.Y)
	//ni := ebiten.NewImage(4, 4)
	//ni.Fill(color.Black)

	//op := &ebiten.DrawImageOptions{}
	//op.GeoM.Translate(x, y)
	//op.GeoM.Scale(cs.Zoom, cs.Zoom)
	//screen.DrawImage(ni, op)
}

func (ls *Lines) DrawTowerHelath(screen *ebiten.Image, c *CameraStore, t *store.Tower) {
	cs := c.GetState()
	if !t.IsColliding(cs.Object) {
		return
	}

	ot := tower.Towers[t.Type]
	// Only draw the Health bar if the Tower has been hit
	if t.Health != ot.Health {
		lbui := cutils.Images.Get(cutils.LifeBarBigUnderKey)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(t.X-cs.X-1, t.Y-cs.Y-float64(lbui.Bounds().Dy()))
		screen.DrawImage(lbui, op)

		lbpi := cutils.Images.Get(cutils.LifeBarBigProgressKey)
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(t.X-cs.X-1, t.Y-cs.Y-float64(lbpi.Bounds().Dy()))
		screen.DrawImage(lbpi.SubImage(image.Rect(0, 0, int(float64(lbpi.Bounds().Dx())*(t.Health/ot.Health)), lbpi.Bounds().Dy())).(*ebiten.Image), op)
	}
}

func (ls *Lines) DrawTowerSelected(screen *ebiten.Image, c *CameraStore, t *store.Tower) {
	cs := c.GetState()
	if !t.IsColliding(cs.Object) {
		return
	}
	x := float64(t.X - cs.X)
	y := float64(t.Y - cs.Y)
	ot := tower.Towers[t.Type]

	vector.StrokeRect(screen, float32(x-1), float32(y-1), 34, 34, 2, cutils.Green, false)
	vector.StrokeCircle(screen, float32(x+16), float32(y+16), float32(ot.Range*32+16), 2, cutils.Green, false)
}

func (ls *Lines) DrawUnitSelected(screen *ebiten.Image, c *CameraStore, u *store.Unit) {
	cs := c.GetState()
	if !u.IsColliding(cs.Object) {
		return
	}
	x := float64(u.X - cs.X)
	y := float64(u.Y - cs.Y)

	vector.StrokeRect(screen, float32(x-1), float32(y-1), 18, 18, 2, cutils.Green, false)
}

func (ls *Lines) DrawUnit(screen *ebiten.Image, c *CameraStore, u *store.Unit) {
	cs := c.GetState()
	// This is to display the full unit calculated path as a line
	// used for testing visually the path
	//for _, s := range u.Path {
	//screen.Set(s.X-cs.X, s.Y-cs.Y, color.Black)
	//}
	if !u.IsColliding(cs.Object) {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(u.X-cs.X, u.Y-cs.Y)
	op.GeoM.Scale(cs.Zoom, cs.Zoom)
	sx := directionToTile[u.Facing] * u.W
	i := (u.MovingCount / 5) % 4
	sy := i * u.H
	if u.HasBuff(buff.Burrowoed) {
		if u.CanUnburrow(time.Now()) {
			screen.DrawImage(cutils.Images.Get(cutils.BuffBurrowedReadyKey), op)
		} else {
			screen.DrawImage(cutils.Images.Get(cutils.BuffBurrowedKey), op)
		}
	} else if u.HasBuff(buff.Resurrecting) {
		screen.DrawImage(cutils.Images.Get(cutils.BuffResurrectingKey), op)
	} else if u.HasAbility(ability.Attack) && len(u.Path) == 0 {
		if (u.AnimationCount/10)%2 == 0 {
			screen.DrawImage(cutils.Images.Get(u.AttackKey()).SubImage(image.Rect(sx, 0, sx+u.W, u.H)).(*ebiten.Image), op)
		} else {
			screen.DrawImage(cutils.Images.Get(u.IdleKey()).SubImage(image.Rect(sx, 0, sx+u.W, u.H)).(*ebiten.Image), op)
		}
	} else {
		screen.DrawImage(cutils.Images.Get(u.WalkKey()).SubImage(image.Rect(sx, sy, sx+u.W, sy+u.H)).(*ebiten.Image), op)
	}

	// Only draw the Health bar if the unit has been hit
	if u.Health != u.MaxHealth {
		lbui := cutils.Images.Get(cutils.LifeBarUnderKey)
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(u.X-cs.X-1, u.Y-cs.Y-float64(lbui.Bounds().Dy()))
		screen.DrawImage(lbui, op)

		lbpi := cutils.Images.Get(cutils.LifeBarProgressKey)
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(u.X-cs.X-1, u.Y-cs.Y-float64(lbpi.Bounds().Dy()))
		screen.DrawImage(lbpi.SubImage(image.Rect(0, 0, int(float64(lbpi.Bounds().Dx())*(u.Health/u.MaxHealth)), lbpi.Bounds().Dy())).(*ebiten.Image), op)
	}

	// Only draw the Shield bar if the unit has been hit
	if u.Shield != u.MaxShield && u.Shield != 0 {
		lbui := cutils.Images.Get(cutils.LifeBarUnderKey)
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(u.X-cs.X-1, u.Y-cs.Y-float64(lbui.Bounds().Dy()))
		screen.DrawImage(lbui, op)

		sbpi := cutils.Images.Get(cutils.ShieldBarProgressKey)
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(u.X-cs.X-1, u.Y-cs.Y-float64(sbpi.Bounds().Dy()))
		screen.DrawImage(sbpi.SubImage(image.Rect(0, 0, int(float64(sbpi.Bounds().Dx())*(u.Shield/u.MaxShield)), sbpi.Bounds().Dy())).(*ebiten.Image), op)
	}

	// TODO: Animation logic
	//if u.HasBuff(buff.Burrowoed) {
	//i := (u.AnimationCount / 15) % 8
	//op = &ebiten.DrawImageOptions{}
	//op.GeoM.Translate(u.X-cs.X-float64(u.W/2), u.Y-cs.Y-float64(u.H/2))
	//img := cutils.Images.Get(buffBurrowedKey)
	//screen.DrawImage(img.SubImage(image.Rect(i*32, 0, i*32+32, i*32+32)).(*ebiten.Image), op)
	//}
}

func (ls *Lines) DrawProjectile(screen *ebiten.Image, c *CameraStore, p *store.Projectile) {
	cs := c.GetState()
	if !p.IsColliding(cs.Object) {
		return
	}
	x := float64(p.X - cs.X)
	y := float64(p.Y - cs.Y)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	op.GeoM.Scale(cs.Zoom, cs.Zoom)
	screen.DrawImage(cutils.Images.Get(p.ImageKey), op)
	//screen.DrawImage(ni, op)
}
