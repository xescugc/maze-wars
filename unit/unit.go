package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"log"

	"github.com/xescugc/maze-wars/assets"
	"github.com/xescugc/maze-wars/unit/ability"
	"github.com/xescugc/maze-wars/unit/environment"
)

type Unit struct {
	Type Type

	Stats

	Environment environment.Environment `json:"environment"`
	Abilities   []ability.Ability       `json:"abilities"`

	Keybind string

	Faceset image.Image
	Walk    image.Image
}

type Stats struct {
	Health        float64 `json:"health"`
	Shield        float64 `json:"shield"`
	Income        int     `json:"income"`
	Gold          int     `json:"gold"`
	MovementSpeed float64 `json:"movement_speed"`
}

func (u *Unit) FacesetKey() string { return fmt.Sprintf("u-f-%s", u.Type) }
func (u *Unit) WalkKey() string    { return fmt.Sprintf("u-w-%s", u.Type) }
func (u *Unit) HasAbility(a ability.Ability) bool {
	for _, ab := range u.Abilities {
		if a == ab {
			return true
		}
	}
	return false
}

var (
	Units map[string]*Unit

	walks = map[Type][]byte{
		Ninja:         assets.NinjaWalk_png,
		Statue:        assets.StatueWalk_png,
		Hunter:        assets.HunterWalk_png,
		Slime:         assets.SlimeWalk_png,
		Mole:          assets.MoleWalk_png,
		SkeletonDemon: assets.SkeletonDemonWalk_png,
		Butterfly:     assets.ButterflyWalk_png,
		NinjaMasked:   assets.NinjaMaskedWalk_png,
		Robot:         assets.RobotWalk_png,
		MonkeyBoxer:   assets.MonkeyBoxerWalk_png,
	}

	facesets = map[Type][]byte{
		Ninja:         assets.NinjaFaceset_png,
		Statue:        assets.StatueFaceset_png,
		Hunter:        assets.HunterFaceset_png,
		Slime:         assets.SlimeFaceset_png,
		Mole:          assets.MoleFaceset_png,
		SkeletonDemon: assets.SkeletonDemonFaceset_png,
		Butterfly:     assets.ButterflyFaceset_png,
		NinjaMasked:   assets.NinjaMaskedFaceset_png,
		Robot:         assets.RobotFaceset_png,
		MonkeyBoxer:   assets.MonkeyBoxerFaceset_png,
	}
)

func init() {
	err := json.Unmarshal(assets.Units_json, &Units)
	if err != nil {
		log.Fatal(err)
	}

	for t, u := range Units {
		ty, err := TypeString(t)
		if err != nil {
			log.Fatal(err)
		}

		fb, ok := facesets[ty]
		if !ok {
			log.Fatalf("Type %s does not have an faceset assigned", ty)
		}
		fi, _, err := image.Decode(bytes.NewReader(fb))
		if err != nil {
			log.Fatal(err)
		}

		wb, ok := walks[ty]
		if !ok {
			log.Fatalf("Type %s does not have an walk assigned", ty)
		}
		wi, _, err := image.Decode(bytes.NewReader(wb))
		if err != nil {
			log.Fatal(err)
		}

		u.Walk = wi
		u.Faceset = fi
		u.Type = ty
		u.Stats.Income = u.Stats.Gold / 5
	}
}
