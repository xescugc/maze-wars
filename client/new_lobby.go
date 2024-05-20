package client

import (
	"image/color"
	"log/slog"
	"strconv"
	"time"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/gofrs/uuid"

	"github.com/hajimehoshi/ebiten/v2"
	cutils "github.com/xescugc/maze-wars/client/utils"
	"github.com/xescugc/maze-wars/utils"
)

type NewLobbyView struct {
	Store  *Store
	Logger *slog.Logger

	ui         *ebitenui.UI
	createBtnW *widget.Button
	nameInputW *widget.TextInput
}

func NewNewLobbyView(s *Store, l *slog.Logger) *NewLobbyView {
	nl := &NewLobbyView{
		Store:  s,
		Logger: l,
	}

	nl.buildUI()

	return nl
}

func (nl *NewLobbyView) Update() error {
	b := time.Now()
	defer utils.LogTime(nl.Logger, b, "new_lobby update")

	nl.ui.Update()
	return nil
}

func (nl *NewLobbyView) Draw(screen *ebiten.Image) {
	b := time.Now()
	defer utils.LogTime(nl.Logger, b, "new_lobby draw")

	nl.createBtnW.GetWidget().Disabled = len(nl.nameInputW.GetText()) == 0

	nl.ui.Draw(screen)
}

func (nl *NewLobbyView) buildUI() {
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	nl.ui = &ebitenui.UI{
		Container: rootContainer,
	}

	mainContainer := widget.NewContainer(
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

	titleW := widget.NewText(
		widget.TextOpts.Text("New Lobby", cutils.NormalFont, color.White),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
			widget.WidgetOpts.MinSize(100, 100),
		),
	)

	nameInputW := widget.NewTextInput(
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
		//If the NineSlice image has a minimum size, the widget will nle that or
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
		widget.TextInputOpts.Placeholder("Enter name"),

		//This is called when the user hits the "Enter" key.
		//There are other options that can configure this behavior
		//widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
		//actionDispatcher.NewLobbySubmit(args.InputText)
		//}),
	)

	sliderTextLabelW := widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
			//Stretch:  true,
		}))),
		widget.LabelOpts.Text("Select number of players:", cutils.SmallFont, &widget.LabelColor{
			Idle:     color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
			Disabled: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		}),
	)
	sliderNumberLabelW := widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionEnd,
			Stretch:  true,
		}))),
		widget.LabelOpts.Text("0", cutils.SmallFont, &widget.LabelColor{
			Idle:     color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
			Disabled: color.NRGBA{R: 200, G: 200, B: 200, A: 255},
		}),
	)
	playersSliderW := widget.NewSlider(
		// Set the slider orientation - n/s vs e/w
		widget.SliderOpts.Direction(widget.DirectionHorizontal),
		// Set the minimum and maximum value for the slider
		widget.SliderOpts.MinMax(2, 6),

		widget.SliderOpts.WidgetOpts(
			// Set the Widget to layout in the center on the screen
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				Stretch:  true,
				//MinWidth: 500,
			}),
			// Set the widget's dimensions
			widget.WidgetOpts.MinSize(465, 6),
		),
		widget.SliderOpts.Images(
			// Set the track images
			&widget.SliderTrackImage{
				Idle:  image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
				Hover: image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
			},
			// Set the handle images
			&widget.ButtonImage{
				Idle:    image.NewNineSliceColor(color.NRGBA{255, 100, 100, 255}),
				Hover:   image.NewNineSliceColor(color.NRGBA{255, 100, 100, 255}),
				Pressed: image.NewNineSliceColor(color.NRGBA{255, 100, 100, 255}),
			},
		),
		// Set the size of the handle
		widget.SliderOpts.FixedHandleSize(6),
		// Set the offset to display the track
		widget.SliderOpts.TrackOffset(0),
		// Set the size to move the handle
		widget.SliderOpts.PageSizeFunc(func() int {
			return 1
		}),
		// Set the callback to call when the slider value is changed
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			sliderNumberLabelW.Label = strconv.Itoa(args.Current)
		}),
	)
	sliderC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				//Stretch:  true,
			}),
		),
	)
	sliderC.AddChild(playersSliderW)
	sliderC.AddChild(sliderNumberLabelW)

	createBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  false,
			}),
		),

		// specify the images to nle
		widget.ButtonOpts.Image(cutils.ButtonImage),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Create", cutils.SmallFont, &widget.ButtonTextColor{
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
			lid := uuid.Must(uuid.NewV4()).String()

			actionDispatcher.CreateLobby(lid, nl.Store.Users.Username(), nameInputW.GetText(), playersSliderW.Current)
			actionDispatcher.SelectLobby(lid)
			actionDispatcher.NavigateTo(utils.ShowLobbyRoute)
		}),
	)

	backBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  false,
			}),
		),

		// specify the images to nle
		widget.ButtonOpts.Image(cutils.ButtonImage),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Back", cutils.SmallFont, &widget.ButtonTextColor{
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
			actionDispatcher.NavigateTo(utils.LobbiesRoute)
		}),
	)

	buttonsC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	nl.createBtnW = createBtnW
	nl.nameInputW = nameInputW
	nameInputW.Focus(true)

	buttonsC.AddChild(backBtnW)
	buttonsC.AddChild(createBtnW)

	mainContainer.AddChild(titleW)
	mainContainer.AddChild(nameInputW)
	mainContainer.AddChild(sliderTextLabelW)
	mainContainer.AddChild(sliderC)
	mainContainer.AddChild(buttonsC)

	rootContainer.AddChild(mainContainer)

}
