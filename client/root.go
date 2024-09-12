package client

import (
	"fmt"
	stdimage "image"
	"image/color"
	"log/slog"
	"math"
	"sort"
	"time"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/gofrs/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	cutils "github.com/xescugc/maze-wars/client/utils"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/unit"
	"github.com/xescugc/maze-wars/utils"
)

var (
	widgetPadding = widget.NewInsetsSimple(25)
	playBtnInsets = widget.Insets{
		Left:   100,
		Right:  100,
		Top:    10,
		Bottom: 10,
	}
	isPressed = true
)

type RootStore struct {
	*flux.ReduceStore

	Store  *Store
	Logger *slog.Logger

	ui *ebitenui.UI

	usernameTextW   *widget.Text
	playBtnW        *widget.Button
	backToLobbyBtnW *widget.Button
	matchInfoTxt    *widget.Text

	setupGameW                  *widget.Window
	waitingRoomW                *widget.Window
	waitingRoomC                *widget.Container
	waitingRoomPlayers          []*widget.Container
	waitingRoomPlayersUsernames []*widget.Text
	waitingRoomSubTitle         *widget.Text

	homeC *widget.Container

	lobbiesC      *widget.Container
	lobbiesTableW *widget.Container
	lobbiesSC     *widget.ScrollContainer

	showLobbyC              *widget.Container
	showLobbyNameW          *widget.Text
	showLobbyPlayersL       *widget.List
	showLobbyOwnerW         *widget.Text
	showLobbyPlayersHeaderW *widget.Text
	showLobbyLeaveBtnW      *widget.Button
	showLobbyDeleteBtnW     *widget.Button
	showLobbyStartBtnW      *widget.Button
	showLobbyBotsC          *widget.Container

	learnC           *widget.Container
	newLobbyW        *widget.Window
	modalBackgroundC *widget.Container
}

type RootState struct {
	SetupGame bool

	Route string

	FindGame *findGame

	WaitingRoom *waitingRoom
}

type findGame struct {
	Size           int
	Ranked         bool
	VsBots         bool
	SearchingSince time.Time
}

type waitingRoom struct {
	Size         int
	Ranked       bool
	Players      []action.SyncWaitingRoomPlayersPayload
	WaitingSince time.Time
}

func NewRootStore(d *flux.Dispatcher, s *Store, l *slog.Logger) (*RootStore, error) {
	rs := &RootStore{
		Store:  s,
		Logger: l,
	}
	rs.ReduceStore = flux.NewReduceStore(d, rs.Reduce, RootState{
		Route: utils.HomeRoute,
	})

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

	rss := rs.GetState().(RootState)
	rs.usernameTextW.Label = rs.Store.Users.Username()
	if rss.SetupGame {
		rs.loadModal(rs.setupGameW)
	} else {
		rs.closeModal(rs.setupGameW)
	}

	cl := rs.Store.Lobbies.FindCurrent()

	rs.homeC.GetWidget().Visibility = widget.Visibility_Hide
	rs.lobbiesC.GetWidget().Visibility = widget.Visibility_Hide
	rs.showLobbyC.GetWidget().Visibility = widget.Visibility_Hide
	rs.learnC.GetWidget().Visibility = widget.Visibility_Hide
	rs.lobbiesSC.GetWidget().Visibility = widget.Visibility_Hide
	rs.lobbiesTableW.GetWidget().Visibility = widget.Visibility_Hide

	switch rss.Route {
	case utils.HomeRoute:
		rs.homeC.GetWidget().Visibility = widget.Visibility_Show
	case utils.LobbiesRoute, utils.NewLobbyRoute:
		rs.lobbiesC.GetWidget().Visibility = widget.Visibility_Show
		rs.lobbiesSC.GetWidget().Visibility = widget.Visibility_Show
		rs.lobbiesTableW.GetWidget().Visibility = widget.Visibility_Show

		lbs := rs.Store.Lobbies.List()

		sort.Slice(lbs, func(i, j int) bool {
			return lbs[i].ID > lbs[j].ID
		})
		if !rs.Store.Lobbies.Seen() {
			rs.addLobbiesToTable(lbs)
			actionDispatcher.SeenLobbies()
		}

		switch rss.Route {
		case utils.NewLobbyRoute:
			rs.loadModal(rs.newLobbyW)
		default:
			rs.closeModal(rs.newLobbyW)
		}
	case utils.ShowLobbyRoute:
		rs.showLobbyC.GetWidget().Visibility = widget.Visibility_Show

		entries := make([]any, 0, len(cl.Players))
		for p := range cl.Players {
			if cl.Owner == p {
				p = p + " (owner)"
			}
			entries = append(entries, cutils.ListEntry{
				ID:   p,
				Text: p,
			})
		}
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].(cutils.ListEntry).ID > entries[j].(cutils.ListEntry).ID
		})
		if !cutils.EqualListEntries(entries, rs.showLobbyPlayersL.Entries()) {
			rs.showLobbyPlayersL.SetEntries(entries)
		}

		rs.showLobbyNameW.Label = fmt.Sprintf("LOBBY: %s", cl.Name)
		rs.showLobbyPlayersHeaderW.Label = fmt.Sprintf("PLAYERS (%d/%d)", len(cl.Players), cl.MaxPlayers)

		if rs.Store.Users.Username() == cl.Owner {
			rs.showLobbyStartBtnW.GetWidget().Visibility = widget.Visibility_Show
			rs.showLobbyStartBtnW.GetWidget().Disabled = len(cl.Players) < 2

			rs.showLobbyDeleteBtnW.GetWidget().Visibility = widget.Visibility_Show

			rs.showLobbyLeaveBtnW.GetWidget().Visibility = widget.Visibility_Hide

			rs.showLobbyBotsC.GetWidget().Visibility = widget.Visibility_Show
		} else {
			rs.showLobbyStartBtnW.GetWidget().Visibility = widget.Visibility_Hide
			rs.showLobbyDeleteBtnW.GetWidget().Visibility = widget.Visibility_Hide

			rs.showLobbyLeaveBtnW.GetWidget().Visibility = widget.Visibility_Show

			rs.showLobbyBotsC.GetWidget().Visibility = widget.Visibility_Hide
		}
	case utils.LearnRoute:
		rs.learnC.GetWidget().Visibility = widget.Visibility_Show
	}

	if rss.FindGame != nil {
		rs.playBtnW.Text().Label = "CANCEL"
		rs.matchInfoTxt.Label = fmt.Sprintf("%s %s", fmtVsRankedBots(rss.FindGame.Size, rss.FindGame.Ranked, rss.FindGame.VsBots), cutils.FmtDuration(time.Now().Sub(rss.FindGame.SearchingSince)))
		rs.matchInfoTxt.GetWidget().Visibility = widget.Visibility_Show
		rs.playBtnW.SetState(widget.WidgetChecked)
	} else {
		rs.playBtnW.GetWidget().Disabled = false
		rs.playBtnW.Text().Label = "PLAY"
		rs.matchInfoTxt.GetWidget().Visibility = widget.Visibility_Hide
		if !rss.SetupGame {
			rs.playBtnW.SetState(widget.WidgetUnchecked)
		}
	}

	if rss.WaitingRoom != nil {
		rs.loadModal(rs.waitingRoomW)
		for i, p := range rss.WaitingRoom.Players {
			rs.waitingRoomPlayersUsernames[i].Label = p.Username
			rs.waitingRoomPlayersUsernames[i].Color = cutils.TextColor
			if p.Accepted {
				rs.waitingRoomPlayersUsernames[i].Color = cutils.GreenTextColor
			}
			rs.waitingRoomPlayers[i].GetWidget().Visibility = widget.Visibility_Show
		}
		rs.waitingRoomSubTitle.Label = fmtVsRankedBots(rss.WaitingRoom.Size, rss.WaitingRoom.Ranked, false)
	} else {
		rs.closeModal(rs.waitingRoomW)
	}

	// This means the player joined/created a Lobby but it's not currently
	// on the lobby page
	rs.playBtnW.GetWidget().Disabled = false
	rs.backToLobbyBtnW.GetWidget().Visibility = widget.Visibility_Hide
	if cl != nil {
		rs.playBtnW.GetWidget().Disabled = true
		if rss.Route != utils.ShowLobbyRoute {
			rs.backToLobbyBtnW.GetWidget().Visibility = widget.Visibility_Show
		}
	}

	rs.ui.Draw(screen)
}

