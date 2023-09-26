package tower

import (
	"bytes"
	"encoding/json"
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/ltw/assets"
)

type Tower struct {
	Range  float64 `json:"range"`
	Damage float64 `json:"damage"`
	Gold   int     `json:"gold"`

	Image image.Image
}

var (
	Towers map[string]*Tower
)

func init() {
	err := json.Unmarshal(assets.Towers_json, &Towers)
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
			i, _, err := image.Decode(bytes.NewReader(assets.TilesetHouse_png))
			if err != nil {
				log.Fatal(err)
			}

			tw.Image = ebiten.NewImageFromImage(i).SubImage(image.Rect(5*16, 17*16, 5*16+16*2, 17*16+16*2))
		default:
			log.Fatalf("failed to load tower %q\n", t)
		}
	}
}
