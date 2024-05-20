package game

import (
	"fmt"
	stdimage "image"
	"image/color"
	"sort"
	"strconv"
	"time"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	cutils "github.com/xescugc/maze-wars/client/utils"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/unit"
	"github.com/xescugc/maze-wars/utils"
)

const (
	unitToolTipTmpl       = "Gold: %d\nHP: %.0f\nIncome: %d\nEnv: %s\nKeybind: %s"
	unitUpdateToolTipTmpl = "Cost: %d\nGold: %d\nHP: %.0f\nIncome: %d"

	towerRemoveToolTipTmpl = "Gold back: %d\nKeybind: %s"
	towerUpdateToolTipTmpl = "Cost: %d\nDamage: %.1f\nKeybind: %s"
	towerUpdateLimit       = "Tower is at it's max level"
)

// HUDStore is in charge of keeping track of all the elements
// on the player HUD that are static and always seen
type HUDStore struct {
	*flux.ReduceStore

	game *Game

	ui *ebitenui.UI

	statsListW   *widget.List
	incomeTextW  *widget.Text
	winLoseTextW *widget.Text

	unitsC       *widget.Container
	unitsTooltip map[string]*widget.Text

	unitUpdatesC       *widget.Container
	unitUpdatesTooltip map[string]*widget.Text

	towersC *widget.Container

	bottomLeftContainer *widget.Container
	towerMenuContainer  *widget.Container
	towerRemoveToolTip  *widget.Text
	towerUpdateToolTip  *widget.Text
	towerUpdateButton   *widget.Button
}