func (rs *RootStore) loadModal(w *widget.Window) {
	if rs.ui.IsWindowOpen(w) {
		return
	}
	//Get the preferred size of the content
	x, y := w.Contents.PreferredSize()
	//Create a rect with the preferred size of the content
	r := stdimage.Rect(0, 0, x, y)

	uirect := rs.ui.Container.GetWidget().Rect
	uix, uiy := uirect.Dx(), uirect.Dy()
	//Use the Add method to move the window to the specified point
	r = r.Add(stdimage.Point{uix/2 - x/2, uiy/2 - y/2})
	//Set the windows location to the rect.
	w.SetLocation(r)
	//Add the window to the UI.
	//Note: If the window is already added, this will just move the window and not add a duplicate.
	rs.ui.AddWindow(w)

	rs.modalBackgroundC.GetWidget().Visibility = widget.Visibility_Show
}

func (rs *RootStore) closeModal(w *widget.Window) {
	if !rs.ui.IsWindowOpen(w) {
		return
	}
	w.Close()
}

func fmtVsRankedBots(vs int, ranked, vsBots bool) string {
	var rk = "Ranked"
	if !ranked {
		rk = "Unranked"
	}
	if vsBots {
		rk += " (Bots)"
	}
	return fmt.Sprintf("Vs %d / %s", vs, rk)
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
	case action.SetupGame:
		rstate.SetupGame = act.SetupGame.Display
	case action.FindGame:
		var s = 1
		if !act.FindGame.Vs1 {
			s = 6
		}
		rstate.FindGame = &findGame{
			Size:           s,
			Ranked:         act.FindGame.Ranked,
			VsBots:         act.FindGame.VsBots,
			SearchingSince: time.Now(),
		}
	case action.ExitSearchingGame:
		rstate.FindGame = nil
	case action.SyncWaitingRoom:
		rstate.FindGame = nil
		rstate.WaitingRoom = &waitingRoom{
			Size:         act.SyncWaitingRoom.Size,
			Ranked:       act.SyncWaitingRoom.Ranked,
			Players:      act.SyncWaitingRoom.Players,
			WaitingSince: act.SyncWaitingRoom.WaitingSince,
		}
	case action.CancelWaitingGame, action.StartGame:
		rstate.WaitingRoom = nil
		rstate.SetupGame = false
	case action.NavigateTo:
		switch act.NavigateTo.Route {
		case utils.HomeRoute, utils.LobbiesRoute, utils.LearnRoute, utils.NewLobbyRoute, utils.ShowLobbyRoute:
			rstate.Route = act.NavigateTo.Route
		}
	}

	return rstate
}

func (rs *RootStore) buildUI() {
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(
			widget.NewStackedLayout(),
		),
		widget.ContainerOpts.BackgroundImage(cutils.ImageToNineSlice(cutils.BGKey)),
	)

	rs.ui = &ebitenui.UI{
		Container: rootContainer,
	}

	rootContainer.AddChild(rs.playUI())
	rootContainer.AddChild(rs.topbarUI())
	rootContainer.AddChild(rs.profileUI())

	rootContainer.AddChild(rs.homeUI())
	rootContainer.AddChild(rs.showLobbyUI())
	rootContainer.AddChild(rs.lobbiesUI())
	rootContainer.AddChild(rs.learnUI())

	rs.setupGameModal()
	rs.waitingRoomModal()
	rs.newLobbyModal()

	rootContainer.AddChild(rs.modalBackgroundUI())
}

func (rs *RootStore) modalBackgroundUI() *widget.Container {
	modalBackgroundC := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(cutils.ImageToNineSlice(cutils.SetupGameBGKey)),
	)
	modalBackgroundC.GetWidget().Visibility = widget.Visibility_Hide

	rs.modalBackgroundC = modalBackgroundC
	return modalBackgroundC
}

