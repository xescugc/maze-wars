package store

import (
	"bytes"
	"image"
	"log"

	"github.com/gofrs/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
	"github.com/xescugc/ltw/assets"
	"github.com/xescugc/ltw/utils"
)

var (
	towerImages = make(map[string]image.Image)
	towerDamage = 1
)

func init() {
	si, _, err := image.Decode(bytes.NewReader(assets.TilesetHouse_png))
	if err != nil {
		log.Fatal(err)
	}

	towerImages["soldier"] = ebiten.NewImageFromImage(si).SubImage(image.Rect(5*16, 17*16, 5*16+16*2, 17*16+16*2))
}

type Towers struct {
	*flux.ReduceStore

	store *Store
}

type TowersState struct {
	Towers map[string]*Tower
}

type Tower struct {
	utils.Object

	Type   string
	LineID int
}

func (t *Tower) Image() image.Image {
	return towerImages[t.Type]
}

func NewTowers(d *flux.Dispatcher, s *Store) *Towers {
	t := &Towers{
		store: s,
	}

	t.ReduceStore = flux.NewReduceStore(d, t.Reduce, TowersState{
		Towers: make(map[string]*Tower),
	})

	return t
}

func (ts *Towers) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	tstate, ok := state.(TowersState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.PlaceTower:
		p := ts.store.Players.GetPlayerByID(act.PlaceTower.PlayerID)

		var w, h float64 = 16 * 2, 16 * 2
		tid := uuid.Must(uuid.NewV4())
		tstate.Towers[tid.String()] = &Tower{
			Object: utils.Object{
				X: float64(act.PlaceTower.X), Y: float64(act.PlaceTower.Y),
				W: w, H: h,
			},
			Type:   act.PlaceTower.Type,
			LineID: p.LineID,
		}
	case action.UpdateState:
		for id, t := range act.UpdateState.Towers.Towers {
			nt := Tower(*t)
			tstate.Towers[id] = &nt
		}
	default:
	}
	return tstate
}
