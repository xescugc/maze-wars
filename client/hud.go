package client

import (
	"fmt"
	stdimage "image"
	"image/color"
	"sort"
	"strconv"
	"strings"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/inputer"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/unit"
	"github.com/xescugc/maze-wars/utils"
)

// HUDStore is in charge of keeping track of all the elements
// on the player HUD that are static and always seen
type HUDStore struct {
	*flux.ReduceStore

	game *Game

	ui *ebitenui.UI

	input inputer.Inputer

	statsListW   *widget.List
	incomeTextW  *widget.Text
	winLoseTextW *widget.Text
	unitsC       *widget.Container
	towersC      *widget.Container
}

// HUDState stores the HUD state
type HUDState struct {
	SelectedTower   *SelectedTower
	TowerOpenMenuID string

	LastCursorPosition utils.Object
	CheckedPath        bool

	ShowStats bool
}

type SelectedTower struct {
	store.Tower

	Invalid bool
}

var (
	// The key value of this maps is the TYPE of the Unit|Tower
	unitKeybinds  = make(map[string]ebiten.Key)
	towerKeybinds = make(map[string]ebiten.Key)
)

func init() {
	for _, u := range unit.Units {
		var k ebiten.Key
		err := k.UnmarshalText([]byte(u.Keybind))
		if err != nil {
			panic(err)
		}
		unitKeybinds[u.Type.String()] = k
	}

	for _, t := range tower.Towers {
		var k ebiten.Key
		err := k.UnmarshalText([]byte(t.Keybind))
		if err != nil {
			panic(err)
		}
		towerKeybinds[t.Type.String()] = k
	}
}

// NewHUDStore creates a new HUDStore with the Dispatcher d and the Game g
func NewHUDStore(d *flux.Dispatcher, i inputer.Inputer, g *Game) (*HUDStore, error) {
	hs := &HUDStore{
		game: g,

		input: i,
	}
	hs.ReduceStore = flux.NewReduceStore(d, hs.Reduce, HUDState{
		ShowStats: true,
	})

	hs.buildUI()

	return hs, nil
}

