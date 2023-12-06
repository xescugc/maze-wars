package client

import (
	"bytes"
	"image"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/action"
	"github.com/xescugc/ltw/assets"
	"github.com/xescugc/ltw/inputer"
	"github.com/xescugc/ltw/store"
	"github.com/xescugc/ltw/utils"
)

type LobbyStore struct {
	*flux.ReduceStore

	Store *store.Store

	Camera *CameraStore
	YesBtn image.Image

	input inputer.Inputer
}

type LobbyState struct {
	YesBtn utils.Object
}

func NewLobbyStore(d *flux.Dispatcher, i inputer.Inputer, s *store.Store, cs *CameraStore) (*LobbyStore, error) {
	bi, _, err := image.Decode(bytes.NewReader(assets.YesButton_png))
	if err != nil {
		return nil, err
	}

	ls := &LobbyStore{
		Store:  s,
		Camera: cs,

		YesBtn: ebiten.NewImageFromImage(bi),

		input: i,
	}
	cst := cs.GetState().(CameraState)
	ls.ReduceStore = flux.NewReduceStore(d, ls.Reduce, LobbyState{
		YesBtn: utils.Object{
			X: float64(cst.W - float64(ls.YesBtn.Bounds().Dx())),
			Y: float64(cst.H - float64(ls.YesBtn.Bounds().Dy())),
			W: float64(ls.YesBtn.Bounds().Dx()),
			H: float64(ls.YesBtn.Bounds().Dy()),
		},
	})
	return ls, nil
}

func (ls *LobbyStore) Update() error {
	ls.Camera.Update()
	x, y := ls.input.CursorPosition()
	lst := ls.GetState().(LobbyState)
	// TODO: Fix all this so it's not calculated each time but stored
	// the button position
	if ls.input.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		obj := utils.Object{
			X: float64(x),
			Y: float64(y),
			W: 1, H: 1,
		}
		if lst.YesBtn.IsColliding(obj) {
			cp := ls.Store.Players.FindCurrent()
			actionDispatcher.PlayerReady(cp.ID)
		}
	}

	players := ls.Store.Players.List()
	if len(players) > 1 {
		allReady := true
		for _, p := range players {
			if !p.Ready {
				allReady = false
				break
			}
		}
		if allReady {
			actionDispatcher.NavigateTo(GameRoute)
			actionDispatcher.StartGame()
			actionDispatcher.GoHome()
		}
	}

	return nil
}

func (ls *LobbyStore) Draw(screen *ebiten.Image) {
	cs := ls.Camera.GetState().(CameraState)
	ps := ls.Store.Players.List()
	lst := ls.GetState().(LobbyState)
	text.Draw(screen, "LOBBY", normalFont, int(cs.W/2), int(cs.H/2), color.White)
	var pcount = 1
	var sortedPlayers = make([]*store.Player, 0, 0)
	for _, p := range ps {
		sortedPlayers = append(sortedPlayers, p)
	}
	sort.Slice(sortedPlayers, func(i, j int) bool { return sortedPlayers[i].LineID < sortedPlayers[j].LineID })
	for _, p := range sortedPlayers {
		var c color.Color = color.White
		if p.Ready {
			c = green
		}
		text.Draw(screen, p.Name, normalFont, int(cs.W/2), int(cs.H/2)+(24*pcount), c)
		pcount++
	}

	ybop := &ebiten.DrawImageOptions{}
	ybop.GeoM.Translate(lst.YesBtn.X, lst.YesBtn.Y)
	screen.DrawImage(ls.YesBtn.(*ebiten.Image), ybop)
}

func (ls *LobbyStore) Reduce(state, a interface{}) interface{} {
	act, ok := a.(*action.Action)
	if !ok {
		return state
	}

	lstate, ok := state.(LobbyState)
	if !ok {
		return state
	}

	switch act.Type {
	case action.WindowResizing:
		ls.GetDispatcher().WaitFor(ls.Camera.GetDispatcherToken())
		cs := ls.Camera.GetState().(CameraState)
		lstate.YesBtn = utils.Object{
			X: float64(cs.W - float64(ls.YesBtn.Bounds().Dx())),
			Y: float64(cs.H - float64(ls.YesBtn.Bounds().Dy())),
			W: float64(ls.YesBtn.Bounds().Dx()),
			H: float64(ls.YesBtn.Bounds().Dy()),
		}
	}

	return lstate
}
