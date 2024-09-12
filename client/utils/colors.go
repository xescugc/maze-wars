package utils

import (
	"image/color"

	"github.com/ebitenui/ebitenui/widget"
)

var (
	Green = color.RGBA{116, 163, 52, 255}
	//LightGreen  = color.RGBA{173, 188, 58, 255}
	Red         = color.RGBA{209, 75, 52, 255}
	Transparent = color.RGBA{0, 0, 0, 0}
	Black       = color.RGBA{0, 0, 0, 255}
	BlackT      = color.RGBA{0, 0, 0, 120}

	TableBlack = color.RGBA{19, 27, 27, 255}

	TextColor = color.RGBA{223, 223, 223, 255}

	GreenTextColor = color.RGBA{116, 163, 52, 255}

	ButtonTextIdleColor     = TextColor
	ButtonTextPressedColor  = color.RGBA{254, 173, 84, 255}
	ButtonTextDisabledColor = color.RGBA{169, 157, 142, 255}
	ButtonTextHoverColor    = color.RGBA{242, 242, 242, 255}

	ButtonTextCancelPressedColor = color.RGBA{238, 207, 155, 255}

	White = TextColor
)

func TextInputColor() *widget.TextInputColor {
	return &widget.TextInputColor{
		Idle:          ButtonTextIdleColor,
		Disabled:      ButtonTextDisabledColor,
		Caret:         ButtonTextIdleColor,
		DisabledCaret: ButtonTextDisabledColor,
	}
}
