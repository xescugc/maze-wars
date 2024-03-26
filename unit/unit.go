package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"log"

	"github.com/xescugc/maze-wars/assets"
)

type Unit struct {
	Type Type

	Health float64 `json:"health"`
	Income int     `json:"income"`
	Gold   int     `json:"gold"`

	Keybind string

	Faceset image.Image
	Sprite  image.Image
}

func (u *Unit) FacesetKey() string { return fmt.Sprintf("u-f-%s", u.Type) }
func (u *Unit) SpriteKey() string  { return fmt.Sprintf("u-s-%s", u.Type) }

var (
	Units map[string]*Unit

	sprites = map[Type][]byte{
		Spirit:    assets.SpiritSprite_png,
		Flam:      assets.FlamSprite_png,
		Raccon:    assets.RacoonSprite_png,
		Cyclope:   assets.CyclopeSprite_png,
		Eye:       assets.EyeSprite_png,
		Beast:     assets.BeastSprite_png,
		Butterfly: assets.ButterflySprite_png,
		Mole:      assets.MoleSprite_png,
		Skull:     assets.SkullSprite_png,
		Snake:     assets.SnakeSprite_png,
	}

	facesets = map[Type][]byte{
		Spirit:    assets.SpiritFaceset_png,
		Flam:      assets.FlamFaceset_png,
		Raccon:    assets.RacoonFaceset_png,
		Cyclope:   assets.CyclopeFaceset_png,
		Eye:       assets.EyeFaceset_png,
		Beast:     assets.BeastFaceset_png,
		Butterfly: assets.ButterflyFaceset_png,
		Mole:      assets.MoleFaceset_png,
		Skull:     assets.SkullFaceset_png,
		Snake:     assets.SnakeFaceset_png,
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

		sb, ok := sprites[ty]
		if !ok {
			log.Fatalf("Type %s does not have an sprite assigned", ty)
		}
		si, _, err := image.Decode(bytes.NewReader(sb))
		if err != nil {
			log.Fatal(err)
		}

		u.Sprite = si
		u.Faceset = fi
		u.Type = ty
	}
}
