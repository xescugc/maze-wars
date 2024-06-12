package tower

import (
	"bytes"
	"fmt"
	"image"
	"log"

	"github.com/xescugc/maze-wars/assets"
	"github.com/xescugc/maze-wars/unit/environment"
)

type Tower struct {
	Type Type

	Damage float64
	Gold   int
	Health float64
	// Range is in a reduced version of 16 pixels, so Range == 1 == 16px
	Range       float64
	AttackSpeed float64
	// AoE is the same as the Range, AoE == 1 == 16px
	AoE       int
	AoEDamage float64

	Targets []environment.Environment `json:"targets"`
	targets map[environment.Environment]struct{}

	Faceset image.Image

	// The Update Cost is the Tower.Gold
	Updates []Type
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
	Towers = map[string]*Tower{
		Range1.String(): &Tower{
			Gold:        7,
			Damage:      2,
			Health:      25,
			AttackSpeed: 0.6,
			Range:       4,
			Targets: []environment.Environment{
				environment.Terrestrial,
				environment.Aerial,
			},
			Updates: []Type{
				Range2,
			},
		},
		Range2.String(): &Tower{
			Gold:        30,
			Damage:      7,
			Health:      50,
			AttackSpeed: 0.6,
			Range:       5,
			Targets: []environment.Environment{
				environment.Terrestrial,
				environment.Aerial,
			},
			Updates: []Type{
				RangeSingel1,
				RangeAoE1,
			},
		},
		RangeSingel1.String(): &Tower{
			Gold:        250,
			Damage:      40,
			Health:      75,
			AttackSpeed: 0.5,
			Range:       6,
			Targets: []environment.Environment{
				environment.Terrestrial,
				environment.Aerial,
			},
			Updates: []Type{
				RangeSingel2,
			},
		},
		RangeSingel2.String(): &Tower{
			Gold:        500,
			Damage:      80,
			Health:      125,
			AttackSpeed: 0.5,
			Range:       7,
			Targets: []environment.Environment{
				environment.Terrestrial,
				environment.Aerial,
			},
		},
		RangeAoE1.String(): &Tower{
			Gold:        250,
			Damage:      150,
			Health:      75,
			AttackSpeed: 2,
			Range:       5,
			AoE:         3,
			AoEDamage:   45,
			Targets: []environment.Environment{
				environment.Terrestrial,
			},
			Updates: []Type{
				RangeAoE2,
			},
		},
		RangeAoE2.String(): &Tower{
			Gold:        500,
			Damage:      180,
			Health:      125,
			AttackSpeed: 2,
			Range:       5,
			AoE:         3,
			AoEDamage:   54,
			Targets: []environment.Environment{
				environment.Terrestrial,
			},
		},
		Melee1.String(): &Tower{
			Gold:        7,
			Damage:      2,
			Health:      25,
			AttackSpeed: 0.3,
			Range:       1,
			Targets: []environment.Environment{
				environment.Terrestrial,
			},
			Updates: []Type{
				Melee2,
			},
		},
		Melee2.String(): &Tower{
			Gold:        30,
			Damage:      8,
			Health:      50,
			AttackSpeed: 0.3,
			Range:       1,
			Targets: []environment.Environment{
				environment.Terrestrial,
			},
			Updates: []Type{
				MeleeSingle1,
				MeleeAoE1,
			},
		},
		MeleeSingle1.String(): &Tower{
			Gold:        250,
			Damage:      50,
			Health:      75,
			AttackSpeed: 0.3,
			Range:       1,
			Targets: []environment.Environment{
				environment.Terrestrial,
			},
			Updates: []Type{
				MeleeSingle2,
			},
		},
		MeleeSingle2.String(): &Tower{
			Gold:        500,
			Damage:      100,
			Health:      125,
			AttackSpeed: 0.3,
			Range:       1,
			Targets: []environment.Environment{
				environment.Terrestrial,
			},
		},
		MeleeAoE1.String(): &Tower{
			Gold:        250,
			Damage:      100,
			Health:      75,
			AttackSpeed: 2,
			Range:       1,
			AoE:         3,
			AoEDamage:   30,
			Targets: []environment.Environment{
				environment.Terrestrial,
				environment.Aerial,
			},
			Updates: []Type{
				MeleeAoE2,
			},
		},
		MeleeAoE2.String(): &Tower{
			Gold:        500,
			Damage:      130,
			Health:      125,
			AttackSpeed: 2,
			Range:       1,
			AoE:         3,
			AoEDamage:   39,
			Targets: []environment.Environment{
				environment.Terrestrial,
				environment.Aerial,
			},
		},
	}

	FirstTowers = []*Tower{
		Towers[Range1.String()],
		Towers[Melee1.String()],
	}
)

func init() {
	img, _, err := image.Decode(bytes.NewReader(assets.Towers_png))
	if err != nil {
		log.Fatal(err)
	}

	wh := 32
	for i, ty := range TypeValues() {
		y := i / 6
		if y != 0 {
			y += 1
		}
		x := i % 6
		if x == 4 {
			y += 1
			x = 2
		} else if x == 5 {
			y += 1
			x = 3
		}
		Towers[ty.String()].Faceset = img.(SubImager).SubImage(image.Rect(x*wh, y*wh, x*wh+wh, y*wh+wh))
		Towers[ty.String()].Type = ty
		Towers[ty.String()].initTargets()
	}
}

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}
