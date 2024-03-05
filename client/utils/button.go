package utils

import (
	"image/color"

	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

var (
	ButtonImage = loadButtonImage()
)

func loadButtonImage() *widget.ButtonImage {
	idle := euiimage.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 180, A: 255})

	hover := euiimage.NewNineSliceColor(color.NRGBA{R: 130, G: 130, B: 150, A: 255})

	pressed := euiimage.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 120, A: 255})

	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: pressed,
	}
}

func ButtonImageFromImage(i *ebiten.Image) *widget.ButtonImage {
	nsi := euiimage.NewNineSliceSimple(i, i.Bounds().Dx(), i.Bounds().Dy())

	cm := colorm.ColorM{}
	cm.Scale(2, 0.5, 0.5, 0.9)

	ni := ebiten.NewImage(i.Bounds().Dx(), i.Bounds().Dy())
	colorm.DrawImage(ni, i, cm, nil)
	dsi := euiimage.NewNineSliceSimple(ni, ni.Bounds().Dx(), ni.Bounds().Dy())
	return &widget.ButtonImage{
		Idle:     nsi,
		Hover:    nsi,
		Pressed:  nsi,
		Disabled: dsi,
	}
}
