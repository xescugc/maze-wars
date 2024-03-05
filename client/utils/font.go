package utils

import (
	"log"

	"github.com/xescugc/maze-wars/assets"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	NormalFont font.Face
	SmallFont  font.Face
)

func init() {
	// Initialize Font
	tt, err := opentype.Parse(assets.Kongtext_ttf)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	NormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	SmallFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    16,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

}
