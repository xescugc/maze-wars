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
	"github.com/xescugc/maze-wars/utils"
)

type Unit struct {
	Type Type

	Stats

	Environment environment.Environment `json:"environment"`
	Abilities   []ability.Ability       `json:"abilities"`

	Keybind string

	Faceset image.Image
	Walk    image.Image
	Attack  image.Image
	Idle    image.Image
	Profile image.Image
}

type Stats struct {
	Health        float64 `json:"health"`
	Damage        float64 `json:"damage"`
	AttackSpeed   float64 `json:"attack_speed"`
	Shield        float64 `json:"shield"`
	Income        int     `json:"income"`
	Gold          int     `json:"gold"`
	MovementSpeed float64 `json:"movement_speed"`
}

func (u *Unit) FacesetKey() string { return fmt.Sprintf("u-f-%s", u.Type) }
func (u *Unit) WalkKey() string    { return fmt.Sprintf("u-w-%s", u.Type) }
func (u *Unit) AttackKey() string  { return fmt.Sprintf("u-a-%s", u.Type) }
func (u *Unit) IdleKey() string    { return fmt.Sprintf("u-i-%s", u.Type) }
func (u *Unit) ProfileKey() string { return fmt.Sprintf("u-p-%s", u.Type) }
func (u *Unit) HasAbility(a ability.Ability) bool {
	for _, ab := range u.Abilities {
		if a == ab {
			return true
		}
	}
	return false
}

func (u *Unit) Name() string {
	n, ok := names[u.Type]
	if !ok {
		return u.Type.String()
	}
	return n
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
		BlendMaster:   assets.BlendMasterWalk_png,
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
		BlendMaster:   assets.BlendMasterFaceset_png,
		Robot:         assets.RobotFaceset_png,
		MonkeyBoxer:   assets.MonkeyBoxerFaceset_png,
	}

	names = map[Type]string{
		Ninja:         "Ninja",
		Statue:        "Statue",
		Hunter:        "Hunter",
		Slime:         "Slime",
		Mole:          "Mole",
		SkeletonDemon: "Skeleton Demon",
		Butterfly:     "Butterfly",
		BlendMaster:   "Blend Master",
		Robot:         "Robot",
		MonkeyBoxer:   "Monkey Boxer",
	}
)

func init() {
	err := json.Unmarshal(assets.Units_json, &Units)
	if err != nil {
		log.Fatal(err)
	}

	profiles, _, err := image.Decode(bytes.NewReader(assets.UnitsProfile_png))
	if err != nil {
		log.Fatal(err)
	}

	pwh := 106
	for i, ty := range TypeValues() {
		u := Units[ty.String()]

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
		if u.HasAbility(ability.Attack) {
			ma, _, err := image.Decode(bytes.NewReader(assets.MonkeyBoxerAttack_png))
			if err != nil {
				log.Fatal(err)
			}
			u.Attack = ma

			mi, _, err := image.Decode(bytes.NewReader(assets.MonkeyBoxerIdle_png))
			if err != nil {
				log.Fatal(err)
			}
			u.Idle = mi
		}
		u.Profile = profiles.(utils.SubImager).SubImage(image.Rect(i*pwh, 0, i*pwh+pwh, pwh))

	}
}
