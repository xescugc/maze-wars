package client

import (
	"fmt"
	"image/color"
	"log/slog"
	"sort"
	"time"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"

	"github.com/hajimehoshi/ebiten/v2"
	cutils "github.com/xescugc/maze-wars/client/utils"
	"github.com/xescugc/maze-wars/utils"
)

type ShowLobbyView struct {
	Store  *Store
	Logger *slog.Logger

	ui *ebitenui.UI

	nameW        *widget.Text
	ownerW       *widget.Text
	playersTextW *widget.Text
	playersListW *widget.List
	startBtnW    *widget.Button
	deleteBtnW   *widget.Button
	leaveBtnW    *widget.Button
	botsC        *widget.Container
}

func NewShowLobbyView(s *Store, l *slog.Logger) *ShowLobbyView {
	sl := &ShowLobbyView{
		Store:  s,
		Logger: l,
	}

	sl.buildUI()

	return sl
}

func (sl *ShowLobbyView) Update() error {
	b := time.Now()
	defer utils.LogTime(sl.Logger, b, "show_lobby update")

	sl.ui.Update()
	return nil
}

func (sl *ShowLobbyView) Draw(screen *ebiten.Image) {
	b := time.Now()
	defer utils.LogTime(sl.Logger, b, "show_lobby draw")

	cl := sl.Store.Lobbies.FindCurrent()

	if cl != nil {
		sl.nameW.Label = fmt.Sprintf("Lobby: %s", cl.Name)
		sl.ownerW.Label = fmt.Sprintf("Owner: %s", cl.Owner)
		sl.playersTextW.Label = fmt.Sprintf("Players: %d/%d", len(cl.Players), cl.MaxPlayers)

		entries := make([]any, 0, len(cl.Players))
		for p := range cl.Players {
			entries = append(entries, cutils.ListEntry{
				ID:   p,
				Text: p,
			})
		}
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].(cutils.ListEntry).ID > entries[j].(cutils.ListEntry).ID
		})
		//if !cutils.EqualListEntries(entries, sl.playersListW.Entries().([]any)) {
		//sl.playersListW.SetEntries(entries)
		//}

		if sl.Store.Users.Username() == cl.Owner {
			sl.startBtnW.GetWidget().Visibility = widget.Visibility_Show
			sl.startBtnW.GetWidget().Disabled = len(cl.Players) < 2

			sl.deleteBtnW.GetWidget().Visibility = widget.Visibility_Show

			sl.leaveBtnW.GetWidget().Visibility = widget.Visibility_Hide

			sl.botsC.GetWidget().Visibility = widget.Visibility_Show
		} else {
			sl.startBtnW.GetWidget().Visibility = widget.Visibility_Hide
			sl.deleteBtnW.GetWidget().Visibility = widget.Visibility_Hide

			sl.leaveBtnW.GetWidget().Visibility = widget.Visibility_Show

			sl.botsC.GetWidget().Visibility = widget.Visibility_Hide
		}
	}

	sl.ui.Draw(screen)
}