func (rs *RootStore) playUI() *widget.Container {
	playBtnC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(widgetPadding),
		)),
	)

	buttonsC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(5),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
			}),
		),
	)

	matchInfoTxt := widget.NewText(
		widget.TextOpts.Text("", cutils.NormalFont, cutils.TextColor),
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
				Position: widget.RowLayoutPositionStart,
				Stretch:  true,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.BigButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("PLAY", cutils.Font80, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(playBtnInsets),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			rss := rs.GetState().(RootState)
			if rss.FindGame == nil {
				// if it's not searching for a game then we start
				// the Setup
				actionDispatcher.SetupGame(true)
			} else {
				actionDispatcher.ExitSearchingGame(
					rs.Store.Users.Username(),
				)
				// If it's already waiting then we need to cancel
			}
		}),
		widget.ButtonOpts.ToggleMode(),

		cutils.BigButtonOptsPressedText,
		cutils.BigButtonOptsReleasedText,
		cutils.BigButtonOptsCursorEnteredText,
		cutils.BigButtonOptsCursorExitText,
		cutils.BigButtonOptsStatedChangeText,
	)

	backToLobbyBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				Stretch:  true,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.BigButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("BACK TO LOBBY", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   5,
			Right:  5,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			actionDispatcher.NavigateTo(utils.ShowLobbyRoute)
		}),

		cutils.BigButtonOptsPressedText,
		cutils.BigButtonOptsReleasedText,
		cutils.BigButtonOptsCursorEnteredText,
		cutils.BigButtonOptsCursorExitText,
	)

	buttonsC.AddChild(backToLobbyBtnW)
	buttonsC.AddChild(matchInfoTxt)
	buttonsC.AddChild(playBtnW)

	rs.playBtnW = playBtnW
	rs.backToLobbyBtnW = backToLobbyBtnW
	rs.matchInfoTxt = matchInfoTxt

	backToLobbyBtnW.GetWidget().Visibility = widget.Visibility_Hide
	matchInfoTxt.GetWidget().Visibility = widget.Visibility_Hide

	playBtnC.AddChild(buttonsC)

	return playBtnC
}

func (rs *RootStore) topbarUI() *widget.Container {
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
	topbarRC.AddChild(rs.topbarMenuUI())

	topbarC.AddChild(topbarRC)

	return topbarC
}

func (rs *RootStore) topbarMenuUI() *widget.Container {
	topbarMenuC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(15),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)

	homeBtnW := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.TextButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("HOME", cutils.Font40, &cutils.ButtonTextColorNoPressed),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			actionDispatcher.NavigateTo(utils.HomeRoute)
		}),
	)
	lobbiesBtnW := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.TextButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("LOBBIES", cutils.Font40, &cutils.ButtonTextColorNoPressed),

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
	learnBtnW := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.TextButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("LEARN", cutils.Font40, &cutils.ButtonTextColorNoPressed),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			actionDispatcher.NavigateTo(utils.LearnRoute)
		}),
	)

	widget.NewRadioGroup(
		widget.RadioGroupOpts.Elements(homeBtnW, lobbiesBtnW, learnBtnW),
		widget.RadioGroupOpts.InitialElement(homeBtnW),
	)

	topbarMenuC.AddChild(homeBtnW)
	topbarMenuC.AddChild(lobbiesBtnW)
	topbarMenuC.AddChild(learnBtnW)

	return topbarMenuC
}

func (rs *RootStore) profileUI() *widget.Container {
	profileC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(widgetPadding),
		)),
	)

	profileRC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
			widget.WidgetOpts.MinSize(357, 0),
		),
	)

	imageBtnC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)

	imageBtnBtnW := widget.NewButton(
		widget.ButtonOpts.Image(cutils.ButtonImageFromKey(cutils.Border4Key, 4, 6)),
	)
	imageBtnGraphicW := widget.NewGraphic(widget.GraphicOpts.Image(cutils.Images.Get(unit.Units[unit.Ninja.String()].FacesetKey())))

	imageBtnC.AddChild(imageBtnBtnW)
	imageBtnC.AddChild(imageBtnGraphicW)

	usernameTextW := widget.NewText(
		widget.TextOpts.Text("", cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	profileRC.AddChild(imageBtnC)
	profileRC.AddChild(usernameTextW)
	profileC.AddChild(profileRC)

	rs.usernameTextW = usernameTextW

	return profileC
}

func (rs *RootStore) setupGameModal() {
	btnPadding := widget.Insets{
		Left:   30,
		Right:  30,
		Top:    15,
		Bottom: 15,
	}
	frameC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(80)),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.SetupGameFrameKey, 1, 1, !isPressed)),
	)

	titleW := widget.NewText(
		widget.TextOpts.Text("Setup Game", cutils.Font60, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	formC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(20),
		)),
	)

	optionsC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(2),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(0, 5),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{true, true}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
	)

	vsLeft := widget.NewButton(

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.BigLeftTabButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Vs 1", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(btnPadding),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			//actionDispatcher.SetupGame(true)
		}),

		cutils.BigButtonOptsPressedText,
		cutils.BigButtonOptsReleasedText,
		cutils.BigButtonOptsCursorEnteredText,
		cutils.BigButtonOptsCursorExitText,
		cutils.BigButtonOptsStatedChangeText,
	)
	vsRight := widget.NewButton(

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.BigRightTabButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Vs 5", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(btnPadding),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			//actionDispatcher.SetupGame(true)
		}),

		cutils.BigButtonOptsPressedText,
		cutils.BigButtonOptsReleasedText,
		cutils.BigButtonOptsCursorEnteredText,
		cutils.BigButtonOptsCursorExitText,
		cutils.BigButtonOptsStatedChangeText,
	)

	rankedLeft := widget.NewButton(

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.BigLeftTabButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Ranked", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(btnPadding),

		cutils.BigButtonOptsPressedText,
		cutils.BigButtonOptsReleasedText,
		cutils.BigButtonOptsCursorEnteredText,
		cutils.BigButtonOptsCursorExitText,
		cutils.BigButtonOptsStatedChangeText,
	)
	rankedRight := widget.NewButton(

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.BigRightTabButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Unranked", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(btnPadding),

		cutils.BigButtonOptsPressedText,
		cutils.BigButtonOptsReleasedText,
		cutils.BigButtonOptsCursorEnteredText,
		cutils.BigButtonOptsCursorExitText,
		cutils.BigButtonOptsStatedChangeText,
	)

	widget.NewRadioGroup(
		widget.RadioGroupOpts.Elements(rankedLeft, rankedRight),
		widget.RadioGroupOpts.InitialElement(rankedRight),
	)

	widget.NewRadioGroup(
		widget.RadioGroupOpts.Elements(vsLeft, vsRight),
		widget.RadioGroupOpts.InitialElement(vsLeft),
	)
	botsC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(10),
			widget.RowLayoutOpts.Padding(widget.Insets{
				Left: 4,
			}),
		)),
	)

	botsTxt := widget.NewText(
		widget.TextOpts.Text("Play with bots?", cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	botsBtnW := widget.NewButton(
		// specify the images to sue
		widget.ButtonOpts.Image(cutils.CheckboxButtonResource()),

		widget.ButtonOpts.ToggleMode(),

		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if args.Button.State() == widget.WidgetUnchecked {
				vsLeft.Click()
				vsRight.GetWidget().Disabled = true
			} else {
				vsRight.GetWidget().Disabled = false
			}
		}),
	)

	findGameBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.BigButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Find Game", cutils.Font40, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(btnPadding),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			actionDispatcher.FindGame(
				rs.Store.Users.Username(),
				vsLeft.State() == widget.WidgetChecked,
				rankedLeft.State() == widget.WidgetChecked,
				botsBtnW.State() == widget.WidgetChecked,
			)
			actionDispatcher.SetupGame(false)
			rs.setupGameW.Close()
		}),

		cutils.BigButtonOptsPressedText,
		cutils.BigButtonOptsReleasedText,
		cutils.BigButtonOptsCursorEnteredText,
		cutils.BigButtonOptsCursorExitText,
	)

	botsC.AddChild(botsTxt)
	botsC.AddChild(botsBtnW)

	botsBtnW.Click()

	optionsC.AddChild(vsLeft)
	optionsC.AddChild(vsRight)

	optionsC.AddChild(rankedLeft)
	optionsC.AddChild(rankedRight)

	formC.AddChild(optionsC)
	formC.AddChild(botsC)

	frameC.AddChild(titleW)
	frameC.AddChild(formC)
	frameC.AddChild(findGameBtnW)

	window := widget.NewWindow(
		widget.WindowOpts.Contents(frameC),
		widget.WindowOpts.Modal(),
		widget.WindowOpts.CloseMode(widget.CLICK_OUT),
		widget.WindowOpts.ClosedHandler(func(args *widget.WindowClosedEventArgs) {
			rs.modalBackgroundC.GetWidget().Visibility = widget.Visibility_Hide
			actionDispatcher.SetupGame(false)
		}),
	)

	rs.setupGameW = window
}