// HUDState stores the HUD state
type HUDState struct {
	SelectedTower *SelectedTower
	OpenTowerMenu *store.Tower

	LastCursorPosition utils.Object

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

	removeTowerKeybind = ebiten.KeyD
	updateTowerKeybind = ebiten.KeyF
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
func NewHUDStore(d *flux.Dispatcher, g *Game) (*HUDStore, error) {
	hs := &HUDStore{
		game:               g,
		unitsTooltip:       make(map[string]*widget.Text),
		unitUpdatesTooltip: make(map[string]*widget.Text),
	}
	hs.ReduceStore = flux.NewReduceStore(d, hs.Reduce, HUDState{
		ShowStats: true,
	})

	hs.buildUI()

	return hs, nil
}

func (hs *HUDStore) Update() error {
	b := time.Now()
	defer utils.LogTime(hs.game.Logger, b, "hud update")

	hs.ui.Update()

	cs := hs.game.Camera.GetState().(CameraState)
	hst := hs.GetState().(HUDState)
	x, y := ebiten.CursorPosition()
	cp := hs.game.Store.Lines.FindCurrentPlayer()
	cl := hs.game.Store.Lines.FindLineByID(cp.LineID)
	tws := cl.Towers
	// Only send a CursorMove when the curso has actually moved
	if hst.LastCursorPosition.X != x || hst.LastCursorPosition.Y != y {
		actionDispatcher.CursorMove(x, y)
	}
	// If the Current player is dead or has no more lives there are no
	// mo actions that can be done
	if cp.Lives == 0 || cp.Winner {
		return nil
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		clickAbsolute := utils.Object{
			X: x + cs.X,
			Y: y + cs.Y,
			W: 1, H: 1,
		}

		if hst.SelectedTower != nil && !hst.SelectedTower.Invalid {
			actionDispatcher.PlaceTower(hst.SelectedTower.Type, cp.ID, hst.SelectedTower.X+cs.X, hst.SelectedTower.Y+cs.Y)
			return nil
		}
		for _, t := range tws {
			if clickAbsolute.IsColliding(t.Object) && cp.ID == t.PlayerID {
				actionDispatcher.OpenTowerMenu(t.ID)
				return nil
			}
		}
		// If we are here no Tower was clicked but a click action was done,
		// so we check if the OpenTowerMenu is set to unset it as this was
		// a click-off
		if hst.OpenTowerMenu != nil {
			p := stdimage.Point{x, y}
			w := hs.towerMenuContainer.GetWidget()
			inside := p.In(w.Rect)
			layer := w.EffectiveInputLayer()

			clickedInside := inside && input.MouseButtonPressedLayer(ebiten.MouseButtonLeft, layer)

			if !clickedInside {
				actionDispatcher.CloseTowerMenu()
			}
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
	if hst.OpenTowerMenu != nil {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			actionDispatcher.CloseTowerMenu()
		}
		if inpututil.IsKeyJustPressed(removeTowerKeybind) {
			actionDispatcher.RemoveTower(cp.ID, hst.OpenTowerMenu.ID)
			actionDispatcher.CloseTowerMenu()
		}
		if inpututil.IsKeyJustPressed(updateTowerKeybind) {
			actionDispatcher.UpdateTower(cp.ID, hst.OpenTowerMenu.ID)
		}
	}
	if hst.SelectedTower != nil {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			actionDispatcher.DeselectTower(hst.SelectedTower.Type)
		} else {
			invalid := !cp.CanPlaceTower(hst.SelectedTower.Type)

			neo := hst.SelectedTower.Object
			neo.X += cs.X
			neo.Y += cs.Y

			if !invalid {
				invalid = !cl.Graph.CanAddTower(neo.X, neo.Y, neo.W, neo.H)
			}

			if !invalid {
				for _, u := range cl.Units {
					if u.IsColliding(neo) {
						invalid = true
						break
					}
				}
			}

			if invalid != hst.SelectedTower.Invalid {
				actionDispatcher.SelectedTowerInvalid(invalid)
			}
		}
	}

	return nil
}

func (hs *HUDStore) Draw(screen *ebiten.Image) {
	b := time.Now()
	defer utils.LogTime(hs.game.Logger, b, "hud draw")

	hst := hs.GetState().(HUDState)
	cs := hs.game.Camera.GetState().(CameraState)
	cp := hs.game.Store.Lines.FindCurrentPlayer()

	psit := hs.game.Store.Lines.GetState().(store.LinesState).IncomeTimer
	entries := make([]any, 0, 0)
	entries = append(entries,
		fmt.Sprintf("%s %s %s",
			cutils.FillIn("Name", 10),
			cutils.FillIn("Lives", 8),
			cutils.FillIn("Income", 8)),
	)

	var sortedPlayers = make([]*store.Player, 0, 0)
	for _, p := range hs.game.Store.Lines.ListPlayers() {
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
				cutils.FillIn(p.Name, 10),
				cutils.FillIn(strconv.Itoa(p.Lives), 8),
				cutils.FillIn(strconv.Itoa(p.Income), 8)),
		)
	}
	hs.statsListW.SetEntries(entries)

	visibility := widget.Visibility_Show
	if !hst.ShowStats {
		visibility = widget.Visibility_Hide_Blocking
	}
	hs.statsListW.GetWidget().Visibility = visibility
	hs.incomeTextW.Label = fmt.Sprintf("Gold: %s Income Timer: %ds", cutils.FillIn(strconv.Itoa(cp.Gold), 5), psit)

	wuts := hs.unitsC.Children()
	for i, u := range sortedUnits() {
		uu := cp.UnitUpdates[u.Type.String()]
		wuts[i].GetWidget().Disabled = !cp.CanSummonUnit(u.Type.String())
		hs.unitsTooltip[u.Type.String()].Label = fmt.Sprintf(unitToolTipTmpl, uu.Current.Gold, uu.Current.Health, uu.Current.Income, u.Environment, u.Keybind)
	}

	wuuts := hs.unitUpdatesC.Children()
	for i, u := range sortedUnits() {
		uu := cp.UnitUpdates[u.Type.String()]
		wuuts[i].GetWidget().Disabled = !cp.CanUpdateUnit(u.Type.String())
		hs.unitUpdatesTooltip[u.Type.String()].Label = fmt.Sprintf(unitUpdateToolTipTmpl, uu.UpdateCost, uu.Next.Gold, uu.Next.Health, uu.Next.Income)
	}

	wtws := hs.towersC.Children()
	for i, t := range sortedTowers() {
		wtws[i].GetWidget().Disabled = !cp.CanPlaceTower(t.Type.String())
	}

	if cp.Lives == 0 {
		hs.winLoseTextW.Label = "YOU LOST"
		hs.winLoseTextW.GetWidget().Visibility = widget.Visibility_Show
	} else if hs.winLoseTextW.GetWidget().Visibility != widget.Visibility_Hide {
		hs.winLoseTextW.Label = ""
		hs.winLoseTextW.GetWidget().Visibility = widget.Visibility_Hide
	}

	if cp.Winner {
		hs.winLoseTextW.Label = "YOU WON!"
		hs.winLoseTextW.GetWidget().Visibility = widget.Visibility_Show
	} else if hs.winLoseTextW.GetWidget().Visibility != widget.Visibility_Hide {
		hs.winLoseTextW.Label = ""
		hs.winLoseTextW.GetWidget().Visibility = widget.Visibility_Hide
	}

	if hst.OpenTowerMenu != nil {
		// TODO: Add the keybind to the tooltip
		hs.bottomLeftContainer.GetWidget().Visibility = widget.Visibility_Show
		tu := tower.FindUpdateByLevel(hst.OpenTowerMenu.Type, hst.OpenTowerMenu.Level+1)
		tc := tower.FindUpdateByLevel(hst.OpenTowerMenu.Type, hst.OpenTowerMenu.Level)
		removeTowerGoldReturn := tower.Towers[hst.OpenTowerMenu.Type].Gold / 2
		if tc != nil {
			removeTowerGoldReturn = tc.UpdateCost / 2
		}
		// Where there is no more updates we disable the button
		if tu == nil {
			hs.towerUpdateButton.GetWidget().Disabled = true
			hs.towerUpdateToolTip.Label = towerUpdateLimit
		} else {
			hs.towerUpdateButton.GetWidget().Disabled = cp.Gold < tu.UpdateCost
			hs.towerUpdateToolTip.Label = fmt.Sprintf(towerUpdateToolTipTmpl, tu.UpdateCost, tu.Stats.Damage, updateTowerKeybind)
		}
		hs.towerRemoveToolTip.Label = fmt.Sprintf(towerRemoveToolTipTmpl, removeTowerGoldReturn, removeTowerKeybind)
	} else {
		hs.bottomLeftContainer.GetWidget().Visibility = widget.Visibility_Hide
	}

	hs.ui.Draw(screen)

	if hst.SelectedTower != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(hst.SelectedTower.X)/cs.Zoom, float64(hst.SelectedTower.Y)/cs.Zoom)
		op.GeoM.Scale(cs.Zoom, cs.Zoom)

		if hst.SelectedTower != nil && hst.SelectedTower.Invalid {
			op.ColorM.Scale(2, 0.5, 0.5, 0.9)
		}

		screen.DrawImage(imagesCache.Get(hst.SelectedTower.FacetKey()), op)
	}
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
		cp := hs.game.Store.Lines.FindCurrentPlayer()
		cs := hs.game.Camera.GetState().(CameraState)
		x, y := fixPosition(cs, act.SelectTower.X, act.SelectTower.Y)
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
		nx := act.CursorMove.X
		ny := act.CursorMove.Y

		hstate.LastCursorPosition.X = nx
		hstate.LastCursorPosition.Y = ny

		if hstate.SelectedTower != nil {
			cs := hs.game.Camera.GetState().(CameraState)

			hstate.SelectedTower.X, hstate.SelectedTower.Y = fixPosition(cs, nx, ny)
		}
	case action.PlaceTower, action.DeselectTower:
		hstate.SelectedTower = nil
	case action.SelectedTowerInvalid:
		if hstate.SelectedTower != nil {
			hstate.SelectedTower.Invalid = act.SelectedTowerInvalid.Invalid
		}
	case action.OpenTowerMenu:
		hstate.OpenTowerMenu = hs.findTowerByID(act.OpenTowerMenu.TowerID)
	case action.UpdateTower:
		hs.GetDispatcher().WaitFor(hs.game.Store.Lines.GetDispatcherToken())

		// As the UpdateTower is done we need to update the OpenTowerMenu
		// so we can display the new information
		hstate.OpenTowerMenu = hs.findTowerByID(act.UpdateTower.TowerID)
	case action.CloseTowerMenu:
		hstate.OpenTowerMenu = nil
	case action.ToggleStats:
		hstate.ShowStats = !hstate.ShowStats
	default:
	}

	return hstate
}

