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

var (
	towerImages         = make(map[string]image.Image)
	towerRange  float64 = 16 * 2
	towerDamage         = 1
)

func init() {
	si, _, err := image.Decode(bytes.NewReader(TilesetHouse_png))
	if err != nil {
		log.Fatal(err)
	}

	towerImages["soldier"] = ebiten.NewImageFromImage(si).SubImage(image.Rect(5*16, 17*16, 5*16+16*2, 17*16+16*2))
}

type TowersStore struct {
	*flux.ReduceStore

	game *Game
}

type TowersState struct {
	Towers      map[int]*Tower
	TotalTowers int
}

type Tower struct {
	Entity

	Type   string
	LineID int
}

func NewTowersStore(d *flux.Dispatcher, g *Game) *TowersStore {
	ts := &TowersStore{
		game: g,
	}
	ts.ReduceStore = flux.NewReduceStore(d, ts.Reduce, TowersState{Towers: make(map[int]*Tower)})

	return ts
}

func (ts *TowersStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	ustate, ok := state.(TowersState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.PlaceTower:
		var w, h float64 = 16 * 2, 16 * 2
		ustate.TotalTowers += 1
		ustate.Towers[ustate.TotalTowers] = &Tower{
			Entity: Entity{
				Object: Object{
					X: float64(act.PlaceTower.X), Y: float64(act.PlaceTower.Y),
					W: w, H: h,
				},
				Image: towerImages[act.PlaceTower.Type],
			},
			Type:   act.PlaceTower.Type,
			LineID: act.PlaceTower.LineID,
		}
	default:
	}
	return ustate
}

func (ts *TowersStore) Update() error {
	uts := ts.game.Units.GetState().(UnitsState).Units
	tws := ts.GetState().(TowersState).Towers
	for uid, u := range uts {
		for _, t := range tws {
			if t.Distance(u.Object) <= towerRange {
				actionDispatcher.TowerAttack(uid)
			}
		}
	}
	return nil
}

func (ts *TowersStore) Draw(screen *ebiten.Image) {
	for _, t := range ts.GetState().(TowersState).Towers {
		t.Draw(screen, ts.game.Camera)
	}
}

func (t *Tower) Draw(screen *ebiten.Image, c *CameraStore) {
	cs := c.GetState().(CameraState)
	if !t.IsColliding(cs.Object) {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(t.X-cs.X, t.Y-cs.Y)
	screen.DrawImage(t.Image.(*ebiten.Image), op)
}