func (hs *HUDStore) Update() error {
	hs.ui.Update()

	cs := hs.game.Camera.GetState().(CameraState)
	hst := hs.GetState().(HUDState)
	x, y := hs.input.CursorPosition()
	cp := hs.game.Store.Players.FindCurrent()
	tws := hs.game.Store.Towers.List()
	// Only send a CursorMove when the curso has actually moved
	if hst.LastCursorPosition.X != float64(x) || hst.LastCursorPosition.Y != float64(y) {
		actionDispatcher.CursorMove(x, y)
	}
	// If the Current player is dead or has no more lives there are no
	// mo actions that can be done
	// TODO Be able to move the camera when won or lose
	if cp.Lives == 0 || cp.Winner {
		return nil
	}
	if hs.input.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		clickAbsolute := utils.Object{
			X: float64(x) + cs.X,
			Y: float64(y) + cs.Y,
			W: 1, H: 1,
		}

		if hst.SelectedTower != nil && !hst.SelectedTower.Invalid {
			// We double check that placing the tower would not block the path
			utws := make([]utils.Object, 0, 0)
			for _, t := range tws {
				// If the tower does not belong to the current user then we can skip
				// as it's outside the Players Building Zone
				if t.PlayerID != cp.ID {
					continue
				}
				utws = append(utws, t.Object)
			}
			var fakex, fakey float64 = hs.game.Store.Map.GetRandomSpawnCoordinatesForLineID(cp.LineID)
			utws = append(utws, utils.Object{
				X: hst.SelectedTower.X + cs.X,
				Y: hst.SelectedTower.Y + cs.Y,
				H: hst.SelectedTower.H, W: hst.SelectedTower.W,
			})
			steps := hs.game.Store.Units.Astar(hs.game.Store.Map, cp.LineID, utils.MovingObject{
				Object: utils.Object{
					X: fakex,
					Y: fakey,
					W: 1, H: 1,
				},
			}, utws)
			if len(steps) != 0 {
				actionDispatcher.PlaceTower(hst.SelectedTower.Type, cp.ID, int(hst.SelectedTower.X+cs.X), int(hst.SelectedTower.Y+cs.Y))
			}
			return nil
		}
		for _, t := range tws {
			if clickAbsolute.IsColliding(t.Object) && cp.ID == t.PlayerID {
				if hst.TowerOpenMenuID != "" {
					// When the user clicks 2 times on the same tower we remove it
					if t.ID == hst.TowerOpenMenuID {
						actionDispatcher.RemoveTower(cp.ID, t.ID, t.Type)
						actionDispatcher.CloseTowerMenu()
						return nil
					}
				} else {
					actionDispatcher.OpenTowerMenu(t.ID)
					return nil
				}
			}
		}
		// If we are here no Tower was clicked but a click action was done,
		// so we check if the TowerOpenMenuID is set to unset it as this was
		// a click-off
		if hst.TowerOpenMenuID != "" {
			actionDispatcher.CloseTowerMenu()
		}
	}

	for ut, kb := range unitKeybinds {
		if cp.CanSummonUnit(ut) && inpututil.IsKeyJustPressed(kb) {
			actionDispatcher.SummonUnit(ut, cp.ID, cp.LineID, hs.game.Store.Map.GetNextLineID(cp.LineID))
			return nil
		}
	}
	for tt, kb := range towerKeybinds {
		if cp.CanPlaceTower(tt) && inpututil.IsKeyJustPressed(kb) {
			actionDispatcher.SelectTower(tt, x, y)
			return nil
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		actionDispatcher.GoHome()
	}
	if hst.TowerOpenMenuID != "" {
		if hs.input.IsKeyJustPressed(ebiten.KeyEscape) {
			actionDispatcher.CloseTowerMenu()
		}
	}
	if hst.SelectedTower != nil {
		if hs.input.IsMouseButtonJustPressed(ebiten.MouseButtonRight) || hs.input.IsKeyJustPressed(ebiten.KeyEscape) {
			actionDispatcher.DeselectTower(hst.SelectedTower.Type)
		} else {
			var invalid bool
			utws := make([]utils.Object, 0, 0)
			// TODO: Check why the selected towers can be placed in top of other already palced towers
			for _, t := range tws {
				// If the tower does not belong to the current user then we can skip
				// as it's outside the Players Building Zone
				if t.PlayerID != cp.ID {
					continue
				}
				utws = append(utws, t.Object)
				// The t.Object has the X and Y relative to the map
				// and the hst.SelectedTower has them relative to the
				// screen so we need to port the t.Object to the same
				// relative values
				neo := t.Object
				neo.X -= cs.X
				neo.Y -= cs.Y
				if hst.SelectedTower.IsColliding(neo) {
					invalid = true
					break
				}
			}
			neo := hst.SelectedTower.Object
			neo.X += cs.X
			neo.Y += cs.Y
			if !hs.game.Store.Map.IsInValidBuildingZone(neo, hst.SelectedTower.LineID) {
				invalid = true
			}

			if !invalid {
				units := hs.game.Store.Units.List()
				for _, u := range units {
					if u.PlayerID != cp.ID {
						continue
					}
					if u.IsColliding(neo) {
						invalid = true
						break
					}
				}
			}

			// Only check if the line is blocked when is still valid position and it has not moved.
			// TODO: We can improve this by storing this result (if blocking or not) so we only validate
			// this once and not when the mouse is static with a selected tower
			if !invalid && (hst.LastCursorPosition.X == float64(x) && hst.LastCursorPosition.Y == float64(y) && !hst.CheckedPath) {
				//var fakex, fakey float64 = hs.game.Store.Map.GetRandomSpawnCoordinatesForLineID(cp.LineID)
				//utws = append(utws, utils.Object{
				//X: hst.SelectedTower.X + cs.X,
				//Y: hst.SelectedTower.Y + cs.Y,
				//H: hst.SelectedTower.H, W: hst.SelectedTower.W,
				//})
				//steps := hs.game.Store.Units.Astar(hs.game.Store.Map, cp.LineID, utils.MovingObject{
				//Object: utils.Object{
				//X: fakex,
				//Y: fakey,
				//W: 1, H: 1,
				//},
				//}, utws)
				//if len(steps) == 0 {
				//invalid = true
				//}
				//actionDispatcher.CheckedPath(true)
			}
			if invalid != hst.SelectedTower.Invalid {
				actionDispatcher.SelectedTowerInvalid(invalid)
			}
		}
	}

	return nil
}

