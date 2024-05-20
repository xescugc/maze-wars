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

const (
	isBot = true
)

type LobbiesView struct {
	Store  *Store
	Logger *slog.Logger

	ui           *ebitenui.UI
	lobbiesListW *widget.List
}

func EqualListEntries(le, nle []any) bool {
	if len(le) != len(nle) {
		return false
	} else {
		for i, e := range nle {
			if le[i].(ListEntry) != e.(ListEntry) {
				return false
			}
		}
	}
	return true
}

type ListEntry struct {
	ID   string
	Text string
}

func NewLobbiesView(s *Store, l *slog.Logger) *LobbiesView {
	lv := &LobbiesView{
		Store:  s,
		Logger: l,
	}

	lv.buildUI()

	return lv
}

func (lv *LobbiesView) Update() error {
	b := time.Now()
	defer utils.LogTime(lv.Logger, b, "lobby update")

	lv.ui.Update()
	return nil
}

func (lv *LobbiesView) Draw(screen *ebiten.Image) {
	b := time.Now()
	defer utils.LogTime(lv.Logger, b, "lobby draw")

	lbs := lv.Store.Lobbies.List()

	entries := make([]any, 0, len(lbs))
	for _, l := range lbs {
		le := ListEntry{
			ID: l.ID,
			Text: fmt.Sprintf("%s %s %s",
				cutils.FillIn(l.Name, 10),
				cutils.FillIn(fmt.Sprintf("%d/%d", len(l.Players), l.MaxPlayers), 10),
				cutils.FillIn(l.Owner, 10),
			),
		}
		entries = append(entries, le)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].(ListEntry).ID > entries[j].(ListEntry).ID
	})
	if !EqualListEntries(entries, lv.lobbiesListW.Entries().([]any)) {
		lv.lobbiesListW.SetEntries(entries)
	}

	lv.ui.Draw(screen)
}

func (lv *LobbiesView) buildUI() {
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	lv.ui = &ebitenui.UI{
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

	titleW := widget.NewText(
		widget.TextOpts.Text("Lobbies", cutils.NormalFont, color.White),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
			widget.WidgetOpts.MinSize(100, 100),
		),
	)

	newBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionEnd,
				Stretch:  false,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(buttonImageL),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("New Lobby", cutils.SmallFont, &widget.ButtonTextColor{
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
			actionDispatcher.NavigateTo(utils.NewLobbyRoute)
		}),
	)

	refreshBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionEnd,
				Stretch:  false,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(buttonImageL),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Refresh", cutils.SmallFont, &widget.ButtonTextColor{
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
			actionDispatcher.RefreshLobbies()
		}),
	)

	backBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				Stretch:  false,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(buttonImageL),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Back", cutils.SmallFont, &widget.ButtonTextColor{
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
			actionDispatcher.NavigateTo(utils.RootRoute)
		}),
	)

	entries := make([]any, 0, 0)
	lobbiesW := widget.NewList(
		// Set how wide the list should be
		widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
			//widget.WidgetOpts.MinSize(150, 0),
			widget.WidgetOpts.MinSize(800, 800),
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
			l := e.(ListEntry)
			return l.Text
		}),
		// Padding for each entry
		widget.ListOpts.EntryTextPadding(widget.NewInsetsSimple(5)),
		// Text position for each entry
		widget.ListOpts.EntryTextPosition(widget.TextPositionStart, widget.TextPositionCenter),
		// This handler defines what function to run when a list item is selected.
		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			l := args.Entry.(ListEntry)
			un := lv.Store.Users.Username()
			lb := lv.Store.Lobbies.FindByID(l.ID)
			if len(lb.Players)+1 > lb.MaxPlayers {
				return
			}
			actionDispatcher.JoinLobby(l.ID, un, !isBot)
			actionDispatcher.SelectLobby(l.ID)
			actionDispatcher.NavigateTo(utils.ShowLobbyRoute)
		}),
	)

	buttonsC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionEnd,
			}),
		),
	)

	buttonsC.AddChild(backBtnW)
	buttonsC.AddChild(refreshBtnW)
	buttonsC.AddChild(newBtnW)
	mainContainer.AddChild(titleW)
	mainContainer.AddChild(buttonsC)
	mainContainer.AddChild(lobbiesW)

	lv.lobbiesListW = lobbiesW

	rootContainer.AddChild(mainContainer)
}