func (hs *HUDStore) findTowerByID(tid string) *store.Tower {
	for _, l := range hs.game.Store.Lines.ListLines() {
		if t, ok := l.Towers[tid]; ok {
			return t
		}
	}
	return nil
}

func fixPosition(cs CameraState, x, y int) (int, int) {
	absnx := x + cs.X
	absny := y + cs.Y
	// We find the closes multiple in case the cursor moves too fast, between FPS reloads,
	// and lands in a position not 'multiple' which means the position of the SelectedTower
	// is not updated and the result is the cursor far away from the Drawing of the SelectedTower
	// as it has stayed on the previous position
	var multiple int = 16
	// If it's == 0 means it's exact but as we want to center it we remove 16 (towers are 32)
	// If it's !=0 then we find what's the remaning for
	if absnx%multiple == 0 {
		x -= 16
	} else {
		x = utils.ClosestMultiple(absnx, multiple) - 16 - cs.X
	}
	if absny%multiple == 0 {
		y -= 16
	} else {
		y = utils.ClosestMultiple(absny, multiple) - 16 - cs.Y
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

func (hs *HUDStore) buildUI() {
	cp := hs.game.Store.Lines.FindCurrentPlayer()
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
		widget.ButtonOpts.Image(cutils.ButtonImage),

		widget.ButtonOpts.Text("HOME(F1)", cutils.SmallFont, &widget.ButtonTextColor{
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
		widget.ButtonOpts.Image(cutils.ButtonImage),

		widget.ButtonOpts.Text("STATS", cutils.SmallFont, &widget.ButtonTextColor{
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
			}, cutils.ButtonImage),
			widget.SliderOpts.MinHandleSize(5),
			// Set how wide the track should be
			widget.SliderOpts.TrackPadding(widget.NewInsetsSimple(2)),
		),
		// Hide the horizontal slider
		widget.ListOpts.HideHorizontalSlider(),
		widget.ListOpts.HideVerticalSlider(),
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
		widget.TextOpts.Text("Gold: 40     Income Timer: 15s", cutils.SmallFont, color.White),
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
		uu := cp.UnitUpdates[u.Type.String()]

		tooltipContainer := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
			widget.ContainerOpts.AutoDisableChildren(),
			widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 230, A: 255})),
		)

		toolTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf(unitToolTipTmpl, uu.Current.Gold, uu.Current.Health, uu.Current.Income, u.Environment, u.Keybind), cutils.SmallFont, color.White),
			widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
		)
		hs.unitsTooltip[u.Type.String()] = toolTxt
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
			widget.ButtonOpts.Image(cutils.ButtonImageFromImage(imagesCache.Get(u.FacesetKey()))),

			// add a handler that reacts to clicking the button
			widget.ButtonOpts.ClickedHandler(func(u *unit.Unit) func(args *widget.ButtonClickedEventArgs) {
				return func(args *widget.ButtonClickedEventArgs) {
					cp := hs.game.Store.Lines.FindCurrentPlayer()
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
			widget.TextOpts.Text(fmt.Sprintf("Gold: %d\nRange: %.0f\nDamage: %.0f\nTargets: %s\nKeybind: %s", t.Gold, t.Range, t.Damage, t.Targets, t.Keybind), cutils.SmallFont, color.White),
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
			widget.ButtonOpts.Image(cutils.ButtonImageFromImage(imagesCache.Get(t.FacesetKey()))),

			// add a handler that reacts to clicking the button
			widget.ButtonOpts.ClickedHandler(func(t *tower.Tower) func(args *widget.ButtonClickedEventArgs) {
				return func(args *widget.ButtonClickedEventArgs) {
					hst := hs.GetState().(HUDState)
					actionDispatcher.SelectTower(t.Type.String(), hst.LastCursorPosition.X, hst.LastCursorPosition.Y)
				}
			}(t)),
		)
		towersC.AddChild(tbtn)
	}
	hs.towersC = towersC
	tabTowers.AddChild(towersC)

	// Create the first tab
	// A TabBookTab is a labelled container. The text here is what will show up in the tab button
	tabUnitsUpdates := widget.NewTabBookTab("UNITS UPDATES",
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 120, A: 255})),
	)

	unitUpdatesC := widget.NewContainer(
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
		uu := cp.UnitUpdates[u.Type.String()]

		tooltipContainer := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
			widget.ContainerOpts.AutoDisableChildren(),
			widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 230, A: 255})),
		)

		toolTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf(unitUpdateToolTipTmpl, uu.UpdateCost, uu.Next.Gold, uu.Next.Health, uu.Next.Income), cutils.SmallFont, color.White),
			widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
		)
		hs.unitUpdatesTooltip[u.Type.String()] = toolTxt
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
			widget.ButtonOpts.Image(cutils.ButtonImageFromImage(imagesCache.Get(u.FacesetKey()))),

			// add a handler that reacts to clicking the button
			widget.ButtonOpts.ClickedHandler(func(u *unit.Unit) func(args *widget.ButtonClickedEventArgs) {
				return func(args *widget.ButtonClickedEventArgs) {
					cp := hs.game.Store.Lines.FindCurrentPlayer()
					actionDispatcher.UpdateUnit(cp.ID, u.Type.String())
				}
			}(u)),
		)
		unitUpdatesC.AddChild(ubtn)
	}
	hs.unitUpdatesC = unitUpdatesC
	tabUnitsUpdates.AddChild(unitUpdatesC)

	tabBook := widget.NewTabBook(
		widget.TabBookOpts.TabButtonImage(cutils.ButtonImage),
		widget.TabBookOpts.TabButtonText(cutils.SmallFont, &widget.ButtonTextColor{Idle: color.White, Disabled: color.White}),
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
		widget.TabBookOpts.Tabs(tabUnits, tabTowers, tabUnitsUpdates),
	)
	bottomRightContainer.AddChild(tabBook)

	bottomLeftContainer := hs.guiBottomLeft()
	bottomLeftContainer.GetWidget().Visibility = widget.Visibility_Hide

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

		widget.ButtonOpts.Image(cutils.ButtonImage),

		widget.ButtonOpts.Text("LEAVE", cutils.SmallFont, &widget.ButtonTextColor{
			Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
		}),

		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			u := hs.game.Store.Lines.FindCurrentPlayer()
			actionDispatcher.RemovePlayer(u.ID)
		}),
	)
	topLeftBtnContainer.AddChild(leaveBtnW)

	centerTextContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	winLoseTextW := widget.NewText(
		widget.TextOpts.Text("", cutils.SmallFont, color.White),
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
	rootContainer.AddChild(bottomLeftContainer)
	rootContainer.AddChild(centerTextContainer)

	hs.ui = &ebitenui.UI{
		Container: rootContainer,
	}
}