func (hs *HUDStore) Draw(screen *ebiten.Image) {

	hst := hs.GetState().(HUDState)
	cs := hs.game.Camera.GetState().(CameraState)
	cp := hs.game.Store.Players.FindCurrent()

	psit := hs.game.Store.Players.GetState().(store.PlayersState).IncomeTimer
	entries := make([]any, 0, 0)
	entries = append(entries,
		fmt.Sprintf("%s %s %s",
			fillIn("Name", 10),
			fillIn("Lives", 8),
			fillIn("Income", 8)),
	)

	var sortedPlayers = make([]*store.Player, 0, 0)
	for _, p := range hs.game.Store.Players.List() {
		sortedPlayers = append(sortedPlayers, p)
	}
	sort.Slice(sortedPlayers, func(i, j int) bool {
		ii := sortedPlayers[i]
		jj := sortedPlayers[j]
		if ii.Income != jj.Income {
			return ii.Income > jj.Income
		}
		return ii.LineID < jj.LineID
	})
	for _, p := range sortedPlayers {
		entries = append(entries,
			fmt.Sprintf("%s %s %s",
				fillIn(p.Name, 10),
				fillIn(strconv.Itoa(p.Lives), 8),
				fillIn(strconv.Itoa(p.Income), 8)),
		)
	}
	hs.statsListW.SetEntries(entries)

	visibility := widget.Visibility_Show
	if !hst.ShowStats {
		visibility = widget.Visibility_Hide_Blocking
	}
	hs.statsListW.GetWidget().Visibility = visibility
	hs.incomeTextW.Label = fmt.Sprintf("Gold: %s Income Timer: %ds", fillIn(strconv.Itoa(cp.Gold), 5), psit)

	wuts := hs.unitsC.Children()
	for i, u := range sortedUnits() {
		wuts[i].GetWidget().Disabled = !cp.CanSummonUnit(u.Type.String())
	}

	wtws := hs.towersC.Children()
	for i, t := range sortedTowers() {
		wtws[i].GetWidget().Disabled = !cp.CanPlaceTower(t.Type.String())
	}

	if cp.Lives == 0 {
		hs.winLoseTextW.Label = "YOU LOST"
		hs.winLoseTextW.GetWidget().Visibility = widget.Visibility_Show
	}

	if cp.Winner {
		hs.winLoseTextW.Label = "YOU WON!"
		hs.winLoseTextW.GetWidget().Visibility = widget.Visibility_Show
	}

	hs.ui.Draw(screen)

	if hst.SelectedTower != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(hst.SelectedTower.X/cs.Zoom, hst.SelectedTower.Y/cs.Zoom)
		op.GeoM.Scale(cs.Zoom, cs.Zoom)

		if hst.SelectedTower != nil && hst.SelectedTower.Invalid {
			op.ColorM.Scale(2, 0.5, 0.5, 0.9)
		}

		screen.DrawImage(ebiten.NewImageFromImage(hst.SelectedTower.Faceset()), op)
	}
}

func fillIn(s string, l int) string {
	ss := make([]string, l, l)
	for i, v := range s {
		ss[i] = string(v)
	}
	for i, v := range ss {
		if string(v) == "" {
			ss[i] = " "
		}
	}
	return strings.Join(ss, "")
}

