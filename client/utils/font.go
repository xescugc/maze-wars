package utils

import (
	"bytes"
	"log"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/xescugc/maze-wars/assets"
)

var (
	NormalFont text.Face
	SmallFont  text.Face
	Font80     text.Face
	Font60     text.Face
	Font40     text.Face
	Font10     text.Face

	SFont20 text.Face
)

func init() {
	tt, err := text.NewGoTextFaceSource(bytes.NewReader(assets.Munro_ttf))
	if err != nil {
		log.Fatal(err)
	}

	stt, err := text.NewGoTextFaceSource(bytes.NewReader(assets.MunroSmall_ttf))
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	Font80 = &text.GoTextFace{
		Source: tt,
		Size:   80,
	}
	Font60 = &text.GoTextFace{
		Source: tt,
		Size:   60,
	}
	Font40 = &text.GoTextFace{
		Source: tt,
		Size:   40,
	}
	NormalFont = &text.GoTextFace{
		Source: tt,
		Size:   30,
	}
	SmallFont = &text.GoTextFace{
		Source: tt,
		Size:   20,
	}
	Font10 = &text.GoTextFace{
		Source: tt,
		Size:   10,
	}

	SFont20 = &text.GoTextFace{
		Source: stt,
		Size:   20,
	}

}