func (rs *RootStore) waitingRoomModal() {
	waitingRoomC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.BackgroundImage(cutils.ImageToNineSlice(cutils.SetupGameBGKey)),
	)

	frameC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(80)),
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
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.SetupGameFrameKey, 1, 1, !isPressed)),
	)

	titleW := widget.NewText(
		widget.TextOpts.Text("Game Found", cutils.Font60, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	subTitleW := widget.NewText(
		widget.TextOpts.Text(fmt.Sprintf("%s / %s", "Vs 1", "Unranked"), cutils.Font40, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	playersC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(50),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	for _ = range make([]int, 6) {
		playerC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			)),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
			),
		)

		imageBtnC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewStackedLayout()),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
			),
		)

		imageBtnBtnW := widget.NewButton(
			widget.ButtonOpts.Image(cutils.ButtonImageFromKey(cutils.Border4Key, 4, 6)),
		)
		imageBtnGraphicW := widget.NewGraphic(widget.GraphicOpts.Image(cutils.Images.Get(unit.Units[unit.Ninja.String()].FacesetKey())))

		imageBtnC.AddChild(imageBtnBtnW)
		imageBtnC.AddChild(imageBtnGraphicW)

		usernameTextW := widget.NewText(
			widget.TextOpts.Text("", cutils.NormalFont, cutils.TextColor),
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
			),
		)

		playerC.AddChild(imageBtnC)
		playerC.AddChild(usernameTextW)
		rs.waitingRoomPlayersUsernames = append(rs.waitingRoomPlayersUsernames, usernameTextW)
		rs.waitingRoomPlayers = append(rs.waitingRoomPlayers, playerC)

		playersC.AddChild(playerC)
		playerC.GetWidget().Visibility = widget.Visibility_Hide
	}

	optionsC := widget.NewContainer(
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
	acceptBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.BigButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Accept", cutils.Font40, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			actionDispatcher.AcceptWaitingGame(rs.Store.Users.Username())
		}),

		cutils.BigButtonOptsPressedText,
		cutils.BigButtonOptsReleasedText,
		cutils.BigButtonOptsCursorEnteredText,
		cutils.BigButtonOptsCursorExitText,
	)
	cancelBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.BigCancelButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Cancel", cutils.Font40, &cutils.ButtonTextCancelColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			actionDispatcher.CancelWaitingGame(rs.Store.Users.Username())
			rs.waitingRoomW.Close()
		}),
	)

	optionsC.AddChild(cancelBtnW)
	optionsC.AddChild(acceptBtnW)

	frameC.AddChild(titleW)
	frameC.AddChild(subTitleW)
	frameC.AddChild(playersC)
	frameC.AddChild(optionsC)

	waitingRoomC.AddChild(frameC)

	window := widget.NewWindow(
		widget.WindowOpts.Contents(waitingRoomC),
		widget.WindowOpts.Modal(),
		widget.WindowOpts.ClosedHandler(func(args *widget.WindowClosedEventArgs) {
			rs.modalBackgroundC.GetWidget().Visibility = widget.Visibility_Hide
		}),
	)

	rs.waitingRoomSubTitle = subTitleW
	rs.waitingRoomW = window
}

func (rs *RootStore) homeUI() *widget.Container {
	homeC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	contentC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.Insets{
				Left:   261,
				Right:  200,
				Top:    200,
				Bottom: 200,
			}),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				StretchHorizontal:  true,
				StretchVertical:    true,
			}),
		),
	)

	titleTextW := widget.NewText(
		widget.TextOpts.Text("HOME", cutils.Font60, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)

	contentC.AddChild(titleTextW)

	homeC.AddChild(contentC)

	rs.homeC = homeC

	return homeC
}

