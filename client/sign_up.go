package client

import (
	"image"
	"image/color"

	"github.com/ebitenui/ebitenui"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/store"
)

var (
	buttonImage, _ = loadButtonImage()
)

type SignUpStore struct {
	*flux.ReduceStore

	Store *store.Store

	Camera *CameraStore

	ui          *ebitenui.UI
	inputErrorW *widget.Text
}

type SignUpState struct {
	Error string
}

func NewSignUpStore(d *flux.Dispatcher, s *store.Store) (*SignUpStore, error) {
	su := &SignUpStore{
		Store: s,
	}
	su.ReduceStore = flux.NewReduceStore(d, su.Reduce, SignUpState{})

	su.buildUI()

	return su, nil
}

func (su *SignUpStore) Update() error {
	su.ui.Update()
	return nil
}

func (su *SignUpStore) Draw(screen *ebiten.Image) {
	sutate := su.GetState().(SignUpState)
	if sutate.Error != "" {
		su.inputErrorW.GetWidget().Visibility = widget.Visibility_Show
		su.inputErrorW.Label = sutate.Error
	}
	su.ui.Draw(screen)
}

func (su *SignUpStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	sutate, ok := state.(SignUpState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.SignUpError:
		sutate.Error = act.SignUpError.Error
	}

	return sutate
}

func (su *SignUpStore) buildUI() {
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	titleInputC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(20)),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				StretchHorizontal:  true,
				StretchVertical:    false,
			}),
		),
	)

	su.ui = &ebitenui.UI{
		Container: rootContainer,
	}

	titleW := widget.NewText(
		widget.TextOpts.Text("Maze Wars", normalFont, color.White),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
			widget.WidgetOpts.MinSize(100, 100),
		),
	)

	inputW := widget.NewTextInput(
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
				MaxWidth: 500,
			}),
		),

		// Set the keyboard type when opened on mobile devices.
		//widget.TextInputOpts.MobileInputMode(jsUtil.TEXT),

		//Set the Idle and Disabled background image for the text input
		//If the NineSlice image has a minimum size, the widget will sue that or
		// widget.WidgetOpts.MinSize; whichever is greater
		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle:     euiimage.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 100, A: 255}),
			Disabled: euiimage.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 100, A: 255}),
		}),

		//Set the font face and size for the widget
		widget.TextInputOpts.Face(smallFont),

		//Set the colors for the text and caret
		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:          color.NRGBA{254, 255, 255, 255},
			Disabled:      color.NRGBA{R: 200, G: 200, B: 200, A: 255},
			Caret:         color.NRGBA{254, 255, 255, 255},
			DisabledCaret: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		}),

		//Set how much padding there is between the edge of the input and the text
		widget.TextInputOpts.Padding(widget.NewInsetsSimple(5)),

		//Set the font and width of the caret
		widget.TextInputOpts.CaretOpts(
			widget.CaretOpts.Size(smallFont, 2),
		),

		//This text is displayed if the input is empty
		widget.TextInputOpts.Placeholder("Enter Username"),

		//This is called when the suer hits the "Enter" key.
		//There are other options that can configure this behavior
		widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
			actionDispatcher.SignUpSubmit(args.InputText)
		}),
	)

	inputErrorW := widget.NewText(
		widget.TextOpts.Text(su.GetState().(SignUpState).Error, normalFont, red),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
	)

	buttonW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  false,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(buttonImage),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Enter", smallFont, &widget.ButtonTextColor{
			Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
		}),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			actionDispatcher.SignUpSubmit(inputW.GetText())
		}),
	)

	inputW.Focus(true)
	inputErrorW.GetWidget().Visibility = widget.Visibility_Hide
	su.inputErrorW = inputErrorW

	titleInputC.AddChild(titleW)
	titleInputC.AddChild(inputW)
	titleInputC.AddChild(inputErrorW)
	titleInputC.AddChild(buttonW)

	rootContainer.AddChild(titleInputC)

}

func loadButtonImage() (*widget.ButtonImage, error) {
	idle := euiimage.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 180, A: 255})

	hover := euiimage.NewNineSliceColor(color.NRGBA{R: 130, G: 130, B: 150, A: 255})

	pressed := euiimage.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 120, A: 255})

	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: pressed,
	}, nil
}

func buttonImageFromImage(i image.Image) *widget.ButtonImage {
	ei := ebiten.NewImageFromImage(i)
	nsi := euiimage.NewNineSliceSimple(ei, i.Bounds().Dx(), i.Bounds().Dy())

	dest := i
	cm := colorm.ColorM{}
	cm.Scale(2, 0.5, 0.5, 0.9)
	edest := ebiten.NewImageFromImage(dest)
	colorm.DrawImage(edest, ei, cm, nil)
	dsi := euiimage.NewNineSliceSimple(edest, dest.Bounds().Dx(), dest.Bounds().Dy())
	return &widget.ButtonImage{
		Idle:     nsi,
		Hover:    nsi,
		Pressed:  nsi,
		Disabled: dsi,
	}
}
