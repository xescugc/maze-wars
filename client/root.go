package client

import (
	"bytes"
	"fmt"
	stdimage "image"
	"image/color"
	"image/gif"
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
	"github.com/xescugc/maze-wars/assets"
	cutils "github.com/xescugc/maze-wars/client/utils"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/unit"
	"github.com/xescugc/maze-wars/unit/ability"
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
	isBot     = true

	unitToolTipTitleTmpl = "%s (%s)"
)

type RootStore struct {
	*flux.ReduceStore

	Store  *Store
	Logger *slog.Logger

	ui *ebitenui.UI

	usernameTextW   *widget.Text
	userImageGW     *widget.Graphic
	playBtnW        *widget.Button
	backToLobbyBtnW *widget.Button
	matchInfoTxt    *widget.Text

	setupGameW                  *widget.Window
	waitingRoomW                *widget.Window
	waitingRoomC                *widget.Container
	waitingRoomPlayers          []*widget.Container
	waitingRoomPlayersUsernames []*widget.Text
	waitingRoomPlayersImages    []*widget.Graphic
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

var (
	unitTypeGifB = map[unit.Type][]byte{
		unit.Ninja:         assets.MWUnitNinja_GIF,
		unit.Statue:        assets.MWUnitStatue_GIF,
		unit.Hunter:        assets.MWUnitHunter_GIF,
		unit.Slime:         assets.MWUnitSlime_GIF,
		unit.Mole:          assets.MWUnitMole_GIF,
		unit.SkeletonDemon: assets.MWUnitSkeletonDemon_GIF,
		unit.Butterfly:     assets.MWUnitButterfly_GIF,
		unit.BlendMaster:   assets.MWUnitBlendMaster_GIF,
		unit.Robot:         assets.MWUnitRobot_GIF,
		unit.MonkeyBoxer:   assets.MWUnitMonkeyBoxer_GIF,
	}

	unitTypeGif = make(map[unit.Type]*gif.GIF)
)

func init() {
	if !cutils.IsWASM() {
		for t, b := range unitTypeGifB {
			g, err := gif.DecodeAll(bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			unitTypeGif[t] = g
		}
	}

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
	rs.userImageGW.Image = cutils.Images.Get(unit.Units[rs.Store.Users.ImageKey()].FacesetKey())
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
			rs.waitingRoomPlayersImages[i].Image = cutils.Images.Get(unit.Units[p.ImageKey].FacesetKey())
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

	rootContainer.AddChild(
		rs.playUI(),
		rs.topbarUI(),
		rs.profileUI(),
		rs.homeUI(),
		rs.showLobbyUI(),
		rs.lobbiesUI(),
		rs.learnUI(),
		rs.modalBackgroundUI(),
	)

	rs.setupGameModal()
	rs.waitingRoomModal()
	rs.newLobbyModal()

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

	topbarRC.AddChild(
		logoBtnW,
		rs.topbarMenuUI(),
	)

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

	topbarMenuC.AddChild(
		homeBtnW,
		lobbiesBtnW,
		learnBtnW,
	)

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
				Position:  widget.RowLayoutPositionStart,
				MaxWidth:  48,
				MaxHeight: 48,
			}),
		),
	)

	imageBtnBtnW := widget.NewButton(
		widget.ButtonOpts.Image(cutils.ButtonImageFromKey(cutils.Border4Key, 4, 4)),
	)
	imageBtnGraphicW := widget.NewGraphic(widget.GraphicOpts.Image(cutils.Images.Get(unit.Units[unit.Ninja.String()].FacesetKey())))

	imageBtnC.AddChild(
		imageBtnBtnW,
		imageBtnGraphicW,
	)

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
	rs.userImageGW = imageBtnGraphicW

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
	rankedLeft.GetWidget().Disabled = true

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

		playerC.AddChild(
			imageBtnC,
			usernameTextW,
		)
		rs.waitingRoomPlayersUsernames = append(rs.waitingRoomPlayersUsernames, usernameTextW)
		rs.waitingRoomPlayersImages = append(rs.waitingRoomPlayersImages, imageBtnGraphicW)
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
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:   261,
				Right:  200,
				Top:    200,
				Bottom: 200,
			}),
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Spacing(0, 20),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}),
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

	warningC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.ToolTipBGKey, 8, 8, !isPressed)),
		widget.ContainerOpts.WidgetOpts(
			func(w *widget.Widget) {
				w.Visibility = widget.Visibility_Hide_Blocking
			},
		),
	)
	warnigTxtW := widget.NewText(
		widget.TextOpts.Text("The server will be in maintenance for 10'", cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				StretchVertical:    true,
			}),
		),
	)

	infoC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Spacing(50, 0),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{true}),
		)),
	)

	newsTAW := widget.NewTextArea(
		//Set gap between scrollbar and text
		widget.TextAreaOpts.ControlWidgetSpacing(2),
		//Tell the textarea to display bbcodes
		widget.TextAreaOpts.ProcessBBCode(true),
		//Set the font color
		widget.TextAreaOpts.FontColor(cutils.TextColor),
		//Set the font face (size) to use
		widget.TextAreaOpts.FontFace(cutils.NormalFont),
		//Set the initial text for the textarea
		//It will automatically line wrap and process newlines characters
		//If ProcessBBCode is true it will parse out bbcode
		widget.TextAreaOpts.Text(`Thank you very much for playing the Maze Wars Alpha!

This is the first game I created and is yet not finished (I have a lot of ideas to add too it) so any recommendations/bugs are more than welcome

Maze Wars will always be a free and OSS.

GitHub: github.com/xescugc/maze-wars
Discord: discord.gg/t2BBFGwj5U
`),
		//Tell the TextArea to show the vertical scrollbar
		//widget.TextAreaOpts.ShowVerticalScrollbar(),
		//Set padding between edge of the widget and where the text is drawn
		widget.TextAreaOpts.TextPadding(widget.NewInsetsSimple(10)),
		//This sets the background images for the scroll container
		widget.TextAreaOpts.ScrollContainerOpts(
			widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
				Idle: cutils.LoadImageNineSlice(cutils.HomeWidgetsBGKey, 1, 1, !isPressed),
				Mask: image.NewNineSliceColor(cutils.Black),
			}),
		),
		//This sets the images to use for the sliders
		widget.TextAreaOpts.SliderOpts(
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
		),
	)

	releaseNotesTAW := widget.NewTextArea(
		widget.TextAreaOpts.ContainerOpts(
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
					Stretch:  true,
				}),
			),
		),
		//Set gap between scrollbar and text
		widget.TextAreaOpts.ControlWidgetSpacing(2),
		//Tell the textarea to display bbcodes
		widget.TextAreaOpts.ProcessBBCode(true),
		//Set the font color
		widget.TextAreaOpts.FontColor(cutils.TextColor),
		//Set the font face (size) to use
		widget.TextAreaOpts.FontFace(cutils.NormalFont),
		//Set the initial text for the textarea
		//It will automatically line wrap and process newlines characters
		//If ProcessBBCode is true it will parse out bbcode
		widget.TextAreaOpts.Text(`Release notes for v1.3.0:

Added:

* New UX/UI
* Created Itchio page
`),
		//Tell the TextArea to show the vertical scrollbar
		//widget.TextAreaOpts.ShowVerticalScrollbar(),
		//Set padding between edge of the widget and where the text is drawn
		widget.TextAreaOpts.TextPadding(widget.NewInsetsSimple(10)),
		//This sets the background images for the scroll container
		widget.TextAreaOpts.ScrollContainerOpts(
			widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
				Idle: cutils.LoadImageNineSlice(cutils.HomeWidgetsBGKey, 1, 1, !isPressed),
				Mask: image.NewNineSliceColor(cutils.Black),
			}),
		),
		//This sets the images to use for the sliders
		widget.TextAreaOpts.SliderOpts(
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
		),
	)

	infoC.AddChild(
		newsTAW,
		releaseNotesTAW,
	)

	warningC.AddChild(
		warnigTxtW,
	)

	contentC.AddChild(
		warningC,
		infoC,
	)

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
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Spacing(0, 30),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}),
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:   261,
				Right:  200,
				Top:    200,
				Bottom: 200,
			}),
		)),
	)

	learnBodyC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(3),
			widget.GridLayoutOpts.Spacing(2, 0),
			widget.GridLayoutOpts.Stretch([]bool{false, true, false}, []bool{true}),
		)),
	)

	contentC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		)),
	)

	generalC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(20),
		)),
	)

	wimwTxtW := widget.NewText(
		widget.TextOpts.Text("What is Maze Wars?", cutils.Font40, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	wimwBodyTxtW := widget.NewText(
		widget.TextOpts.Text(`
Maze Wars is a game in which you send units to the enemy player at your right and at the same time
you build a maze to protect from the units sent by the enemy at your left.

If a unit reaches the end of the line it'll steal a live from the player of that line and give it
to the owner of the unit, the unit will proceed to the next line until it dies or reaches the
owner line.

The winner is the player that steals the lives of all the other players.

`, cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	htsmwgTxtW := widget.NewText(
		widget.TextOpts.Text("How to start Maze Wars game?", cutils.Font40, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	// TODO: Change this text when we are no longer in Alpha or when matchmaking is
	// enabled globally on the game
	htsmwgBodyTxtW := widget.NewText(
		widget.TextOpts.Text(`
There are 2 different ways in which you can start a Maze Wars game.

Playing against bots by clicking "Play > (select 'Play with Bots?') > Start". This will
match you against a bot (1 or 5 depending on your selection). For this Alpha version
matchmaking with other players has been disabled momentarily.

Creating a custom lobby by going to "Lobbies > Create Lobby > Create" and then waiting for
people to joint this lobby.
`, cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	generalC.AddChild(
		wimwTxtW,
		wimwBodyTxtW,
		htsmwgTxtW,
		htsmwgBodyTxtW,
	)

	inGameC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			func(w *widget.Widget) {
				w.Visibility = widget.Visibility_Hide
			},
		),
	)

	igTxtW := widget.NewText(
		widget.TextOpts.Text("How to play?", cutils.Font40, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	igBodyTxtW := widget.NewText(
		widget.TextOpts.Text(`
Once the Game has started there are 4 main components:
* Your Line
* The display to summon Units and place Towers
* The information bar
* The Scoreboard
`, cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	igLineBodyTxtW := widget.NewText(
		widget.TextOpts.Text(`
There is one Line for each player, the line on the left is the one that sends units
to the player and the one on the right is the one receiving units from the player.
Your line is divided in 3 parts: top, middle and bottom.
The Top part of the Line is where the enemy units are summoned in a random position.
The Middle is where the player has to build the Maze with towers to kill the enemy
units.
The Bottom is where if the units reach it will steal one live and then move to the
line to the right.
`, cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	igDisplayBodyTxtW := widget.NewText(
		widget.TextOpts.Text(`
The Display can be seen at the bottom middle of the screen and it's divided in 2 
parts, the Units (left) and the Towers (right).
`, cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	igDisplayUnitsBodyTxtW := widget.NewText(
		widget.TextOpts.Text(`
There are 10 units that can be summoned using 1-0 keybinds (in the order they are
displayed, being 1 the first one and 0 the last one). When a unit can be updated
it'll display a dashed yellow inline. Shift clicking or Shift+Keybind will update
the unit making it stronger. When displayed in read it means there is not enough
gold to send it.
`, cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	igDisplayTowersBodyTxtW := widget.NewText(
		widget.TextOpts.Text(`
There are 2 types of towers that can be placed into the line, one range and the
other melee. Each one of those has 9 possible updates(Tiers).
To update a tower you have to select it and click the updates or Z or X keybinds.
Towers an also be sold for a 75% of the total cost.
`, cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	igInfoBarBodyTxtW := widget.NewText(
		widget.TextOpts.Text(`
The info bar can be seen in the top middle of the screen and has the following 
information in order from left to right:
* Time: Time that the game as been going.
* Gold: The amount of gold the player has.
* Capacity: The maximum amount of units that can be send.
* Lives: Total number of lives.
* Income: Amount of gold received every income tick.
* Income timer: Time left (every 15s) to receive the Income as gold.
`, cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	igScoreboardBodyTxtW := widget.NewText(
		widget.TextOpts.Text(`
The Scoreboard can be found by clicking TAB and it'll display the stats of all
the players on the current game.
`, cutils.NormalFont, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	inGameC.AddChild(
		igTxtW,
		igBodyTxtW,
		igLineBodyTxtW,
		igDisplayBodyTxtW,
		igDisplayUnitsBodyTxtW,
		igDisplayTowersBodyTxtW,
		igInfoBarBodyTxtW,
		igScoreboardBodyTxtW,
	)

	unitsC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				//Position: widget.RowLayoutPositionCenter,
				Stretch: true,
			}),
			func(w *widget.Widget) {
				w.Visibility = widget.Visibility_Hide
			},
		),
	)

	unitsFacetsC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(1),
			//Define how much padding to inset the child content
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(3)),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(2, 2),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{false}, []bool{false}),
		)),
		//widget.ContainerOpts.BackgroundImage(cutils.LoadImageNineSlice(cutils.ToolTipBGKey, 8, 8, !isPressed)),
		//widget.ContainerOpts.WidgetOpts(
		//widget.WidgetOpts.LayoutData(widget.RowLayoutData{
		//Position: widget.RowLayoutPositionCenter,
		//}),
		//),
		widget.ContainerOpts.WidgetOpts(
			func(w *widget.Widget) {
				w.Visibility = widget.Visibility_Hide
			},
		),
	)

	unitsDescriptionC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
	)

	unitsDescription := map[string]*widget.Container{}

	unitsBtns := make([]widget.RadioGroupElement, 0, 0)
	for _, t := range unit.TypeStrings() {
		aux := t
		u := unit.Units[aux]
		uu := store.CalculateUnitUpdate(aux, store.UnitUpdate{}, 1)
		imageBtnC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewStackedLayout()),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					MaxWidth:  46,
					MaxHeight: 46,
				}),
			),
		)
		ubtn := widget.NewButton(
			// specify the images to sue
			widget.ButtonOpts.Image(cutils.ButtonBorderResource()),

			// add a handler that reacts to clicking the button
			widget.ButtonOpts.ClickedHandler(func(t string) func(args *widget.ButtonClickedEventArgs) {
				return func(args *widget.ButtonClickedEventArgs) {
					for k, c := range unitsDescription {
						c.GetWidget().Visibility = widget.Visibility_Hide
						if k == aux {
							c.GetWidget().Visibility = widget.Visibility_Show
						}
					}
				}
			}(aux)),
		)
		imageGraphicW := widget.NewGraphic(
			widget.GraphicOpts.Image(
				cutils.Images.Get(unit.Units[aux].FacesetKey()),
			),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.MouseButtonPressedHandler(func(args *widget.WidgetMouseButtonPressedEventArgs) {
					//actionDispatcher.UserSignUpChangeImage(aux)
				}),
			),
		)
		imageBtnC.AddChild(
			ubtn,
			imageGraphicW,
		)
		unitsFacetsC.AddChild(imageBtnC)
		unitsBtns = append(unitsBtns, ubtn)

		unitInfoC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionVertical),
				widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(20)),
				widget.RowLayoutOpts.Spacing(5),
			)),
			widget.ContainerOpts.WidgetOpts(
				func(w *widget.Widget) {
					// If it's not the first one we hide it
					if aux != unit.TypeStrings()[0] {
						w.Visibility = widget.Visibility_Hide
					}
				},
			),
		)

		unitInfoDetailsC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewGridLayout(
				//Define number of columns in the grid
				//widget.GridLayoutOpts.Columns(2),
				widget.GridLayoutOpts.Columns(1),
				//Define how much padding to inset the child content
				//Define how far apart the rows and columns should be
				widget.GridLayoutOpts.Spacing(10, 0),
				//Define how to stretch the rows and columns. Note it is required to
				//specify the Stretch for each row and column.
				widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false}),
			)),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Stretch: true,
				}),
			),
			widget.ContainerOpts.AutoDisableChildren(),
		)

		unitInfoDetailsRows := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewGridLayout(
				//Define number of columns in the grid
				//widget.GridLayoutOpts.Columns(2),
				widget.GridLayoutOpts.Columns(3),
				//Define how far apart the rows and columns should be
				widget.GridLayoutOpts.Spacing(20, 0),
				//Define how to stretch the rows and columns. Note it is required to
				//specify the Stretch for each row and column.
				widget.GridLayoutOpts.Stretch([]bool{false, false, true}, []bool{false, false}),
			)),
		)

		ttTitleTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf(unitToolTipTitleTmpl, u.Name(), u.Keybind), cutils.Font40, cutils.White),
		)

		goldIconC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		)
		ttGoldGW := widget.NewGraphic(
			widget.GraphicOpts.Image(cutils.Images.Get(cutils.GoldIconKey)),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionStart,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		)
		goldIconC.AddChild(ttGoldGW)

		ttGoldTxtW := widget.NewText(
			widget.TextOpts.Text(fmt.Sprint(uu.Current.Gold), cutils.NormalFont, cutils.White),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
			),
		)

		incomeIconC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		)
		ttIncomeGW := widget.NewGraphic(
			widget.GraphicOpts.Image(cutils.Images.Get(cutils.IncomeIconKey)),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionStart,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		)
		incomeIconC.AddChild(ttIncomeGW)

		ttIncomeTxtW := widget.NewText(
			widget.TextOpts.Text(fmt.Sprint(uu.Current.Income), cutils.NormalFont, cutils.White),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
			),
		)

		healthIconC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		)
		ttHealthGW := widget.NewGraphic(
			widget.GraphicOpts.Image(cutils.Images.Get(cutils.HeartIconKey)),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionStart,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		)
		healthIconC.AddChild(ttHealthGW)

		ttHealthTxtW := widget.NewText(
			widget.TextOpts.Text(fmt.Sprint(uu.Current.Health), cutils.NormalFont, cutils.White),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
			),
		)

		movementSpeedIconC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		)
		ttMovementSpeedGW := widget.NewGraphic(
			widget.GraphicOpts.Image(cutils.Images.Get(cutils.MovementSpeedIconKey)),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionStart,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		)
		movementSpeedIconC.AddChild(ttMovementSpeedGW)
		ttMovementSpeedTxtW := widget.NewText(
			widget.TextOpts.Text(fmt.Sprint(uu.Current.MovementSpeed), cutils.NormalFont, cutils.White),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
			),
		)

		nGoldC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			)),
		)
		ttNextGoldTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf("%d", uu.Next.Gold), cutils.NormalFont, cutils.White),
		)
		ttNextGoldDiffTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf(" (+%d)", uu.Next.Gold-uu.Current.Gold), cutils.NormalFont, cutils.Green),
		)
		nGoldC.AddChild(ttNextGoldTxt, ttNextGoldDiffTxt)

		nIncomeC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			)),
		)
		ttNextIncomeTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf("%d", uu.Next.Income), cutils.NormalFont, cutils.White),
		)
		ttNextIncomeDiffTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf(" (+%d)", uu.Next.Income-uu.Current.Income), cutils.NormalFont, cutils.Green),
		)
		nIncomeC.AddChild(ttNextIncomeTxt, ttNextIncomeDiffTxt)

		nHealthC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			)),
		)
		ttNextHealthTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf("%.0f", uu.Next.Health), cutils.NormalFont, cutils.White),
		)
		ttNextHealthDiffTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf(" (+%.0f)", uu.Next.Health-uu.Current.Health), cutils.NormalFont, cutils.Green),
		)
		nHealthC.AddChild(ttNextHealthTxt, ttNextHealthDiffTxt)

		nMovementSpeedC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			)),
		)
		ttNextMovementSpeedTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf("%.0f", uu.Next.MovementSpeed), cutils.NormalFont, cutils.White),
		)
		ttNextMovementSpeedDiffTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text(fmt.Sprintf(" (+%.0f)", uu.Next.MovementSpeed-uu.Current.MovementSpeed), cutils.NormalFont, cutils.Green),
		)
		nMovementSpeedC.AddChild(ttNextMovementSpeedTxt, ttNextMovementSpeedDiffTxt)

		ttCurrentLvlTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text("1", cutils.NormalFont, cutils.White),
		)

		ttNextLvlTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text("2", cutils.NormalFont, cutils.White),
		)

		unitInfoDetailsRows.AddChild(
			widget.NewText(
				widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
				widget.TextOpts.Text("lvl", cutils.NormalFont, cutils.White),
			),
			ttCurrentLvlTxt,
			ttNextLvlTxt,

			goldIconC,
			ttGoldTxtW,
			nGoldC,

			incomeIconC,
			ttIncomeTxtW,
			nIncomeC,

			healthIconC,
			ttHealthTxtW,
			nHealthC,

			movementSpeedIconC,
			ttMovementSpeedTxtW,
			nMovementSpeedC,
		)

		unitInfoDetailsC.AddChild(
			unitInfoDetailsRows,
		)

		ttAbilitiesTxt := widget.NewText(
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.Text("Abilities:", cutils.NormalFont, cutils.White),
		)

		unitGIFC := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			)),
		)

		if g, ok := unitTypeGif[u.Type]; ok {
			unitGIFG := widget.NewGraphic(
				widget.GraphicOpts.GIF(g),
			)
			unitGIFC.AddChild(unitGIFG)
		}

		unitInfoC.AddChild(
			ttTitleTxt,
			unitInfoDetailsC,
			ttAbilitiesTxt,
		)

		for _, a := range u.Abilities {
			unitInfoC.AddChild(widget.NewText(
				widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
				widget.TextOpts.Text(fmt.Sprintf("%s: %s", ability.Name(a), ability.Description(a)), cutils.NormalFont, cutils.White),
			))
		}
		unitInfoC.AddChild(
			unitGIFC,
		)

		unitsDescription[aux] = unitInfoC
	}

	widget.NewRadioGroup(
		widget.RadioGroupOpts.Elements(unitsBtns...),
		widget.RadioGroupOpts.InitialElement(unitsBtns[0]),
	)

	for _, c := range unitsDescription {
		unitsDescriptionC.AddChild(c)
	}

	unitsC.AddChild(
		unitsDescriptionC,
	)

	towersC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(20),
		)),
		widget.ContainerOpts.WidgetOpts(
			func(w *widget.Widget) {
				w.Visibility = widget.Visibility_Hide
			},
		),
	)

	towersTxtW := widget.NewText(
		widget.TextOpts.Text("List of Towers", cutils.Font40, cutils.TextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	towersC.AddChild(
		towersTxtW,
	)

	contentC.AddChild(
		generalC,
		inGameC,
		unitsC,
		towersC,
	)

	//Create the new ScrollContainer object
	scrollContainer := widget.NewScrollContainer(
		//Set the content that will be scrolled
		widget.ScrollContainerOpts.Content(contentC),
		//Tell the container to stretch the content width to match available space
		widget.ScrollContainerOpts.StretchContentWidth(),
		//Set the background images for the scrollable container
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle: image.NewNineSliceColor(cutils.Transparent),
			Mask: image.NewNineSliceColor(cutils.Black),
		}),
	)

	pageSizeFunc := func() int {
		return int(math.Round(float64(scrollContainer.ViewRect().Dy()) / float64(contentC.GetWidget().Rect.Dy()) * 1000))
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

	navbarC := widget.NewContainer(
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

	generalBtnW := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.TextButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("General", cutils.NormalFont, &cutils.ButtonTextColorNoPressed),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   10,
			Right:  10,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			generalC.GetWidget().Visibility = widget.Visibility_Show

			inGameC.GetWidget().Visibility = widget.Visibility_Hide
			unitsC.GetWidget().Visibility = widget.Visibility_Hide
			towersC.GetWidget().Visibility = widget.Visibility_Hide
			unitsFacetsC.GetWidget().Visibility = widget.Visibility_Hide
		}),
	)
	inGameBtnW := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.TextButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("In Game", cutils.NormalFont, &cutils.ButtonTextColorNoPressed),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   10,
			Right:  10,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			inGameC.GetWidget().Visibility = widget.Visibility_Show

			generalC.GetWidget().Visibility = widget.Visibility_Hide
			unitsC.GetWidget().Visibility = widget.Visibility_Hide
			towersC.GetWidget().Visibility = widget.Visibility_Hide
			unitsFacetsC.GetWidget().Visibility = widget.Visibility_Hide
		}),
	)
	unitsBtnW := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.TextButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Units", cutils.NormalFont, &cutils.ButtonTextColorNoPressed),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   10,
			Right:  10,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			unitsC.GetWidget().Visibility = widget.Visibility_Show
			unitsFacetsC.GetWidget().Visibility = widget.Visibility_Show

			generalC.GetWidget().Visibility = widget.Visibility_Hide
			inGameC.GetWidget().Visibility = widget.Visibility_Hide
			towersC.GetWidget().Visibility = widget.Visibility_Hide
		}),
	)
	towersBtnW := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),

		// specify the images to sue
		widget.ButtonOpts.Image(cutils.TextButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Towers", cutils.NormalFont, &cutils.ButtonTextColorNoPressed),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   10,
			Right:  10,
			Top:    5,
			Bottom: 5,
		}),

		// add a handler that reacts to clicking the button
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			towersC.GetWidget().Visibility = widget.Visibility_Show

			generalC.GetWidget().Visibility = widget.Visibility_Hide
			inGameC.GetWidget().Visibility = widget.Visibility_Hide
			unitsC.GetWidget().Visibility = widget.Visibility_Hide
			unitsFacetsC.GetWidget().Visibility = widget.Visibility_Hide
		}),
	)

	widget.NewRadioGroup(
		widget.RadioGroupOpts.Elements(generalBtnW, inGameBtnW, unitsBtnW, towersBtnW),
		widget.RadioGroupOpts.InitialElement(generalBtnW),
	)

	navbarC.AddChild(
		generalBtnW,
		inGameBtnW,
		unitsBtnW,
		towersBtnW,
	)

	learnBodyC.AddChild(
		unitsFacetsC,
		scrollContainer,
		vSlider,
	)

	learnC.AddChild(
		navbarC,
		learnBodyC,
	)

	rs.learnC = learnC

	return learnC
}