func (hs *HUDStore) guiBottomLeft() *widget.Container {
	bottomLeftContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	towerMenuC := widget.NewContainer(
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
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
			}),
		),
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 230, A: 255})),
	)
	// Remove button
	removeToolTipContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
		widget.ContainerOpts.AutoDisableChildren(),
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 230, A: 255})),
	)

	removeToolTxt := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.Text(fmt.Sprintf(towerRemoveToolTipTmpl, 0, removeTowerKeybind), cutils.SmallFont, color.White),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
	)
	hs.towerRemoveToolTip = removeToolTxt
	removeToolTipContainer.AddChild(removeToolTxt)

	rbtn := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				MaxWidth:  38,
				MaxHeight: 38,
			}),
			widget.WidgetOpts.ToolTip(widget.NewToolTip(
				widget.ToolTipOpts.Content(removeToolTipContainer),
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
		widget.ButtonOpts.Image(cutils.ButtonImageFromImage(imagesCache.Get(crossImageKey))),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func() func(args *widget.ButtonClickedEventArgs) {
			return func(args *widget.ButtonClickedEventArgs) {
				cp := hs.game.Store.Lines.FindCurrentPlayer()
				otm := hs.GetState().(HUDState).OpenTowerMenu
				actionDispatcher.RemoveTower(cp.ID, otm.ID)
				actionDispatcher.CloseTowerMenu()
			}
		}()),
	)
	towerMenuC.AddChild(rbtn)

	// Update button
	updateToolTipContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical))),
		widget.ContainerOpts.AutoDisableChildren(),
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 230, A: 255})),
	)

	updateToolTxt := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.Text(fmt.Sprintf(towerUpdateToolTipTmpl, 0, 0.0, updateTowerKeybind), cutils.SmallFont, color.White),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.MinSize(100, 0)),
	)
	hs.towerUpdateToolTip = updateToolTxt
	updateToolTipContainer.AddChild(updateToolTxt)

	ubtn := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				MaxWidth:  38,
				MaxHeight: 38,
			}),
			widget.WidgetOpts.ToolTip(widget.NewToolTip(
				widget.ToolTipOpts.Content(updateToolTipContainer),
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
		widget.ButtonOpts.Image(cutils.ButtonImageFromImage(imagesCache.Get(arrowImageKey))),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func() func(args *widget.ButtonClickedEventArgs) {
			return func(args *widget.ButtonClickedEventArgs) {
				cp := hs.game.Store.Lines.FindCurrentPlayer()
				tomid := hs.GetState().(HUDState).OpenTowerMenu.ID
				actionDispatcher.UpdateTower(cp.ID, tomid)
			}
		}()),
	)
	towerMenuC.AddChild(ubtn)
	bottomLeftContainer.AddChild(towerMenuC)
	hs.bottomLeftContainer = bottomLeftContainer
	hs.towerMenuContainer = towerMenuC
	hs.towerUpdateButton = ubtn

	return bottomLeftContainer
}