func (rs *RootStore) lobbiesUI() *widget.Container {
	lobbiesC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(1),
			//Define how much padding to inset the child content
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(30)),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(20, 10),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}),

			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:   261,
				Right:  200,
				Top:    200,
				Bottom: 200,
			}),
		)),
	)
	buttonsWrapperC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	buttonsC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)

	refreshBtnW := widget.NewButton(
		// specify the images to sue
		widget.ButtonOpts.Image(cutils.ButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("REFRESH", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			actionDispatcher.RefreshLobbies()
		}),

		cutils.ButtonOptsPressedText,
		cutils.ButtonOptsReleasedText,
		cutils.ButtonOptsCursorEnteredText,
		cutils.ButtonOptsCursorExitText,
	)
	newBtnW := widget.NewButton(
		// specify the images to sue
		widget.ButtonOpts.Image(cutils.ButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("CREATE LOBBY", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			actionDispatcher.NavigateTo(utils.NewLobbyRoute)
		}),

		cutils.ButtonOptsPressedText,
		cutils.ButtonOptsReleasedText,
		cutils.ButtonOptsCursorEnteredText,
		cutils.ButtonOptsCursorExitText,
	)
	buttonsC.AddChild(refreshBtnW)
	buttonsC.AddChild(newBtnW)

	tableWrapperC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(1),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(20, 10),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}),
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(4)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionStart,
			}),
		),
		widget.ContainerOpts.BackgroundImage(cutils.ImageToNineSlice(cutils.LobbiesTableKey)),
	)

	tableHeaderC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(3),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(20, 10),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true, true, false}, []bool{false}),
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left: 14,
				Top:  5,
			}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionStart,
			}),
		),
	)

	nameHeaderText := widget.NewText(
		widget.TextOpts.Text("NAME", cutils.Font40, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionStart,
			}),
		),
	)

	playersHeaderText := widget.NewText(
		widget.TextOpts.Text("PLAYERS", cutils.Font40, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionStart,
			}),
		),
	)

	tableHeaderC.AddChild(nameHeaderText)
	tableHeaderC.AddChild(playersHeaderText)

	lobbiesTableC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(2),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionStart,
			}),
		),
	)

	lobbiesTableW := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(1),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(20, 10),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false}),

			widget.GridLayoutOpts.Padding(widget.Insets{
				//Left: 14,
				Top: 5,
			}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionStart,
			}),
		),
	)

	//Create the new ScrollContainer object
	scrollContainer := widget.NewScrollContainer(
		//Set the content that will be scrolled
		widget.ScrollContainerOpts.Content(lobbiesTableW),
		//Tell the container to stretch the content width to match available space
		widget.ScrollContainerOpts.StretchContentWidth(),
		//Set the background images for the scrollable container
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle: image.NewNineSliceColor(cutils.Transparent),
			Mask: image.NewNineSliceColor(cutils.Black),
		}),
	)

	pageSizeFunc := func() int {
		ps := int(math.Round(float64(scrollContainer.ViewRect().Dy()) / float64(lobbiesTableW.GetWidget().Rect.Dy()) * 1000))
		if ps < 1000 {
			return 1000
		}
		return ps
	}

	//Create a vertical Slider bar to control the ScrollableContainer
	vSlider := widget.NewSlider(
		widget.SliderOpts.Direction(widget.DirectionVertical),
		widget.SliderOpts.MinMax(0, 1000),
		widget.SliderOpts.PageSizeFunc(pageSizeFunc),
		//On change update scroll location based on the Slider's value
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			scrollContainer.ScrollTop = float64(args.Slider.Current) / 1000
		}),
		widget.SliderOpts.Images(
			// Set the track images
			&widget.SliderTrackImage{
				Idle:  image.NewNineSliceColor(cutils.Transparent),
				Hover: image.NewNineSliceColor(cutils.Transparent),
			},
			// Set the handle images
			&widget.ButtonImage{
				Idle:    image.NewNineSliceColor(cutils.TableBlack),
				Pressed: image.NewNineSliceColor(cutils.TableBlack),
			},
		),
	)

	//Set the slider's position if the scrollContainer is scrolled by other means than the slider
	scrollContainer.GetWidget().ScrolledEvent.AddHandler(func(args interface{}) {
		a := args.(*widget.WidgetScrolledEventArgs)
		p := pageSizeFunc() / 3
		if p < 1 {
			p = 1
		}
		vSlider.Current -= int(math.Round(a.Y * float64(p)))
	})

	lobbiesTableC.AddChild(scrollContainer)
	lobbiesTableC.AddChild(vSlider)

	tableWrapperC.AddChild(tableHeaderC)
	tableWrapperC.AddChild(lobbiesTableC)
	buttonsWrapperC.AddChild(buttonsC)
	lobbiesC.AddChild(buttonsWrapperC)
	lobbiesC.AddChild(tableWrapperC)

	rs.lobbiesC = lobbiesC
	rs.lobbiesTableW = lobbiesTableW
	rs.lobbiesSC = scrollContainer

	return lobbiesC
}

