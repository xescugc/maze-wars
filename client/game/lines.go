package game

import (
	"bytes"
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/maze-wars/assets"
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

	for _, l := range ls.game.Store.Lines.ListLines() {
		for _, t := range l.Towers {
			ls.DrawTower(screen, ls.game.Camera, t)
		}
		for _, u := range l.ListSortedUnits() {
			ls.DrawUnit(screen, ls.game.Camera, u)
		}
		for _, t := range l.Towers {
			ls.DrawTowerHelath(screen, ls.game.Camera, t)
		}
	}
}

func (ls *Lines) DrawTower(screen *ebiten.Image, c *CameraStore, t *store.Tower) {
	cs := c.GetState().(CameraState)
	if !t.IsColliding(cs.Object) {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(t.X-cs.X), float64(t.Y-cs.Y))
	op.GeoM.Scale(cs.Zoom, cs.Zoom)
	screen.DrawImage(imagesCache.Get(t.FacetKey()), op)
}

func (ls *Lines) DrawTowerHelath(screen *ebiten.Image, c *CameraStore, t *store.Tower) {
	cs := c.GetState().(CameraState)
	if !t.IsColliding(cs.Object) {
		return
	}

	ot := tower.Towers[t.Type]
	// Only draw the Health bar if the Tower has been hit
	if t.Health != ot.Health {
		lbui := imagesCache.Get(lifeBarBigUnderKey)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(t.X-cs.X-1, t.Y-cs.Y-float64(lbui.Bounds().Dy()))
		screen.DrawImage(lbui, op)

		lbpi := imagesCache.Get(lifeBarBigProgressKey)
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(t.X-cs.X-1, t.Y-cs.Y-float64(lbpi.Bounds().Dy()))
		screen.DrawImage(lbpi.SubImage(image.Rect(0, 0, int(float64(lbpi.Bounds().Dx())*(t.Health/ot.Health)), lbpi.Bounds().Dy())).(*ebiten.Image), op)
	}
}

func (ls *Lines) DrawUnit(screen *ebiten.Image, c *CameraStore, u *store.Unit) {
	cs := c.GetState().(CameraState)
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
			screen.DrawImage(imagesCache.Get(buffBurrowedReadyKey), op)
		} else {
			screen.DrawImage(imagesCache.Get(buffBurrowedKey), op)
		}
	} else if u.HasBuff(buff.Resurrecting) {
		screen.DrawImage(imagesCache.Get(buffResurrectingKey), op)
	} else if u.HasAbility(ability.Attack) && len(u.Path) == 0 {
		if (u.AnimationCount/10)%2 == 0 {
			screen.DrawImage(imagesCache.Get(u.AttackKey()).SubImage(image.Rect(sx, 0, sx+u.W, u.H)).(*ebiten.Image), op)
		} else {
			screen.DrawImage(imagesCache.Get(u.IdleKey()).SubImage(image.Rect(sx, 0, sx+u.W, u.H)).(*ebiten.Image), op)
		}
	} else {
		screen.DrawImage(imagesCache.Get(u.WalkKey()).SubImage(image.Rect(sx, sy, sx+u.W, sy+u.H)).(*ebiten.Image), op)
	}

	// Only draw the Health bar if the unit has been hit
	if u.Health != u.MaxHealth {
		lbui := imagesCache.Get(lifeBarUnderKey)
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(u.X-cs.X-1, u.Y-cs.Y-float64(lbui.Bounds().Dy()))
		screen.DrawImage(lbui, op)

		lbpi := imagesCache.Get(lifeBarProgressKey)
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(u.X-cs.X-1, u.Y-cs.Y-float64(lbpi.Bounds().Dy()))
		screen.DrawImage(lbpi.SubImage(image.Rect(0, 0, int(float64(lbpi.Bounds().Dx())*(u.Health/u.MaxHealth)), lbpi.Bounds().Dy())).(*ebiten.Image), op)
	}

	// Only draw the Shield bar if the unit has been hit
	if u.Shield != u.MaxShield && u.Shield != 0 {
		lbui := imagesCache.Get(lifeBarUnderKey)
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(u.X-cs.X-1, u.Y-cs.Y-float64(lbui.Bounds().Dy()))
		screen.DrawImage(lbui, op)

		sbpi := imagesCache.Get(shieldBarProgressKey)
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(u.X-cs.X-1, u.Y-cs.Y-float64(sbpi.Bounds().Dy()))
		screen.DrawImage(sbpi.SubImage(image.Rect(0, 0, int(float64(sbpi.Bounds().Dx())*(u.Shield/u.MaxShield)), sbpi.Bounds().Dy())).(*ebiten.Image), op)
	}

	// TODO: Animation logic
	//if u.HasBuff(buff.Burrowoed) {
	//i := (u.AnimationCount / 15) % 8
	//op = &ebiten.DrawImageOptions{}
	//op.GeoM.Translate(u.X-cs.X-float64(u.W/2), u.Y-cs.Y-float64(u.H/2))
	//img := imagesCache.Get(buffBurrowedKey)
	//screen.DrawImage(img.SubImage(image.Rect(i*32, 0, i*32+32, i*32+32)).(*ebiten.Image), op)
	//}
}
