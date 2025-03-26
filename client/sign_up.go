package client

import (
	"image"
	"log/slog"
	"time"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/maze-wars/action"
	cutils "github.com/xescugc/maze-wars/client/utils"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/unit"
	"github.com/xescugc/maze-wars/utils"
)

type SignUpStore struct {
	*flux.ReduceStore[SignUpState, *action.Action]

	Store  *store.Store
	Logger *slog.Logger

	ui         *ebitenui.UI
	textErrorW *widget.Text
	inputW     *widget.TextInput
	buttonW    *widget.Button
	imageG     *widget.Graphic
}

type SignUpState struct {
	Error string

	VersionError string

	ImageKey string
	Username string
}

func NewSignUpStore(d *flux.Dispatcher[*action.Action], s *store.Store, un, ik string, l *slog.Logger) (*SignUpStore, error) {
	su := &SignUpStore{
		Store:  s,
		Logger: l,
	}
	if ik == "" {
		ik = unit.TypeStrings()[0]
	}
	su.ReduceStore = flux.NewReduceStore(d, su.Reduce, SignUpState{
		ImageKey: ik,
		Username: un,
	})

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

	sutate := su.GetState()
	su.textErrorW.GetWidget().Visibility = widget.Visibility_Hide
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
	su.imageG.Image = cutils.Images.Get(unit.Units[sutate.ImageKey].ProfileKey())
	su.ui.Draw(screen)
}

func (su *SignUpStore) Reduce(state SignUpState, act *action.Action) SignUpState {
	switch act.Type {
	case action.SignUpError:
		state.Error = act.SignUpError.Error
	case action.UserSignUpChangeImage:
		state.ImageKey = act.UserSignUpChangeImage.ImageKey
	case action.VersionError:
		state.VersionError = act.VersionError.Error
	}

	return state
}

func (su *SignUpStore) buildUI() {
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(
			widget.NewStackedLayout(),
		),
		widget.ContainerOpts.BackgroundImage(cutils.ImageToNineSlice(cutils.BGKey)),
	)

	su.ui = &ebitenui.UI{
		Container: rootContainer,
	}

	rootContainer.AddChild(
		su.topbarUI(),
		su.signUpFormUI(),
		su.errorMessageUI(),
	)
}

func (su *SignUpStore) errorMessageUI() *widget.Container {
	errorMessageC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(widgetPadding),
		)),
	)

	errorMessageTxt := widget.NewText(
		widget.TextOpts.Text("ERROR", cutils.NormalFont, cutils.Red),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.Insets(widget.Insets{
			Top: 100,
		}),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	su.textErrorW = errorMessageTxt

	errorMessageC.AddChild(errorMessageTxt)

	return errorMessageC
}

func (su *SignUpStore) topbarUI() *widget.Container {
	topbarC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(widgetPadding),
		)),
	)

	topbarRC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(30),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)

	logoBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.LogoButtonResource()),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			//actionDispatcher.NavigateTo(utils.LobbiesRoute)
		}),
	)

	topbarRC.AddChild(logoBtnW)

	topbarC.AddChild(topbarRC)

	return topbarC
}