func (rs *RootStore) showLobbyUI() *widget.Container {
	lobbyC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(1),
			//Define how much padding to inset the child content
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(30)),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(20, 10),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true, false}),

			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:   261,
				Right:  200,
				Top:    200,
				Bottom: 200,
			}),
		)),
	)

	titleW := widget.NewText(
		widget.TextOpts.Text("LOBBY", cutils.Font60, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				//HorizontalPosition: widget.GridLayoutPositionStart,
				//VerticalPosition:   widget.GridLayoutPositionStart,
			}),
		),
	)

	botsC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)

	botsTextLW := widget.NewText(
		widget.TextOpts.Text("Bots: ", cutils.NormalFont, color.White),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	addBotBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),

		// specify the images to sle
		widget.ButtonOpts.Image(cutils.ButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("+", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			cl := rs.Store.Lobbies.FindCurrent()
			if len(cl.Players)+1 > cl.MaxPlayers {
				return
			}
			actionDispatcher.JoinLobby(cl.ID, fmt.Sprintf("Bot-%d", len(cl.Players)+1), isBot)
		}),

		cutils.ButtonOptsPressedText,
		cutils.ButtonOptsReleasedText,
		cutils.ButtonOptsCursorEnteredText,
		cutils.ButtonOptsCursorExitText,
	)

	removeBotBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				//Stretch:  true,
			}),
		),

		// specify the images to sle
		widget.ButtonOpts.Image(cutils.ButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("-", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			cl := rs.Store.Lobbies.FindCurrent()
			for p, ib := range cl.Players {
				if ib {
					actionDispatcher.LeaveLobby(cl.ID, p)
					break
				}
			}
		}),

		cutils.ButtonOptsPressedText,
		cutils.ButtonOptsReleasedText,
		cutils.ButtonOptsCursorEnteredText,
		cutils.ButtonOptsCursorExitText,
	)

	entries := make([]any, 0, 0)
	playersListWrapperC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(1),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(20, 10),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}),
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(4)),
		)),
		widget.ContainerOpts.WidgetOpts(
		//widget.WidgetOpts.LayoutData(widget.GridLayoutData{
		//HorizontalPosition: widget.GridLayoutPositionStart,
		//VerticalPosition:   widget.GridLayoutPositionStart,
		//}),
		),
		widget.ContainerOpts.BackgroundImage(cutils.ImageToNineSlice(cutils.LobbiesTableKey)),
	)
	playersHeaderWC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	playersHeaderW := widget.NewText(
		widget.TextOpts.Text("PLAYER", cutils.Font40, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)
	playersListW := widget.NewList(
		// Set how wide the list should be
		widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				Stretch:  true,
			}),
		)),
		// Set the entries in the list
		widget.ListOpts.Entries(entries),
		widget.ListOpts.ScrollContainerOpts(
			// Set the background images/color for the list
			widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
				Idle: image.NewNineSliceColor(cutils.Transparent),
				//Disabled: image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
				Mask: image.NewNineSliceColor(cutils.Black),
			}),
		),
		widget.ListOpts.SliderOpts(
			// Set the background images/color for the background of the slider track
			widget.SliderOpts.Images(
				&widget.SliderTrackImage{
					Idle:  image.NewNineSliceColor(cutils.Transparent),
					Hover: image.NewNineSliceColor(cutils.Transparent),
				},
				&widget.ButtonImage{
					Idle:    image.NewNineSliceColor(cutils.TableBlack),
					Pressed: image.NewNineSliceColor(cutils.TableBlack),
				},
			),
			//widget.SliderOpts.MinHandleSize(5),
			// Set how wide the track should be
			//widget.SliderOpts.TrackPadding(widget.NewInsetsSimple(2)),
		),
		// Hide the horizontal slider
		widget.ListOpts.HideHorizontalSlider(),
		// Set the font for the list options
		widget.ListOpts.EntryFontFace(cutils.NormalFont),
		// Set the colors for the list
		widget.ListOpts.EntryColor(&widget.ListEntryColor{
			Selected:   cutils.TextColor, // Foreground color for the unfocused selected entry
			Unselected: cutils.TextColor,
		}),
		// This required function returns the string displayed in the list
		widget.ListOpts.EntryLabelFunc(func(e interface{}) string {
			return e.(cutils.ListEntry).Text
		}),
		// Padding for each entry
		widget.ListOpts.EntryTextPadding(widget.NewInsetsSimple(5)),
		// Text position for each entry
		widget.ListOpts.EntryTextPosition(widget.TextPositionStart, widget.TextPositionCenter),
	)

	actionBtnsWC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	actionBtnC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)
	leaveBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				Stretch:  true,
			}),
		),

		// specify the images to sle
		widget.ButtonOpts.Image(cutils.ButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("LEAVE", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			cl := rs.Store.Lobbies.FindCurrent()
			un := rs.Store.Users.Username()
			actionDispatcher.LeaveLobby(cl.ID, un)
			actionDispatcher.NavigateTo(utils.LobbiesRoute)
		}),

		cutils.ButtonOptsPressedText,
		cutils.ButtonOptsReleasedText,
		cutils.ButtonOptsCursorEnteredText,
		cutils.ButtonOptsCursorExitText,
	)
	deleteBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				Stretch:  true,
			}),
		),

		// specify the images to sle
		widget.ButtonOpts.Image(cutils.ButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("DELETE", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			cl := rs.Store.Lobbies.FindCurrent()
			actionDispatcher.DeleteLobby(cl.ID)
			actionDispatcher.NavigateTo(utils.LobbiesRoute)
		}),

		cutils.ButtonOptsPressedText,
		cutils.ButtonOptsReleasedText,
		cutils.ButtonOptsCursorEnteredText,
		cutils.ButtonOptsCursorExitText,
	)
	startBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				Stretch:  true,
			}),
		),

		// specify the images to sle
		widget.ButtonOpts.Image(cutils.ButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("START", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			cl := rs.Store.Lobbies.FindCurrent()
			actionDispatcher.StartLobby(cl.ID)
		}),

		cutils.ButtonOptsPressedText,
		cutils.ButtonOptsReleasedText,
		cutils.ButtonOptsCursorEnteredText,
		cutils.ButtonOptsCursorExitText,
	)

	actionBtnC.AddChild(leaveBtnW)
	actionBtnC.AddChild(deleteBtnW)
	actionBtnC.AddChild(startBtnW)

	actionBtnsWC.AddChild(actionBtnC)

	botsC.AddChild(botsTextLW)
	botsC.AddChild(removeBotBtnW)
	botsC.AddChild(addBotBtnW)

	playersHeaderWC.AddChild(playersHeaderW)
	playersHeaderWC.AddChild(botsC)

	playersListWrapperC.AddChild(playersHeaderWC)
	playersListWrapperC.AddChild(playersListW)

	lobbyC.AddChild(titleW)
	lobbyC.AddChild(playersListWrapperC)
	lobbyC.AddChild(actionBtnsWC)

	rs.showLobbyC = lobbyC
	rs.showLobbyNameW = titleW
	rs.showLobbyPlayersHeaderW = playersHeaderW
	rs.showLobbyPlayersL = playersListW
	rs.showLobbyLeaveBtnW = leaveBtnW
	rs.showLobbyDeleteBtnW = deleteBtnW
	rs.showLobbyStartBtnW = startBtnW
	rs.showLobbyBotsC = botsC

	return lobbyC
	//return rs.showLobbiesNotlink()
}

