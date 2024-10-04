package utils

import (
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

var (
	ButtonImage = loadButtonImage()

	isPressed = true
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

func TextInputResource() *widget.TextInputImage {
	ii := LoadImageNineSlice(DarkInputBGKey, 12, 4, !isPressed)

	return &widget.TextInputImage{
		Idle:     ii,
		Disabled: ii,
	}
}

func ButtonResource() *widget.ButtonImage {
	bp := LoadImageNineSlice(ButtonPressedKey, 12, 4, isPressed)
	bn := LoadImageNineSlice(ButtonNormalKey, 12, 4, !isPressed)
	bh := LoadImageNineSlice(ButtonHoverKey, 12, 4, !isPressed)
	bd := LoadImageNineSlice(ButtonDisabledKey, 12, 4, !isPressed)

	return &widget.ButtonImage{
		Idle:     bn,
		Hover:    bh,
		Pressed:  bp,
		Disabled: bd,
	}
}

func ButtonBorderResource() *widget.ButtonImage {
	bp := LoadImageNineSlice(ButtonBorderPressedKey, 2, 2, !isPressed)
	bn := LoadImageNineSlice(ButtonBorderNormalKey, 2, 2, !isPressed)
	bh := LoadImageNineSlice(ButtonBorderHoverKey, 2, 2, !isPressed)
	bd := LoadImageNineSlice(ButtonBorderDisabledKey, 2, 2, !isPressed)

	return &widget.ButtonImage{
		Idle:     bn,
		Hover:    bh,
		Pressed:  bp,
		Disabled: bd,
	}
}

func BigButtonResource() *widget.ButtonImage {
	bp := LoadImageNineSlice(BigButtonPressedKey, 16, 16, isPressed)
	bn := LoadImageNineSlice(BigButtonNormalKey, 16, 16, !isPressed)
	bh := LoadImageNineSlice(BigButtonHoverKey, 16, 16, !isPressed)
	bd := LoadImageNineSlice(BigButtonDisabledKey, 16, 16, !isPressed)

	return &widget.ButtonImage{
		Idle:     bn,
		Hover:    bh,
		Pressed:  bp,
		Disabled: bd,
	}
}

func BigCancelButtonResource() *widget.ButtonImage {
	bp := LoadImageNineSlice(BigCancelButtonPressedKey, 16, 16, isPressed)
	bn := LoadImageNineSlice(BigCancelButtonNormalKey, 16, 16, !isPressed)
	bh := LoadImageNineSlice(BigCancelButtonHoverKey, 16, 16, !isPressed)

	return &widget.ButtonImage{
		Idle:    bn,
		Hover:   bh,
		Pressed: bp,
	}
}

func BigLeftTabButtonResource() *widget.ButtonImage {
	bp := LoadImageNineSlice(BigButtonPressedLeftTabKey, 16, 16, isPressed)
	bn := LoadImageNineSlice(BigButtonNormalLeftTabKey, 16, 16, !isPressed)
	bh := LoadImageNineSlice(BigButtonHoverLeftTabKey, 16, 16, !isPressed)
	bd := LoadImageNineSlice(BigButtonDisabledLeftTabKey, 16, 16, !isPressed)

	return &widget.ButtonImage{
		Idle:     bn,
		Hover:    bh,
		Pressed:  bp,
		Disabled: bd,
	}
}

func BigRightTabButtonResource() *widget.ButtonImage {
	bp := LoadImageNineSlice(BigButtonPressedRightTabKey, 16, 16, isPressed)
	bn := LoadImageNineSlice(BigButtonNormalRightTabKey, 16, 16, !isPressed)
	bh := LoadImageNineSlice(BigButtonHoverRightTabKey, 16, 16, !isPressed)
	bd := LoadImageNineSlice(BigButtonDisabledRightTabKey, 16, 16, !isPressed)

	return &widget.ButtonImage{
		Idle:     bn,
		Hover:    bh,
		Pressed:  bp,
		Disabled: bd,
	}
}

func CheckboxButtonResource() *widget.ButtonImage {
	bp := LoadImageNineSlice(CheckboxCheckedKey, 44, 44, !isPressed)
	bn := LoadImageNineSlice(CheckboxUncheckedKey, 44, 44, !isPressed)

	return &widget.ButtonImage{
		Idle:    bn,
		Pressed: bp,
	}
}

func LogoButtonResource() *widget.ButtonImage {
	return ButtonImageFromImage(Images.Get(LogoKey))
}

func TextButtonResource() *widget.ButtonImage {
	idle := euiimage.NewNineSliceColor(color.NRGBA{R: 0, G: 0, B: 0, A: 0})

	bp := LoadImageNineSlice(MenuButtonPressedKey, 16, 16, !isPressed)
	bh := LoadImageNineSlice(MenuButtonHoverKey, 16, 16, !isPressed)

	return &widget.ButtonImage{
		Idle:     idle,
		Hover:    bh,
		Pressed:  bp,
		Disabled: idle,
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

func DisableImage(i *ebiten.Image) *ebiten.Image {
	cm := colorm.ColorM{}
	cm.Scale(2, 0.5, 0.5, 0.9)

	ni := ebiten.NewImage(i.Bounds().Dx(), i.Bounds().Dy())
	colorm.DrawImage(ni, i, cm, nil)
	return ni
}

func ButtonImageFromKey(k string, cw, ch int) *widget.ButtonImage {
	img := LoadImageNineSlice(k, cw, ch, !isPressed)
	return &widget.ButtonImage{
		Idle:     img,
		Hover:    img,
		Pressed:  img,
		Disabled: img,
	}
}

func ImageToNineSlice(k string) *image.NineSlice {
	return LoadImageNineSlice(k, 0, 0, !isPressed)
}

var (
	ButtonTextColor = widget.ButtonTextColor{
		Idle:     ButtonTextIdleColor,
		Disabled: ButtonTextDisabledColor,
		Hover:    ButtonTextHoverColor,
		Pressed:  ButtonTextPressedColor,
	}
	ButtonTextCancelColor = widget.ButtonTextColor{
		Idle:     ButtonTextIdleColor,
		Disabled: ButtonTextDisabledColor,
		Hover:    ButtonTextHoverColor,
		Pressed:  ButtonTextCancelPressedColor,
	}
	ButtonTextColorNoPressed = widget.ButtonTextColor{
		Idle:     ButtonTextIdleColor,
		Disabled: ButtonTextDisabledColor,
		Hover:    ButtonTextHoverColor,
	}

	//Move the text down and right on press
	ButtonOptsPressedText = widget.ButtonOpts.PressedHandler(func(args *widget.ButtonPressedEventArgs) {
		args.Button.Text().Inset.Top = 2
		args.Button.GetWidget().CustomData = true
	})
	//Move the text back to start on press released
	ButtonOptsReleasedText = widget.ButtonOpts.ReleasedHandler(func(args *widget.ButtonReleasedEventArgs) {
		if args.Button.State() == widget.WidgetChecked {
			return
		}
		args.Button.Text().Inset.Top = 0
		args.Button.GetWidget().CustomData = false
	})
	ButtonOptsCursorEnteredText = widget.ButtonOpts.CursorEnteredHandler(func(args *widget.ButtonHoverEventArgs) {
		//If we moved the Text because we clicked on this button previously, move the text down and right
		if args.Button.GetWidget().CustomData == true {
			args.Button.Text().Inset.Top = 2
		}
	})
	ButtonOptsCursorExitText = widget.ButtonOpts.CursorExitedHandler(func(args *widget.ButtonHoverEventArgs) {
		if args.Button.State() == widget.WidgetChecked {
			return
		}
		//Reset the Text inset if the cursor is no longer over the button
		args.Button.Text().Inset.Top = 0
	})

	BigButtonOptsPressedText = widget.ButtonOpts.PressedHandler(func(args *widget.ButtonPressedEventArgs) {
		args.Button.Text().Inset.Top = 4
		args.Button.GetWidget().CustomData = true
	})
	//Move the text back to start on press released
	BigButtonOptsReleasedText = widget.ButtonOpts.ReleasedHandler(func(args *widget.ButtonReleasedEventArgs) {
		if args.Button.State() == widget.WidgetChecked {
			return
		}
		args.Button.Text().Inset.Top = 0
		args.Button.GetWidget().CustomData = false
	})
	BigButtonOptsCursorEnteredText = widget.ButtonOpts.CursorEnteredHandler(func(args *widget.ButtonHoverEventArgs) {
		//If we moved the Text because we clicked on this button previously, move the text down and right
		if args.Button.GetWidget().CustomData == true {
			args.Button.Text().Inset.Top = 4
		}
	})
	BigButtonOptsCursorExitText = widget.ButtonOpts.CursorExitedHandler(func(args *widget.ButtonHoverEventArgs) {
		if args.Button.State() == widget.WidgetChecked {
			return
		}
		//Reset the Text inset if the cursor is no longer over the button
		args.Button.Text().Inset.Top = 0
	})
	BigButtonOptsStatedChangeText = widget.ButtonOpts.StateChangedHandler(func(args *widget.ButtonChangedEventArgs) {
		args.Button.Text().Inset.Top = 0
		args.Button.GetWidget().CustomData = false
		if args.Button.State() == widget.WidgetChecked {
			args.Button.Text().Inset.Top = 4
			args.Button.GetWidget().CustomData = true
		}
	})
)

func LoadImageNineSlice(key string, centerWidth, centerHeight int, isPressed bool) *image.NineSlice {
	i := Images.Get(key)

	w := i.Bounds().Dx()
	h := i.Bounds().Dy()
	p := 0
	if isPressed {
		p = centerHeight / 4
	}
	// This means to do it 3x3 equally
	if centerWidth == 0 && centerHeight == 0 {
		centerWidth = w / 3
		centerHeight = h / 3
	}
	return image.NewNineSlice(i,
		[3]int{(w - centerWidth) / 2, centerWidth, (w - centerWidth) / 2},
		[3]int{(h-centerHeight)/2 + p, centerHeight, (h-centerHeight)/2 - p},
	)
}