func (hs *HUDStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	hstate, ok := state.(HUDState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.SelectTower:
		hs.GetDispatcher().WaitFor(hs.game.Store.Players.GetDispatcherToken())
		cp := hs.game.Store.Players.FindCurrent()
		cs := hs.game.Camera.GetState().(CameraState)
		x, y := fixPosition(cs, float64(act.SelectTower.X), float64(act.SelectTower.Y))
		hstate.SelectedTower = &SelectedTower{
			Tower: store.Tower{
				Object: utils.Object{
					// The Buttons have 16*2 so we want to place it on the middle so just 16
					X: x,
					Y: y,
					W: 32,
					H: 32,
				},
				Type:   act.SelectTower.Type,
				LineID: cp.LineID,
			},
		}
	case action.CursorMove:
		// We update the last seen cursor position to not resend unnecessary events
		nx := float64(act.CursorMove.X)
		ny := float64(act.CursorMove.Y)

		hstate.LastCursorPosition.X = nx
		hstate.LastCursorPosition.Y = ny

		if hstate.SelectedTower != nil {
			cs := hs.game.Camera.GetState().(CameraState)

			hstate.SelectedTower.X, hstate.SelectedTower.Y = fixPosition(cs, nx, ny)
		}
		// If it has moved we set the CheckedPath as not checked as it's only checked
		// when the Cursor has not moved
		hstate.CheckedPath = false
	case action.PlaceTower, action.DeselectTower:
		hstate.SelectedTower = nil
	case action.SelectedTowerInvalid:
		if hstate.SelectedTower != nil {
			hstate.SelectedTower.Invalid = act.SelectedTowerInvalid.Invalid
		}
	case action.OpenTowerMenu:
		hstate.TowerOpenMenuID = act.OpenTowerMenu.TowerID
	case action.CloseTowerMenu:
		hstate.TowerOpenMenuID = ""
	case action.CheckedPath:
		hstate.CheckedPath = act.CheckedPath.Checked
	case action.ToggleStats:
		hstate.ShowStats = !hstate.ShowStats
	default:
	}

	return hstate
}

func fixPosition(cs CameraState, x, y float64) (float64, float64) {
	//cs := hs.game.Camera.GetState().(CameraState)

	absnx := x + cs.X
	absny := y + cs.Y
	// We find the closes multiple in case the cursor moves too fast, between FPS reloads,
	// and lands in a position not 'multiple' which means the position of the SelectedTower
	// is not updated and the result is the cursor far away from the Drawing of the SelectedTower
	// as it has stayed on the previous position
	var multiple int = 16
	if int(absnx)%multiple == 0 {
		x -= 16
	} else {
		x = float64(closestMultiple(int(absnx), multiple)) - 16 - cs.X
	}
	if int(absny)%multiple == 0 {
		y -= 16
	} else {
		y = float64(closestMultiple(int(absny), multiple)) - 16 - cs.Y
	}

	return x, y
}

func sortedUnits() []*unit.Unit {
	us := make([]*unit.Unit, 0, 0)
	for _, u := range unit.Units {
		us = append(us, u)
	}
	sort.Slice(us, func(i, j int) bool {
		return us[i].Gold < us[j].Gold
	})
	return us
}

func sortedTowers() []*tower.Tower {
	ts := make([]*tower.Tower, 0, 0)
	for _, t := range tower.Towers {
		ts = append(ts, t)
	}
	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Type > ts[j].Type
	})
	return ts
}

// closestMultiple finds the coses multiple of 'b' for the number 'a'
func closestMultiple(a, b int) int {
	a = a + b/2
	a = a - (a % b)
	return a
}

