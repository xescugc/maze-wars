package unit

import (
	"bytes"
	"encoding/json"
	"image"
	"log"

	"github.com/xescugc/maze-wars/assets"
)

type Unit struct {
	Type Type

	Health float64 `json:"health"`
	Income int     `json:"income"`
	Gold   int     `json:"gold"`

	Faceset image.Image
	Sprite  image.Image
}

var (
	Units map[string]*Unit

	sprites = map[Type][]byte{
		Spirit:     assets.Spirit_png,
		Spirit2:    assets.Spirit2_png,
		Flam2:      assets.Flam2_png,
		Flam:       assets.Flam_png,
		Octopus:    assets.Octopus_png,
		Octopus2:   assets.Octopus2_png,
		Raccon:     assets.Racoon_png,
		GoldRacoon: assets.GoldRacoon_png,
		Cyclope2:   assets.Cyclope2_png,
		Cyclope:    assets.Cyclope_png,
	}

	facesets = map[Type][]byte{
		Spirit:     assets.SpiritFaceset_png,
		Spirit2:    assets.Spirit2Faceset_png,
		Flam2:      assets.Flam2Faceset_png,
		Flam:       assets.FlamFaceset_png,
		Octopus:    assets.OctopusFaceset_png,
		Octopus2:   assets.Octopus2Faceset_png,
		Raccon:     assets.RacoonFaceset_png,
		GoldRacoon: assets.GoldRacoonFaceset_png,
		Cyclope2:   assets.Cyclope2Faceset_png,
		Cyclope:    assets.CyclopeFaceset_png,
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
