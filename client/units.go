package client

import (
	"bytes"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/ltw/assets"
	"github.com/xescugc/ltw/store"
	"github.com/xescugc/ltw/unit"
)

type Units struct {
	game *Game

	lifeBarProgress image.Image
	lifeBarUnder    image.Image
}

var (
	facingToTile = map[ebiten.Key]int{
		ebiten.KeyS: 0,
		ebiten.KeyW: 1,
		ebiten.KeyA: 2,
		ebiten.KeyD: 3,
	}
)

func NewUnits(g *Game) (*Units, error) {
	lbpi, _, err := image.Decode(bytes.NewReader(assets.LifeBarMiniProgress_png))
	if err != nil {
		return nil, err
	}

	lbui, _, err := image.Decode(bytes.NewReader(assets.LifeBarMiniUnder_png))
	if err != nil {
		return nil, err
	}

	us := &Units{
		game:            g,
		lifeBarProgress: ebiten.NewImageFromImage(lbpi),
		lifeBarUnder:    ebiten.NewImageFromImage(lbui),
	}

	return us, nil
}

func (us *Units) Update() error {
	actionDispatcher.MoveUnit()
	cp := us.game.Store.Players.GetCurrentPlayer()

	for _, u := range us.game.Store.Units.GetUnits() {
		// Only do the events as the owern of the unit if not the actionDispatcher
		// will also dispatch it to the server and the event will be done len(players)
		// amount of times
		if cp.ID == u.PlayerID {
			if u.Health == 0 {
				p := us.game.Store.Players.GetByLineID(u.CurrentLineID)
				actionDispatcher.UnitKilled(p.ID, u.Type)
				actionDispatcher.RemoveUnit(u.ID)
				continue
			}
			if us.game.Store.Map.IsAtTheEnd(u.Object, u.CurrentLineID) {
				p := us.game.Store.Players.GetByLineID(u.CurrentLineID)
				actionDispatcher.StealLive(p.ID, u.PlayerID)
				nlid := us.game.Store.Map.GetNextLineID(u.CurrentLineID)
				if nlid == u.PlayerLineID {
					actionDispatcher.RemoveUnit(u.ID)
				} else {
					// TODO: Send to next line
					// this will need to be done once
					// we add more than 2 players
				}
			}
		}
	}

	return nil
}

func (us *Units) Draw(screen *ebiten.Image) {
	for _, u := range us.game.Store.Units.GetUnits() {
		us.DrawUnit(screen, us.game.Camera, u)
	}
}

func (us *Units) DrawUnit(screen *ebiten.Image, c *CameraStore, u *store.Unit) {
	cs := c.GetState().(CameraState)
	for _, s := range u.Path {
		screen.Set(int(s.X-cs.X), int(s.Y-cs.Y), color.Black)
	}
	if !u.IsColliding(cs.Object) {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(u.X-cs.X, u.Y-cs.Y)
	op.GeoM.Scale(cs.Zoom, cs.Zoom)
	sx := facingToTile[u.Facing] * int(u.W)
	i := (u.MovingCount / 5) % 4
	sy := i * int(u.H)
	screen.DrawImage(u.Sprite().(*ebiten.Image).SubImage(image.Rect(sx, sy, sx+int(u.W), sy+int(u.H))).(*ebiten.Image), op)

	// Only draw the Health bar if the unit has been hit
	h := unit.Units[u.Type].Health
	if unit.Units[u.Type].Health != u.Health {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(u.X-cs.X, u.Y-cs.Y-float64(us.lifeBarUnder.Bounds().Dy()))
		screen.DrawImage(us.lifeBarUnder.(*ebiten.Image), op)

		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(u.X-cs.X, u.Y-cs.Y-float64(us.lifeBarProgress.Bounds().Dy()))
		screen.DrawImage(us.lifeBarProgress.(*ebiten.Image).SubImage(image.Rect(0, 0, int(float64(us.lifeBarProgress.Bounds().Dx())*(u.Health/h)), us.lifeBarProgress.Bounds().Dy())).(*ebiten.Image), op)
	}
}
