package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
)

//go:embed assets/cyclope/Cyclopes.png
var Cyclopes_png []byte

var (
	unitImages = make(map[string]image.Image)
	unitGold   = 10
)

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

	Health float64

	Path []Step
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
		// We wait for the towers store as we need to interact with it
		us.GetDispatcher().WaitFor(us.game.Towers.GetDispatcherToken())
		var w, h float64 = 16, 16
		var x, y float64 = us.game.Map.GetRandomSpawnCoordinatesForLineID(act.SummonUnit.CurrentLineID)
		ustate.TotalUnits += 1
		u := &Unit{
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
			Health:        10,
		}
		ts := us.game.Towers.GetState().(TowersState)
		tws := make([]Object, 0, 0)
		for _, t := range ts.Towers {
			if t.LineID == u.CurrentLineID {
				tws = append(tws, t.Entity.Object)
			}
		}
		u.Path = us.astar(us.game.Map, u.CurrentLineID, u.MovingEntity, tws)
		ustate.Units[ustate.TotalUnits] = u
	case action.MoveUnit:
		// We wait for the towers store as we need to interact with it
		for _, u := range ustate.Units {
			if len(u.Path) > 0 {
				nextStep := u.Path[0]
				u.Path = u.Path[1:]
				u.MovingCount += 1
				u.Y = nextStep.Y
				u.X = nextStep.X
				u.Facing = nextStep.Facing
			}
		}
	case action.PlaceTower:
		// We wait for the towers store as we need to interact with it
		us.GetDispatcher().WaitFor(us.game.Towers.GetDispatcherToken())
		ts := us.game.Towers.GetState().(TowersState)
		for _, u := range ustate.Units {
			// Only need to recalculate path for each unit when the placed tower
			// is on the same LineID as the unit
			if u.CurrentLineID == act.PlaceTower.LineID {
				tws := make([]Object, 0, 0)
				for _, t := range ts.Towers {
					if t.LineID == u.CurrentLineID {
						tws = append(tws, t.Entity.Object)
					}
				}

				u.Path = us.astar(us.game.Map, u.CurrentLineID, u.MovingEntity, tws)
			}
		}
	case action.RemoveUnit:
		delete(ustate.Units, act.RemoveUnit.UnitID)
	case action.TowerAttack:
		u := ustate.Units[act.TowerAttack.UnitID]
		// For now the damage is just 1
		u.Health -= float64(towerDamage)
		if u.Health <= 0 {
			u.Health = 0
		}
	default:
	}
	return ustate
}

func (us *UnitsStore) Update() error {
	actionDispatcher.MoveUnit()

	for id, u := range us.GetState().(UnitsState).Units {
		if u.Health == 0 {
			actionDispatcher.UnitKilled(u.CurrentLineID, u.Type)
			actionDispatcher.RemoveUnit(id)
			continue
		}
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
	screen.DrawImage(u.Image.(*ebiten.Image).SubImage(image.Rect(sx, sy, sx+int(u.W), sy+int(u.H))).(*ebiten.Image), op)
}
