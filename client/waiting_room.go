package client

import (
	"fmt"
	"image/color"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
)

type WaitingRoomStore struct {
	*flux.ReduceStore

	Store *Store

	ui           *ebitenui.UI
	textPlayersW *widget.Text
	textColdownW *widget.Text
}

type WaitingRoomState struct {
	TotalPlayers int
	Size         int
	Countdown    int
}

func NewWaitingRoomStore(d *flux.Dispatcher, s *Store) *WaitingRoomStore {
	wr := &WaitingRoomStore{
		Store: s,
	}
	wr.ReduceStore = flux.NewReduceStore(d, wr.Reduce, WaitingRoomState{})

	wr.buildUI()

	return wr
}

func (wr *WaitingRoomStore) Update() error {
	wr.ui.Update()
	return nil
}

func (wr *WaitingRoomStore) Draw(screen *ebiten.Image) {
	wrstate := wr.GetState().(WaitingRoomState)
	wr.textPlayersW.Label = fmt.Sprintf("%d/%d", wrstate.TotalPlayers, wrstate.Size)
	wr.textColdownW.Label = fmt.Sprintf("(%ds to reduce the size, minimum is 2)", wrstate.Countdown)
	wr.ui.Draw(screen)
}

func (wr *WaitingRoomStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	wrtate, ok := state.(WaitingRoomState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.SyncWaitingRoom:
		wrtate.TotalPlayers = act.SyncWaitingRoom.TotalPlayers
		wrtate.Size = act.SyncWaitingRoom.Size
		wrtate.Countdown = act.SyncWaitingRoom.Countdown
	}

	return wrtate
}

func (wr *WaitingRoomStore) buildUI() {
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	waitingRoomC := widget.NewContainer(
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

	wr.ui = &ebitenui.UI{
		Container: rootContainer,
	}

	titleW := widget.NewText(
		widget.TextOpts.Text("Waiting for players to join", normalFont, color.White),
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
		widget.TextOpts.Text("", smallFont, color.White),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
	)

	textColdownW := widget.NewText(
		widget.TextOpts.Text("", smallFont, color.White),
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
		widget.ButtonOpts.Text("EXIT", smallFont, &widget.ButtonTextColor{
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
			actionDispatcher.ExitWaitingRoom(wr.Store.Users.Username())
		}),
	)

	wr.textPlayersW = textPlayersW
	wr.textColdownW = textColdownW

	waitingRoomC.AddChild(titleW)
	waitingRoomC.AddChild(textPlayersW)
	waitingRoomC.AddChild(textColdownW)
	waitingRoomC.AddChild(buttonW)

	rootContainer.AddChild(waitingRoomC)

}
