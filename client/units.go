package main

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
)

//go:embed assets/cyclope/Cyclopes.png
var Cyclopes_png []byte

var unitImages = make(map[string]image.Image)

func init() {
	ci, _, err := image.Decode(bytes.NewReader(Cyclopes_png))
	if err != nil {
		log.Fatal(err)
	}

	unitImages["cyclope"] = ebiten.NewImageFromImage(ci)
}

type UnitsStore struct {
	*flux.ReduceStore

	game *Game
}

type UnitsState struct {
	Units      map[int]*Unit
	TotalUnits int
}

type Unit struct {
	MovingEntity

	Type          string
	PlayerID      int
	PlayerLineID  int
	CurrentLineID int
}

var (
	facingToTile = map[ebiten.Key]int{
		ebiten.KeyS: 0,
		ebiten.KeyW: 1,
		ebiten.KeyA: 2,
		ebiten.KeyD: 3,
	}
)

func NewUnitsStore(d *flux.Dispatcher, g *Game) *UnitsStore {
	us := &UnitsStore{
		game: g,
	}
	us.ReduceStore = flux.NewReduceStore(d, us.Reduce, UnitsState{Units: make(map[int]*Unit)})

	return us
}

func (us *UnitsStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	ustate, ok := state.(UnitsState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.SummonUnit:
		var w, h float64 = 16, 16
		var x, y float64 = us.game.Map.GetRandomSpawnCoordinatesForLineID(act.SummonUnit.CurrentLineID)
		ustate.TotalUnits += 1
		ustate.Units[ustate.TotalUnits] = &Unit{
			MovingEntity: MovingEntity{
				Entity: Entity{
					Object: Object{
						X: x, Y: y,
						W: w, H: h,
					},
					Image: unitImages[act.SummonUnit.Type],
				},
				Facing: ebiten.KeyS,
			},
			Type:          act.SummonUnit.Type,
			PlayerID:      act.SummonUnit.PlayerID,
			PlayerLineID:  act.SummonUnit.PlayerLineID,
			CurrentLineID: act.SummonUnit.CurrentLineID,
		}
	case action.MoveUnit:
		for _, u := range ustate.Units {
			u.MovingCount += 1
			u.Y += 1
		}
	case action.RemoveUnit:
		delete(ustate.Units, act.RemoveUnit.UnitID)
	default:
	}
	return ustate
}

func (us *UnitsStore) Update() error {
	actionDispatcher.MoveUnit()

	for id, u := range us.GetState().(UnitsState).Units {
		if us.game.Map.IsAtTheEnd(u.Object, u.CurrentLineID) {
			p := us.game.Players.GetByLineID(u.CurrentLineID)
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

	return nil
}

func (us *UnitsStore) Draw(screen *ebiten.Image) {
	for _, u := range us.GetState().(UnitsState).Units {
		u.Draw(screen, us.game.Camera)
	}
}

func (u *Unit) Draw(screen *ebiten.Image, c *CameraStore) {
	cs := c.GetState().(CameraState)
	if !u.IsColliding(cs.Object) {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(u.X-cs.X, u.Y-cs.Y)
	sx := facingToTile[u.Facing] * int(u.W)
	i := (u.MovingCount / 5) % 4
	sy := i * int(u.H)
	screen.DrawImage(u.Image.(*ebiten.Image).SubImage(image.Rect(sx, sy, sx+int(u.W), sy+int(u.H))).(*ebiten.Image), op)
}