func (rs *RootStore) showLobbiesNotlink() *widget.Container {
	lobbyC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(1),
			//Define how much padding to inset the child content
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(30)),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(20, 10),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true, false}),

			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:   261,
				Right:  200,
				Top:    200,
				Bottom: 200,
			}),
		)),
	)

	titleW := widget.NewText(
		widget.TextOpts.Text("LOBBY", cutils.Font60, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				//HorizontalPosition: widget.GridLayoutPositionStart,
				//VerticalPosition:   widget.GridLayoutPositionStart,
			}),
		),
	)

	botsC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)

	botsTextLW := widget.NewText(
		widget.TextOpts.Text("Bots: ", cutils.NormalFont, color.White),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	addBotBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),

		// specify the images to sle
		widget.ButtonOpts.Image(cutils.ButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("+", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			cl := rs.Store.Lobbies.FindCurrent()
			if len(cl.Players)+1 > cl.MaxPlayers {
				return
			}
			actionDispatcher.JoinLobby(cl.ID, fmt.Sprintf("Bot-%d", len(cl.Players)+1), isBot)
		}),

		cutils.ButtonOptsPressedText,
		cutils.ButtonOptsReleasedText,
		cutils.ButtonOptsCursorEnteredText,
		cutils.ButtonOptsCursorExitText,
	)

	removeBotBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				//Stretch:  true,
			}),
		),

		// specify the images to sle
		widget.ButtonOpts.Image(cutils.ButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("-", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			cl := rs.Store.Lobbies.FindCurrent()
			for p, ib := range cl.Players {
				if ib {
					actionDispatcher.LeaveLobby(cl.ID, p)
					break
				}
			}
		}),

		cutils.ButtonOptsPressedText,
		cutils.ButtonOptsReleasedText,
		cutils.ButtonOptsCursorEnteredText,
		cutils.ButtonOptsCursorExitText,
	)

	entries := make([]any, 0, 0)
	playersListWrapperC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(1),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(20, 10),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}),
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(4)),
		)),
		widget.ContainerOpts.WidgetOpts(
		//widget.WidgetOpts.LayoutData(widget.GridLayoutData{
		//HorizontalPosition: widget.GridLayoutPositionStart,
		//VerticalPosition:   widget.GridLayoutPositionStart,
		//}),
		),
		widget.ContainerOpts.BackgroundImage(cutils.ImageToNineSlice(cutils.LobbiesTableKey)),
	)
	playersHeaderWC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	playersHeaderW := widget.NewText(
		widget.TextOpts.Text("PLAYER", cutils.Font40, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)
	playersListW := widget.NewList(
		// Set how wide the list should be
		widget.ListOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				Stretch:  true,
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
		widget.ListOpts.EntryFontFace(cutils.NormalFont),
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
	)

	actionBtnsWC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	actionBtnC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)
	leaveBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				Stretch:  true,
			}),
		),

		// specify the images to sle
		widget.ButtonOpts.Image(cutils.ButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("LEAVE", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			cl := rs.Store.Lobbies.FindCurrent()
			un := rs.Store.Users.Username()
			actionDispatcher.LeaveLobby(cl.ID, un)
			actionDispatcher.NavigateTo(utils.LobbiesRoute)
		}),

		cutils.ButtonOptsPressedText,
		cutils.ButtonOptsReleasedText,
		cutils.ButtonOptsCursorEnteredText,
		cutils.ButtonOptsCursorExitText,
	)
	deleteBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				Stretch:  true,
			}),
		),

		// specify the images to sle
		widget.ButtonOpts.Image(cutils.ButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("DELETE", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			cl := rs.Store.Lobbies.FindCurrent()
			actionDispatcher.DeleteLobby(cl.ID)
			actionDispatcher.NavigateTo(utils.LobbiesRoute)
		}),

		cutils.ButtonOptsPressedText,
		cutils.ButtonOptsReleasedText,
		cutils.ButtonOptsCursorEnteredText,
		cutils.ButtonOptsCursorExitText,
	)
	startBtnW := widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				Stretch:  true,
			}),
		),

		// specify the images to sle
		widget.ButtonOpts.Image(cutils.ButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("START", cutils.NormalFont, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   30,
			Right:  30,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			cl := rs.Store.Lobbies.FindCurrent()
			actionDispatcher.StartLobby(cl.ID)
		}),

		cutils.ButtonOptsPressedText,
		cutils.ButtonOptsReleasedText,
		cutils.ButtonOptsCursorEnteredText,
		cutils.ButtonOptsCursorExitText,
	)

	actionBtnC.AddChild(leaveBtnW)
	actionBtnC.AddChild(deleteBtnW)
	actionBtnC.AddChild(startBtnW)

	actionBtnsWC.AddChild(actionBtnC)

	botsC.AddChild(botsTextLW)
	botsC.AddChild(removeBotBtnW)
	botsC.AddChild(addBotBtnW)

	playersHeaderWC.AddChild(playersHeaderW)
	playersHeaderWC.AddChild(botsC)

	playersListWrapperC.AddChild(playersHeaderWC)
	playersListWrapperC.AddChild(playersListW)

	lobbyC.AddChild(titleW)
	lobbyC.AddChild(playersListWrapperC)
	lobbyC.AddChild(actionBtnsWC)
	//lobbyC.AddChild(startBtnW)
	//lobbyC.AddChild(deleteBtnW)

	return lobbyC
}

func (rs *RootStore) addLobbiesToTable(ls []*store.Lobby) {
	rs.lobbiesTableW.RemoveChildren()
	for _, l := range ls {
		nameText := widget.NewText(
			widget.TextOpts.Text(l.Name, cutils.NormalFont, cutils.TextColor),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionStart,
					VerticalPosition:   widget.GridLayoutPositionStart,
				}),
			),
		)

		playersText := widget.NewText(
			widget.TextOpts.Text(fmt.Sprintf("%d/%d", len(l.Players), l.MaxPlayers), cutils.NormalFont, cutils.TextColor),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionStart,
					VerticalPosition:   widget.GridLayoutPositionStart,
				}),
			),
		)
		var lobbiesTableRowC *widget.Container
		lobbiesTableRowC = widget.NewContainer(
			// the container will use an anchor layout to layout its single child widget
			widget.ContainerOpts.Layout(widget.NewGridLayout(
				//Define number of columns in the grid
				widget.GridLayoutOpts.Columns(2),
				widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{false}),
				widget.GridLayoutOpts.Padding(widget.Insets{
					Left: 14,
					Top:  5,
				}),
			)),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: widget.GridLayoutPositionStart,
					VerticalPosition:   widget.GridLayoutPositionStart,
				}),
				widget.WidgetOpts.MouseButtonPressedHandler(func(args *widget.WidgetMouseButtonPressedEventArgs) {
					cl := rs.Store.Lobbies.FindCurrent()
					if cl != nil {
						// TODO: Return an error saying that "You already belong to a Lobby"
						return
					}
					un := rs.Store.Users.Username()
					actionDispatcher.JoinLobby(l.ID, un, !isBot)
					actionDispatcher.SelectLobby(l.ID)
					actionDispatcher.NavigateTo(utils.ShowLobbyRoute)
				}),
				widget.WidgetOpts.CursorMoveHandler(func(args *widget.WidgetCursorMoveEventArgs) {
					// This is the HOVER button main color just with a half Alpha
					lobbiesTableRowC.BackgroundImage = image.NewNineSliceColor(color.NRGBA{R: 254, G: 173, B: 84, A: 126})
					nameText.Color = cutils.ButtonTextHoverColor
					playersText.Color = cutils.ButtonTextHoverColor
				}),
				widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
					lobbiesTableRowC.BackgroundImage = nil
					nameText.Color = cutils.ButtonTextIdleColor
					playersText.Color = cutils.ButtonTextIdleColor
				}),
			),
		)
		lobbiesTableRowC.AddChild(nameText)
		lobbiesTableRowC.AddChild(playersText)
		rs.lobbiesTableW.AddChild(lobbiesTableRowC)
	}
}

