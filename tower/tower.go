package tower

import (
	"bytes"
	"fmt"
	"image"
	"log"

	"github.com/xescugc/maze-wars/assets"
	"github.com/xescugc/maze-wars/unit/environment"
	"github.com/xescugc/maze-wars/utils"
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

	// Idle is 32x32 used on the map
	Idle image.Image
	// Faceset is 38x38 used on the buttons
	Faceset image.Image
	// Profile is 106x106 used when selected
	Profile image.Image

	// The Update Cost is the Tower.Gold
	Updates []Type
}

func (t *Tower) FacesetKey() string { return fmt.Sprintf("t-f-%s", t.Type) }
func (t *Tower) IdleKey() string    { return fmt.Sprintf("t-i-%s", t.Type) }
func (t *Tower) ProfileKey() string { return fmt.Sprintf("t-p-%s", t.Type) }

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

func (t *Tower) Name() string {
	n, ok := names[t.Type]
	if !ok {
		return t.Type.String()
	}
	return n
}

func (t *Tower) Description() string {
	n, ok := descriptions[t.Type]
	if !ok {
		return t.Type.String()
	}
	return n
}

var (
	names = map[Type]string{
		Range1:       "Range - T1",
		Range2:       "Range - T2",
		RangeSingle1: "Range Singe - T3",
		RangeSingle2: "Range Singe - T4",
		RangeAoE1:    "Range AoE - T3",
		RangeAoE2:    "Range AoE - T4",

		Melee1:       "Melee - T1",
		Melee2:       "Melee - T2",
		MeleeSingle1: "Melee Single - T3",
		MeleeSingle2: "Melee Single - T4",
		MeleeAoE1:    "Melee AoE - T3",
		MeleeAoE2:    "Melee AoE - T4",
	}

	descriptions = map[Type]string{
		Range1:       `Basic range tower`,
		Range2:       `Updated basic range tower`,
		RangeSingle1: `Powerful single target range tower`,
		RangeSingle2: `More powerful single target range tower`,
		RangeAoE1:    `Ground AoE tower, slow but powerful`,
		RangeAoE2:    `Updated ground AoE tower, slow but powerful`,

		Melee1:       `Basic melee tower`,
		Melee2:       `Updated basic melee tower`,
		MeleeSingle1: `Improved single target melee tower`,
		MeleeSingle2: `Powerful single target melee tower`,
		MeleeAoE1:    `Flying AoE tower`,
		MeleeAoE2:    `Updated Flying AoE tower`,
	}

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
				RangeSingle1,
				RangeAoE1,
			},
		},
		RangeSingle1.String(): &Tower{
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
				RangeSingle2,
			},
		},
		RangeSingle2.String(): &Tower{
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
	face, _, err := image.Decode(bytes.NewReader(assets.TowersFacet_png))
	if err != nil {
		log.Fatal(err)
	}
	profiles, _, err := image.Decode(bytes.NewReader(assets.TowersProfile_png))
	if err != nil {
		log.Fatal(err)
	}

	wh := 32
	fwh := 38
	pwh := 106
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
		Towers[ty.String()].Idle = img.(utils.SubImager).SubImage(image.Rect(x*wh, y*wh, x*wh+wh, y*wh+wh))
		Towers[ty.String()].Faceset = face.(utils.SubImager).SubImage(image.Rect(x*fwh, y*fwh, x*fwh+fwh, y*fwh+fwh))
		Towers[ty.String()].Profile = profiles.(utils.SubImager).SubImage(image.Rect(x*pwh, y*pwh, x*pwh+pwh, y*pwh+pwh))
		Towers[ty.String()].Type = ty
		Towers[ty.String()].initTargets()
	}
}
