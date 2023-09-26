package unit

import (
	"bytes"
	"encoding/json"
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/ltw/assets"
)

type Unit struct {
	Health float64 `json:"health"`
	Income int     `json:"income"`
	Gold   int     `json:"gold"`

	Image image.Image
}

var (
	Units map[string]*Unit
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
		switch ty {
		case Cyclope:
			i, _, err := image.Decode(bytes.NewReader(assets.Cyclopes_png))
			if err != nil {
				log.Fatal(err)
			}

			u.Image = ebiten.NewImageFromImage(i)
		default:
			log.Fatalf("failed to load unit %q\n", t)
		}
	}
}