func (rs *RootStore) newLobbyModal() {
	frameC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(80)),
			widget.RowLayoutOpts.Spacing(50),
		)),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.SetupGameFrameKey, 1, 1, !isPressed)),
	)

	titleW := widget.NewText(
		widget.TextOpts.Text("New Lobby", cutils.Font60, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	formC := widget.NewContainer(
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(1),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(0, 20),
			widget.GridLayoutOpts.Padding(widget.Insets{
				Bottom: 20,
			}),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, false}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				Stretch:  true,
			}),
		),
	)

	nameRC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(5),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionStart,
			}),
		),
	)

	playersRC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(5),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionStart,
			}),
		),
	)

	nameLabelW := widget.NewText(
		widget.TextOpts.Text("Name", cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				Stretch:  true,
			}),
		),
	)

	var createLobbyBtnW *widget.Button

	nameInputW := widget.NewTextInput(
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
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
		widget.TextInputOpts.Placeholder("Lobby name"),

		//This is called when the user hits the "Enter" key.
		//There are other options that can configure this behavior
		widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
			createLobbyBtnW.Click()
		}),
	)

	playersLabelW := widget.NewText(
		widget.TextOpts.Text("Players", cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				Stretch:  true,
			}),
		),
	)

	playersSliderW := widget.NewSlider(
		// Set the slider orientation - n/s vs e/w
		widget.SliderOpts.Direction(widget.DirectionHorizontal),
		// Set the minimum and maximum value for the slider
		widget.SliderOpts.MinMax(2, 6),

		widget.SliderOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
				Stretch:  true,
			}),
			widget.WidgetOpts.MinSize(0, 30),
		),
		widget.SliderOpts.Images(
			// Set the track images
			&widget.SliderTrackImage{
				Idle: cutils.ImageToNineSlice(cutils.GrayInputBGKey),
			},
			// Set the handle images
			&widget.ButtonImage{
				Idle:     cutils.LoadImageNineSlice(cutils.HSliderGrabberKey, 9, 7, !isPressed),
				Hover:    cutils.LoadImageNineSlice(cutils.HSliderGrabberHoverKey, 9, 7, !isPressed),
				Pressed:  cutils.LoadImageNineSlice(cutils.HSliderGrabberHoverKey, 9, 7, !isPressed),
				Disabled: cutils.LoadImageNineSlice(cutils.HSliderGrabberDisabledKey, 9, 7, !isPressed),
			},
		),
		// Set the size of the handle
		//widget.SliderOpts.FixedHandleSize(6),
		//widget.SliderOpts.MinHandleSize(5),
		// Set the offset to display the track
		widget.SliderOpts.TrackOffset(0),
		// Set the size to move the handle
		widget.SliderOpts.PageSizeFunc(func() int {
			return 1
		}),
		// Set the callback to call when the slider value is changed
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			playersLabelW.Label = fmt.Sprintf("Players: %d", args.Current)
		}),
	)

	createLobbyBtnW = widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.BigButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("CREATE", cutils.Font40, &cutils.ButtonTextColor),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   50,
			Right:  50,
			Top:    10,
			Bottom: 10,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			lid := uuid.Must(uuid.NewV4()).String()

			actionDispatcher.CreateLobby(lid, rs.Store.Users.Username(), nameInputW.GetText(), playersSliderW.Current)
			actionDispatcher.SelectLobby(lid)
			rs.newLobbyW.Close()
		}),

		cutils.BigButtonOptsPressedText,
		cutils.BigButtonOptsReleasedText,
		cutils.BigButtonOptsCursorEnteredText,
		cutils.BigButtonOptsCursorExitText,
	)

	nameInputW.Focus(true)

	nameRC.AddChild(nameLabelW)
	nameRC.AddChild(nameInputW)

	playersRC.AddChild(playersLabelW)
	playersRC.AddChild(playersSliderW)

	formC.AddChild(nameRC)
	formC.AddChild(playersRC)

	frameC.AddChild(titleW)
	frameC.AddChild(formC)
	frameC.AddChild(createLobbyBtnW)

	window := widget.NewWindow(
		widget.WindowOpts.Contents(frameC),
		widget.WindowOpts.Modal(),
		widget.WindowOpts.CloseMode(widget.CLICK_OUT),
		widget.WindowOpts.ClosedHandler(func(args *widget.WindowClosedEventArgs) {
			cl := rs.Store.Lobbies.FindCurrent()
			r := utils.LobbiesRoute
			if cl != nil {
				r = utils.ShowLobbyRoute
			}
			actionDispatcher.NavigateTo(r)
			rs.modalBackgroundC.GetWidget().Visibility = widget.Visibility_Hide
		}),
	)

	rs.newLobbyW = window
}

func (rs *RootStore) learnUI() *widget.Container {
	learnC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	contentC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.Insets{
				Left:   261,
				Right:  200,
				Top:    200,
				Bottom: 200,
			}),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				StretchHorizontal:  true,
				StretchVertical:    true,
			}),
		),
	)

	titleTextW := widget.NewText(
		widget.TextOpts.Text("LEARN", cutils.Font60, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)

	contentC.AddChild(titleTextW)

	learnC.AddChild(contentC)

	rs.learnC = learnC

	return learnC
}