func (hs *HUDStore) buildUI() {
	topRightContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	topRightVerticalRowC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)

	topRightVerticalRowWraperC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
	)

	topRightHorizontalRowC := widget.NewContainer(
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

	homeBtnW := widget.NewButton(
		widget.ButtonOpts.Image(buttonImage),

		widget.ButtonOpts.Text("HOME(F1)", smallFont, &widget.ButtonTextColor{
			Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
		}),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			actionDispatcher.GoHome()
		}),
	)

	statsBtnW := widget.NewButton(
		widget.ButtonOpts.Image(buttonImage),

		widget.ButtonOpts.Text("STATS", smallFont, &widget.ButtonTextColor{
			Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
		}),

		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			actionDispatcher.ToggleStats()
		}),
	)

	topRightStatsC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
	)

	entries := make([]any, 0, 0)
	statsListW := widget.NewList(
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
			}, buttonImage),
			widget.SliderOpts.MinHandleSize(5),
			// Set how wide the track should be
			widget.SliderOpts.TrackPadding(widget.NewInsetsSimple(2)),
		),
		// Hide the horizontal slider
		widget.ListOpts.HideHorizontalSlider(),
		widget.ListOpts.HideVerticalSlider(),
		// Set the font for the list options
		widget.ListOpts.EntryFontFace(smallFont),
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
			return e.(string)
		}),
		// Padding for each entry
		widget.ListOpts.EntryTextPadding(widget.NewInsetsSimple(5)),
		// Text position for each entry
		widget.ListOpts.EntryTextPosition(widget.TextPositionStart, widget.TextPositionCenter),
		// This handler defines what function to run when a list item is selected.
		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			//entry := args.Entry.(ListEntry)
			//fmt.Println("Entry Selected: ", entry)
		}),
	)

	incomeTextW := widget.NewText(
		widget.TextOpts.Text("Gold: 40     Income Timer: 15s", smallFont, color.White),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)

	bottomRightContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	// Create the first tab
	// A TabBookTab is a labelled container. The text here is what will show up in the tab button
	tabUnits := widget.NewTabBookTab("UNITS",
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 120, A: 255})),
	)

	unitsC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(5),
			//Define how much padding to inset the child content
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(6)),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(5, 5),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{false, false, false, false, false}, []bool{false, false, false, false, false}),
		)),
	)
	for _, u := range sortedUnits() {
		tooltipContainer := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
			widget.ContainerOpts.AutoDisableChildren(),
			widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 230, A: 255})),
		)

		toolTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf("Gold: %d\nHP: %.0f\nIncome: %d\nKeybind: %s", u.Gold, u.Health, u.Income, u.Keybind), smallFont, color.White),
			widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
		)
		tooltipContainer.AddChild(toolTxt)

		ubtn := widget.NewButton(
			// set general widget options
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					MaxWidth:  38,
					MaxHeight: 38,
				}),
				widget.WidgetOpts.ToolTip(widget.NewToolTip(
					widget.ToolTipOpts.Content(tooltipContainer),
					//widget.WidgetToolTipOpts.Delay(1*time.Second),
					widget.ToolTipOpts.Offset(stdimage.Point{-5, 5}),
					widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
					//When the Position is set to TOOLTIP_POS_WIDGET, you can configure where it opens with the optional parameters below
					//They will default to what you see below if you do not provide them
					widget.ToolTipOpts.WidgetOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
					widget.ToolTipOpts.WidgetOriginVertical(widget.TOOLTIP_ANCHOR_END),
					widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
					widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_START),
				)),
			),

			// specify the images to sue
			widget.ButtonOpts.Image(buttonImageFromImage(u.Faceset)),

			// add a handler that reacts to clicking the button
			widget.ButtonOpts.ClickedHandler(func(u *unit.Unit) func(args *widget.ButtonClickedEventArgs) {
				return func(args *widget.ButtonClickedEventArgs) {
					cp := hs.game.Store.Players.FindCurrent()
					actionDispatcher.SummonUnit(u.Type.String(), cp.ID, cp.LineID, hs.game.Store.Map.GetNextLineID(cp.LineID))
				}
			}(u)),
		)
		unitsC.AddChild(ubtn)
	}
	hs.unitsC = unitsC
	tabUnits.AddChild(unitsC)

	tabTowers := widget.NewTabBookTab("TOWERS",
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 120, A: 255})),
	)
	towersC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(1),
			//Define how much padding to inset the child content
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(6)),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(5, 5),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{false, false, false, false, false}, []bool{false, false, false, false, false}),
		)),
	)
	for _, t := range sortedTowers() {
		tooltipContainer := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
			widget.ContainerOpts.AutoDisableChildren(),
			widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 230, A: 255})),
		)

		toolTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf("Gold: %d\nRange: %.0f\nDamage: %.0f\nKeybind: %s", t.Gold, t.Range, t.Damage, t.Keybind), smallFont, color.White),
			widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
		)
		tooltipContainer.AddChild(toolTxt)
		tbtn := widget.NewButton(
			// set general widget options
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					MaxWidth:  38,
					MaxHeight: 38,
				}),
				widget.WidgetOpts.ToolTip(widget.NewToolTip(
					widget.ToolTipOpts.Content(tooltipContainer),
					//widget.WidgetToolTipOpts.Delay(1*time.Second),
					widget.ToolTipOpts.Offset(stdimage.Point{-5, 5}),
					widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
					//When the Position is set to TOOLTIP_POS_WIDGET, you can configure where it opens with the optional parameters below
					//They will default to what you see below if you do not provide them
					widget.ToolTipOpts.WidgetOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
					widget.ToolTipOpts.WidgetOriginVertical(widget.TOOLTIP_ANCHOR_END),
					widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_END),
					widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_START),
				)),
			),

			// specify the images to sue
			widget.ButtonOpts.Image(buttonImageFromImage(t.Faceset)),

			// add a handler that reacts to clicking the button
			widget.ButtonOpts.ClickedHandler(func(t *tower.Tower) func(args *widget.ButtonClickedEventArgs) {
				return func(args *widget.ButtonClickedEventArgs) {
					hst := hs.GetState().(HUDState)
					actionDispatcher.SelectTower(t.Type.String(), int(hst.LastCursorPosition.X), int(hst.LastCursorPosition.Y))
				}
			}(t)),
		)
		towersC.AddChild(tbtn)
	}
	hs.towersC = towersC
	tabTowers.AddChild(towersC)

	tabBook := widget.NewTabBook(
		widget.TabBookOpts.TabButtonImage(buttonImage),
		widget.TabBookOpts.TabButtonText(smallFont, &widget.ButtonTextColor{Idle: color.White, Disabled: color.White}),
		widget.TabBookOpts.TabButtonSpacing(0),
		widget.TabBookOpts.ContainerOpts(
			widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
			})),
		),
		widget.TabBookOpts.TabButtonOpts(
			widget.ButtonOpts.TextPadding(widget.NewInsetsSimple(5)),
			widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(98, 0)),
		),
		widget.TabBookOpts.Tabs(tabUnits, tabTowers),
	)
	bottomRightContainer.AddChild(tabBook)

	hs.incomeTextW = incomeTextW
	hs.statsListW = statsListW

	topRightStatsC.AddChild(incomeTextW)
	topRightStatsC.AddChild(statsListW)

	topRightHorizontalRowC.AddChild(statsBtnW)
	topRightHorizontalRowC.AddChild(homeBtnW)
	topRightVerticalRowWraperC.AddChild(topRightHorizontalRowC)
	topRightVerticalRowC.AddChild(topRightVerticalRowWraperC)
	topRightVerticalRowC.AddChild(topRightStatsC)
	topRightContainer.AddChild(topRightVerticalRowC)

	topLeftBtnContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	leaveBtnW := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),

		widget.ButtonOpts.Image(buttonImage),

		widget.ButtonOpts.Text("LEAVE", smallFont, &widget.ButtonTextColor{
			Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
		}),

		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			u := hs.game.Store.Players.FindCurrent()
			actionDispatcher.RemovePlayer(u.ID)
		}),
	)
	topLeftBtnContainer.AddChild(leaveBtnW)

	centerTextContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	winLoseTextW := widget.NewText(
		widget.TextOpts.Text("", smallFont, color.White),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)
	centerTextContainer.AddChild(winLoseTextW)
	winLoseTextW.GetWidget().Visibility = widget.Visibility_Hide
	hs.winLoseTextW = winLoseTextW

	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout(widget.StackedLayoutOpts.Padding(widget.NewInsetsSimple(25)))),
	)

	rootContainer.AddChild(topRightContainer)
	rootContainer.AddChild(topLeftBtnContainer)
	rootContainer.AddChild(bottomRightContainer)
	rootContainer.AddChild(centerTextContainer)

	hs.ui = &ebitenui.UI{
		Container: rootContainer,
	}
}
