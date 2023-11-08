package main

import (
	"bytes"
	"image"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/assets"
	"github.com/xescugc/ltw/store"
	"github.com/xescugc/ltw/utils"
)

type Lobby struct {
	*flux.ReduceStore

	Store *store.Store

	Camera *CameraStore
	YesBtn image.Image
}

func NewLobby(d *flux.Dispatcher, s *store.Store, c *CameraStore) (*Lobby, error) {
	bi, _, err := image.Decode(bytes.NewReader(assets.YesButton_png))
	if err != nil {
		return nil, err
	}

	l := &Lobby{
		Store:  s,
		Camera: c,

		YesBtn: ebiten.NewImageFromImage(bi),
	}
	return l, nil
}

func (l *Lobby) Update() error {
	l.Camera.Update()
	cs := l.Camera.GetState().(CameraState)
	x, y := ebiten.CursorPosition()
	// TODO: Fix all this so it's not calculated each time but stored
	// the button position
	ybtn := utils.Object{
		X: float64(cs.W - float64(l.YesBtn.Bounds().Dx())),
		Y: float64(cs.H - float64(l.YesBtn.Bounds().Dy())),
		W: float64(l.YesBtn.Bounds().Dx()),
		H: float64(l.YesBtn.Bounds().Dy()),
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		obj := utils.Object{
			X: float64(x),
			Y: float64(y),
			W: 1, H: 1,
		}
		if ybtn.IsColliding(obj) {
			cp := l.Store.Players.GetCurrentPlayer()
			actionDispatcher.PlayerReady(cp.ID)
		}
	}

	players := l.Store.Players.GetPlayers()
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
		}
	}

	return nil
}

func (l *Lobby) Draw(screen *ebiten.Image) {
	cs := l.Camera.GetState().(CameraState)
	ps := l.Store.Players.GetPlayers()
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
	ybop.GeoM.Translate(float64(cs.W-float64(l.YesBtn.Bounds().Dx())), float64(cs.H-float64(l.YesBtn.Bounds().Dy())))
	screen.DrawImage(l.YesBtn.(*ebiten.Image), ybop)
}
