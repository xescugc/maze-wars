package client

import (
	"image/color"
	"log/slog"
	"time"

	"github.com/ebitenui/ebitenui"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	cutils "github.com/xescugc/maze-wars/client/utils"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/utils"
)

type SignUpStore struct {
	*flux.ReduceStore

	Store  *store.Store
	Logger *slog.Logger

	ui         *ebitenui.UI
	textErrorW *widget.Text
	inputW     *widget.TextInput
	buttonW    *widget.Button
}

type SignUpState struct {
	Error string

	VersionError string
}

func NewSignUpStore(d *flux.Dispatcher, s *store.Store, l *slog.Logger) (*SignUpStore, error) {
	su := &SignUpStore{
		Store:  s,
		Logger: l,
	}
	su.ReduceStore = flux.NewReduceStore(d, su.Reduce, SignUpState{})

	su.buildUI()

	return su, nil
}

func (su *SignUpStore) Update() error {
	b := time.Now()
	defer utils.LogTime(su.Logger, b, "sign_up update")

	su.ui.Update()
	return nil
}

func (su *SignUpStore) Draw(screen *ebiten.Image) {
	b := time.Now()
	defer utils.LogTime(su.Logger, b, "sign_up draw")

	sutate := su.GetState().(SignUpState)
	if sutate.Error != "" {
		su.textErrorW.GetWidget().Visibility = widget.Visibility_Show
		su.textErrorW.Label = sutate.Error
	}
	if sutate.VersionError != "" {
		su.textErrorW.GetWidget().Visibility = widget.Visibility_Show
		su.textErrorW.Label = sutate.VersionError
		su.inputW.GetWidget().Disabled = true
		su.buttonW.GetWidget().Disabled = true
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
	case action.VersionError:
		sutate.VersionError = act.VersionError.Error
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
		widget.TextOpts.Text("Maze Wars", cutils.NormalFont, color.White),
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
		widget.TextInputOpts.Face(cutils.SmallFont),

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
			widget.CaretOpts.Size(cutils.SmallFont, 2),
		),

		//This text is displayed if the input is empty
		widget.TextInputOpts.Placeholder("Enter Username"),

		//This is called when the suer hits the "Enter" key.
		//There are other options that can configure this behavior
		widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
			actionDispatcher.SignUpSubmit(args.InputText)
		}),
	)

	textErrorW := widget.NewText(
		widget.TextOpts.Text(su.GetState().(SignUpState).Error, cutils.NormalFont, cutils.Red),
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
		widget.ButtonOpts.Image(cutils.ButtonImage),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Enter", cutils.SmallFont, &widget.ButtonTextColor{
			Idle:     color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
			Disabled: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
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
	textErrorW.GetWidget().Visibility = widget.Visibility_Hide
	su.textErrorW = textErrorW
	su.buttonW = buttonW
	su.inputW = inputW

	titleInputC.AddChild(titleW)
	titleInputC.AddChild(inputW)
	titleInputC.AddChild(textErrorW)
	titleInputC.AddChild(buttonW)

	rootContainer.AddChild(titleInputC)

}
