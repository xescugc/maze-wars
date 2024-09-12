package ability

import (
	"bytes"
	"fmt"
	"image"
	"log"

	"github.com/xescugc/maze-wars/assets"
	"github.com/xescugc/maze-wars/utils"
)

//go:generate enumer -type=Ability -transform=lower -json -transform=snake -output=ability_string.go

type Ability int

const (
	// Default abilities, order is the same as the asset
	Efficiency Ability = iota
	Tank
	Fast
	Split
	Burrow
	Resurrection
	Fly
	Camouflage
	Hybrid
	Attack
	// End of default abilities
)

var (
	Images map[string]image.Image

	descriptions = map[Ability]string{
		Efficiency: "Makes the unit cost efficient compared to the rest",
		Tank:       "Provides extra Health",
		Fast:       "Provides extra speed to the unit",
		Split: `Once it dies it'll split in 2 small
units (live of which is halved).`,
		Burrow: `Burrows once it reaches 1/2 life,
unburrows once the next wave reaches it.

This means that it'll be undergrond for 15s and
then the next unit that steps on it it'll unburrow,
or after 45s it'll unburrow itself.`,
		Resurrection: `Returns to life with 25% after
1.5 seconds when killed. Can only trigger once.
Next rank is 50% and next 75%.`,
		Fly: `Allows unit to go over the towers
and can only be targeted by  towers that
can attack flying units.`,
		Camouflage: `This unit won't draw attention from towers
and will always be attacked as a last priority`,
		Hybrid: `Gains life, shields, and movement speed base on
percent difference between attacker and defender`,
		Attack: `Instead of traversing the maze to reach the end
this units will attack the towers in order to destroy the
maze.`,
	}
	names = map[Ability]string{
		Efficiency:   "Efficiency",
		Tank:         "Tank",
		Fast:         "Fast",
		Split:        "Split",
		Burrow:       "Burrow",
		Resurrection: "Resurrection",
		Fly:          "Fly",
		Camouflage:   "Camouflage",
		Hybrid:       "Hybrid",
		Attack:       "Attack",
	}
)

func Key(a string) string          { return fmt.Sprintf("u-ab-%s", a) }
func Description(a Ability) string { return descriptions[a] }
func Name(a Ability) string        { return names[a] }
func init() {
	Images = make(map[string]image.Image)

	ubai, _, err := image.Decode(bytes.NewReader(assets.UnitsBasicAbilities_png))
	if err != nil {
		log.Fatal(err)
	}

	w := 38
	for i, a := range AbilityValues() {
		y := i / 5
		x := i % 5
		Images[a.String()] = ubai.(utils.SubImager).SubImage(image.Rect(x*w, y*w, x*w+w, y*w+w))
	}
}
