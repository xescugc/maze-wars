package main

import (
	"bytes"
	"image"

	_ "embed"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/maps/1v1.png
var M_1v1_png []byte

func init() {
}

type Map struct {
	Image image.Image
}

func NewMap() (*Map, error) {
	mi, _, err := image.Decode(bytes.NewReader(M_1v1_png))
	if err != nil {
		return nil, err
	}

	return &Map{
		Image: ebiten.NewImageFromImage(mi),
	}, nil
}

func (m *Map) GetX() int { return m.Image.Bounds().Dx() }
func (m *Map) GetY() int { return m.Image.Bounds().Dy() }
