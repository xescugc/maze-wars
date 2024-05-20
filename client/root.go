package client

import (
	"fmt"
	"image/color"
	"log/slog"
	"time"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	cutils "github.com/xescugc/maze-wars/client/utils"
	"github.com/xescugc/maze-wars/utils"
)

var (
	buttonImageL, _ = loadButtonImageL()
)

type RootStore struct {
	*flux.ReduceStore

	Store  *Store
	Logger *slog.Logger

	ui           *ebitenui.UI
	textPlayersW *widget.Text
}

type RootState struct {
	TotalUsers int
}

func NewRootStore(d *flux.Dispatcher, s *Store, l *slog.Logger) (*RootStore, error) {
	rs := &RootStore{
		Store:  s,
		Logger: l,
	}
	rs.ReduceStore = flux.NewReduceStore(d, rs.Reduce, RootState{})

	rs.buildUI()

	return rs, nil
}

func (rs *RootStore) Update() error {
	b := time.Now()
	defer utils.LogTime(rs.Logger, b, "root update")

	rs.ui.Update()
	return nil
}

func (rs *RootStore) Draw(screen *ebiten.Image) {
	b := time.Now()
	defer utils.LogTime(rs.Logger, b, "root draw")

	rstate := rs.GetState().(RootState)
	rs.textPlayersW.Label = fmt.Sprintf("Users online: %d", rstate.TotalUsers)
	rs.ui.Draw(screen)
}

func (rs *RootStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	rstate, ok := state.(RootState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.SyncUsers:
		rstate.TotalUsers = act.SyncUsers.TotalUsers
	}

	return rstate
}

func (rs *RootStore) buildUI() {
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

	rs.ui = &ebitenui.UI{
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

	textPlayersW := widget.NewText(
		widget.TextOpts.Text("", cutils.SmallFont, color.White),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
	)

	playBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
				MaxWidth: 150,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(buttonImageL),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Play", cutils.SmallFont, &widget.ButtonTextColor{
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
			actionDispatcher.JoinWaitingRoom(rs.Store.Users.Username())
		}),
	)

	lobbiesBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
				MaxWidth: 150,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(buttonImageL),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Lobbies", cutils.SmallFont, &widget.ButtonTextColor{
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
			actionDispatcher.NavigateTo(utils.LobbiesRoute)
		}),
	)

	rs.textPlayersW = textPlayersW

	titleInputC.AddChild(titleW)
	titleInputC.AddChild(textPlayersW)
	titleInputC.AddChild(playBtnW)
	titleInputC.AddChild(lobbiesBtnW)

	rootContainer.AddChild(titleInputC)

}

func loadButtonImageL() (*widget.ButtonImage, error) {
	idle := image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 180, A: 255})

	hover := image.NewNineSliceColor(color.NRGBA{R: 130, G: 130, B: 150, A: 255})

	pressed := image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 120, A: 255})

	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: pressed,
	}, nil
}