func (sl *ShowLobbyView) buildUI() {
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	sl.ui = &ebitenui.UI{
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
				StretchHorizontal:  false,
				StretchVertical:    false,
			}),
		),
	)

	nameW := widget.NewText(
		widget.TextOpts.Text("Lobby: NAME", cutils.NormalFont, color.White),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
			widget.WidgetOpts.MinSize(100, 100),
		),
	)
	ownerW := widget.NewText(
		widget.TextOpts.Text("Owner: NAME", cutils.NormalFont, color.White),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
			widget.WidgetOpts.MinSize(100, 100),
		),
	)
	playersTextW := widget.NewText(
		widget.TextOpts.Text("Players 0/6", cutils.NormalFont, color.White),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
			widget.WidgetOpts.MinSize(100, 100),
		),
	)

	botsTextW := widget.NewText(
		widget.TextOpts.Text("Bots", cutils.NormalFont, color.White),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
			widget.WidgetOpts.MinSize(100, 100),
		),
	)
	addBotBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  false,
			}),
		),

		// specify the images to sle
		widget.ButtonOpts.Image(cutils.ButtonImage),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("+", cutils.SmallFont, &widget.ButtonTextColor{
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
			cl := sl.Store.Lobbies.FindCurrent()
			if len(cl.Players)+1 > cl.MaxPlayers {
				return
			}
			actionDispatcher.JoinLobby(cl.ID, fmt.Sprintf("Bot-%d", len(cl.Players)+1), isBot)
		}),
	)
	removeBotBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  false,
			}),
		),

		// specify the images to sle
		widget.ButtonOpts.Image(cutils.ButtonImage),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("-", cutils.SmallFont, &widget.ButtonTextColor{
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
			cl := sl.Store.Lobbies.FindCurrent()
			for p, ib := range cl.Players {
				if ib {
					actionDispatcher.LeaveLobby(cl.ID, p)
					break
				}
			}
		}),
	)
	botsC := widget.NewContainer(
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
	botsC.AddChild(botsTextW)
	botsC.AddChild(removeBotBtnW)
	botsC.AddChild(addBotBtnW)

	entries := make([]any, 0, 0)
	playersListW := widget.NewList(
		// Set how wide the list should be
		widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
			//widget.WidgetOpts.MinSize(150, 0),
			widget.WidgetOpts.MinSize(800, 200),
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
				StretchVertical:    true,
			}),
		)),
		// Set the entries in the list
		widget.ListOpts.Entries(entries),
		widget.ListOpts.ScrollContainerOpts(
			// Set the background images/color for the list
			widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
				Idle:     image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
				Disabled: image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
				Mask:     image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
			}),
		),
		widget.ListOpts.SliderOpts(
			// Set the background images/color for the background of the slider track
			widget.SliderOpts.Images(&widget.SliderTrackImage{
				Idle:  image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
				Hover: image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
			}, cutils.ButtonImage),
			widget.SliderOpts.MinHandleSize(5),
			// Set how wide the track should be
			widget.SliderOpts.TrackPadding(widget.NewInsetsSimple(2)),
		),
		// Hide the horizontal slider
		widget.ListOpts.HideHorizontalSlider(),
		// Set the font for the list options
		widget.ListOpts.EntryFontFace(cutils.SmallFont),
		// Set the colors for the list
		widget.ListOpts.EntryColor(&widget.ListEntryColor{
			Selected:                   color.NRGBA{0, 255, 0, 255},                 // Foreground color for the unfocused selected entry
			Unselected:                 color.NRGBA{254, 255, 255, 255},             // Foreground color for the unfocused unselected entry
			SelectedBackground:         color.NRGBA{R: 130, G: 130, B: 200, A: 255}, // Background color for the unfocused selected entry
			SelectedFocusedBackground:  color.NRGBA{R: 130, G: 130, B: 170, A: 255}, // Background color for the focused selected entry
			FocusedBackground:          color.NRGBA{R: 170, G: 170, B: 180, A: 255}, // Background color for the focused unselected entry
			DisabledUnselected:         color.NRGBA{100, 100, 100, 255},             // Foreground color for the disabled unselected entry
			DisabledSelected:           color.NRGBA{100, 100, 100, 255},             // Foreground color for the disabled selected entry
			DisabledSelectedBackground: color.NRGBA{100, 100, 100, 255},             // Background color for the disabled selected entry
		}),
		// This required function returns the string displayed in the list
		widget.ListOpts.EntryLabelFunc(func(e interface{}) string {
			return e.(cutils.ListEntry).Text
		}),
		// Padding for each entry
		widget.ListOpts.EntryTextPadding(widget.NewInsetsSimple(5)),
		// Text position for each entry
		widget.ListOpts.EntryTextPosition(widget.TextPositionStart, widget.TextPositionCenter),
		// This handler defines what function to run when a list item is selected.
		//widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
		//entry := args.Entry.(ListEntry)
		//fmt.Println("Entry Selected")
		//}),
	)

	startBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  false,
			}),
		),

		// specify the images to sle
		widget.ButtonOpts.Image(cutils.ButtonImage),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Start", cutils.SmallFont, &widget.ButtonTextColor{
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
			cl := sl.Store.Lobbies.FindCurrent()
			actionDispatcher.StartLobby(cl.ID)
		}),
	)

	deleteBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  false,
			}),
		),

		// specify the images to sle
		widget.ButtonOpts.Image(cutils.ButtonImage),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Delete", cutils.SmallFont, &widget.ButtonTextColor{
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
			cl := sl.Store.Lobbies.FindCurrent()
			actionDispatcher.DeleteLobby(cl.ID)
			actionDispatcher.NavigateTo(utils.LobbiesRoute)
		}),
	)

	leaveBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  false,
			}),
		),

		// specify the images to sle
		widget.ButtonOpts.Image(cutils.ButtonImage),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Leave", cutils.SmallFont, &widget.ButtonTextColor{
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
			cl := sl.Store.Lobbies.FindCurrent()
			un := sl.Store.Users.Username()
			actionDispatcher.LeaveLobby(cl.ID, un)
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

	sl.nameW = nameW
	sl.ownerW = ownerW
	sl.playersTextW = playersTextW
	sl.playersListW = playersListW
	sl.startBtnW = startBtnW
	sl.deleteBtnW = deleteBtnW
	sl.leaveBtnW = leaveBtnW
	sl.botsC = botsC

	buttonsC.AddChild(leaveBtnW)
	buttonsC.AddChild(deleteBtnW)
	buttonsC.AddChild(startBtnW)

	mainContainer.AddChild(nameW)
	mainContainer.AddChild(ownerW)
	mainContainer.AddChild(playersTextW)
	mainContainer.AddChild(botsC)
	mainContainer.AddChild(playersListW)
	mainContainer.AddChild(buttonsC)

	rootContainer.AddChild(mainContainer)
}
