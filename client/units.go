package main

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/ltw/store"
)

type Units struct {
	game *Game
}

var (
	facingToTile = map[ebiten.Key]int{
		ebiten.KeyS: 0,
		ebiten.KeyW: 1,
		ebiten.KeyA: 2,
		ebiten.KeyD: 3,
	}
)

func NewUnits(g *Game) *Units {
	us := &Units{
		game: g,
	}

	return us
}

func (us *Units) Update() error {
	actionDispatcher.MoveUnit()
	cp := us.game.Players.GetCurrentPlayer()

	for id, u := range us.game.Store.Units.GetState().(store.UnitsState).Units {
		// Only do the events as the owern of the unit if not the actionDispatcher
		// will also dispatch it to the server and the event will be done len(players)
		// amount of times
		if cp.ID == u.PlayerID {
			if u.Health == 0 {
				p := us.game.Store.Players.GetByLineID(u.CurrentLineID)
				actionDispatcher.UnitKilled(p.ID, u.Type)
				actionDispatcher.RemoveUnit(id)
				continue
			}
			if us.game.Map.IsAtTheEnd(u.Object, u.CurrentLineID) {
				p := us.game.Store.Players.GetByLineID(u.CurrentLineID)
				actionDispatcher.StealLive(p.ID, u.PlayerID)
				nlid := us.game.Map.GetNextLineID(u.CurrentLineID)
				if nlid == u.PlayerLineID {
					actionDispatcher.RemoveUnit(id)
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
	for _, u := range us.game.Store.Units.GetState().(store.UnitsState).Units {
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
	screen.DrawImage(u.Image().(*ebiten.Image).SubImage(image.Rect(sx, sy, sx+int(u.W), sy+int(u.H))).(*ebiten.Image), op)
}
