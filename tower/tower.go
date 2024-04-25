package tower

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"log"
	"math"

	"github.com/xescugc/maze-wars/assets"
	"github.com/xescugc/maze-wars/unit/environment"
)

type Tower struct {
	Type Type

	Stats

	Gold  int     `json:"gold"`
	Range float64 `json:"range"`

	Targets []environment.Environment `json:"targets"`
	targets map[environment.Environment]struct{}

	Keybind string

	Faceset image.Image
}

type Stats struct {
	Damage float64 `json:"damage"`
}

type Update struct {
	UpdateCost int
	Stats
}

func (t *Tower) FacesetKey() string { return fmt.Sprintf("t-f-%s", t.Type) }

func (t *Tower) CanTarget(env environment.Environment) bool {
	_, ok := t.targets[env]
	return ok
}

// initTargets will map the Targets to a map for easy access
func (t *Tower) initTargets() {
	t.targets = make(map[environment.Environment]struct{})
	for _, tg := range t.Targets {
		t.targets[tg] = struct{}{}
	}
}

var (
	Towers map[string]*Tower

	towerUpdates = make(map[string]map[int]Update)

	updateCostBase   = 100
	updateCostFactor = 0.5

	updateStateDamageFactor = 1.5
)

func init() {
	err := json.Unmarshal(assets.Towers_json, &Towers)
	if err != nil {
		log.Fatal(err)
	}

	i, _, err := image.Decode(bytes.NewReader(assets.TilesetHouse_png))
	if err != nil {
		log.Fatal(err)
	}

	for t, tw := range Towers {
		ty, err := TypeString(t)
		if err != nil {
			log.Fatal(err)
		}
		switch ty {
		case Soldier:
			tw.Faceset = i.(SubImager).SubImage(image.Rect(5*16, 17*16, 5*16+16*2, 17*16+16*2))
		case Monk:
			tw.Faceset = i.(SubImager).SubImage(image.Rect(5*16, 15*16, 5*16+16*2, 15*16+16*2))
		default:
			log.Fatalf("failed to load tower %q\n", t)
		}
		tw.Type = ty
		tw.initTargets()

		towerUpdates[t] = make(map[int]Update)
		for i := 2; i < 5; i++ {
			towerUpdates[t][i] = Update{
				UpdateCost: int(float64(updateCostBase) * updateCostFactor * math.Pow(2, float64((i-1)))),
				Stats: Stats{
					Damage: levelToValue(i, tw.Damage),
				},
			}
		}
	}
}

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func FindUpdateByLevel(t string, lvl int) *Update {
	tu, ok := towerUpdates[t][lvl]
	if !ok {
		return nil
	}
	return &tu
}

func levelToValue(lvl int, base float64) float64 {
	for i := 1; i < lvl; i++ {
		base = base * updateStateDamageFactor
	}
	return base
}