func (su *SignUpStore) signUpFormUI() *widget.Container {
	ss := su.GetState()
	signUpFormC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	frameC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(80)),
			widget.RowLayoutOpts.Spacing(40),
		)),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.SetupGameFrameKey, 1, 1, !isPressed)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	unitsFacetsC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(5),
			//Define how much padding to inset the child content
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(3)),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(2, 2),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{false, false, false, false, false}, []bool{false, false}),
		)),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.ToolTipBGKey, 8, 8, !isPressed)),
	)

	unitsFacetsW := widget.NewWindow(
		widget.WindowOpts.Contents(unitsFacetsC),
		widget.WindowOpts.Modal(),
		widget.WindowOpts.CloseMode(widget.CLICK),
	)

	for _, t := range unit.TypeStrings() {
		aux := t
		imageGraphicW := widget.NewGraphic(
			widget.GraphicOpts.Image(
				cutils.Images.Get(unit.Units[aux].FacesetKey()),
			),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.MouseButtonPressedHandler(func(args *widget.WidgetMouseButtonPressedEventArgs) {
					actionDispatcher.UserSignUpChangeImage(aux)
				}),
			),
		)
		unitsFacetsC.AddChild(imageGraphicW)
	}

	imageProfileC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(116, 116),
		),
	)

	imageGraphicBtnW := widget.NewButton(
		widget.ButtonOpts.Image(cutils.ButtonImageFromKey(cutils.Border4Key, 4, 4)),
	)
	imageGraphicW := widget.NewGraphic(
		widget.GraphicOpts.Image(
			cutils.Images.Get(unit.Units[ss.ImageKey].ProfileKey()),
		),
	)

	editBtnC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	editBtnStackWC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
			}),
		),
	)
	editBtnGridWC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(1),
		)),
	)

	editBtnW := widget.NewButton(
		widget.ButtonOpts.Image(cutils.ButtonImageFromKey(cutils.Border4Key, 4, 4)),
		//widget.ButtonOpts.Image(cutils.ButtonBorderResource()),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				MaxWidth:  24,
				MaxHeight: 24,
			}),
		),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			//Get the preferred size of the content
			x, y := unitsFacetsW.Contents.PreferredSize()
			//Create a rect with the preferred size of the content
			r := image.Rect(0, 0, x, y)
			//Use the Add method to move the window to the specified point
			bw := args.Button.GetWidget()
			r = r.Add(image.Point{bw.Rect.Min.X + 25, bw.Rect.Min.Y})
			//Set the windows location to the rect.
			unitsFacetsW.SetLocation(r)
			//Add the window to the UI.
			//Note: If the window is already added, this will just move the window and not add a duplicate.
			su.ui.AddWindow(unitsFacetsW)
		}),
	)
	editGraphicW := widget.NewGraphic(
		widget.GraphicOpts.Image(
			cutils.Images.Get(cutils.EditIconKey),
		),
	)

	editBtnGridWC.AddChild(
		editBtnW,
	)
	editBtnStackWC.AddChild(
		editBtnGridWC,
		editGraphicW,
	)
	editBtnC.AddChild(
		editBtnStackWC,
	)

	imageProfileC.AddChild(
		imageGraphicBtnW,
		imageGraphicW,
		editBtnC,
	)

	var enterBtnW *widget.Button
	nameInputW := widget.NewTextInput(
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),

		widget.TextInputOpts.Image(cutils.TextInputResource()),

		//Set the font face and size for the widget
		widget.TextInputOpts.Face(cutils.SmallFont),

		//Set the colors for the text and caret
		widget.TextInputOpts.Color(cutils.TextInputColor()),

		//Set how much padding there is between the edge of the input and the text
		widget.TextInputOpts.Padding(widget.NewInsetsSimple(5)),

		//Set the font and width of the caret
		widget.TextInputOpts.CaretOpts(
			widget.CaretOpts.Size(cutils.SmallFont, 2),
		),

		//This text is displayed if the input is empty
		widget.TextInputOpts.Placeholder("Username"),

		//This is called when the user hits the "Enter" key.
		//There are other options that can configure this behavior
		widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
			enterBtnW.Click()
		}),
	)
	nameInputW.SetText(ss.Username)

	enterBtnW = widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.BigButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("ENTER", cutils.Font40, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   50,
			Right:  50,
			Top:    10,
			Bottom: 10,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			ss := su.GetState()
			actionDispatcher.SignUpSubmit(nameInputW.GetText(), ss.ImageKey)
		}),

		cutils.BigButtonOptsPressedText,
		cutils.BigButtonOptsReleasedText,
		cutils.BigButtonOptsCursorEnteredText,
		cutils.BigButtonOptsCursorExitText,
	)

	nameInputW.Focus(true)

	frameC.AddChild(
		imageProfileC,
		nameInputW,
		enterBtnW,
	)

	signUpFormC.AddChild(frameC)

	su.inputW = nameInputW
	su.buttonW = enterBtnW
	su.imageG = imageGraphicW

	return signUpFormC
}
