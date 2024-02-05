package integration_test

import (
	"context"
	"io/ioutil"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/client"
	"github.com/xescugc/maze-wars/server"
	"github.com/xescugc/maze-wars/store"
)

var (
	// The actual one is 4
	serverGameTick = time.Second / 2

	// The actual one is 60
	clientTPS = time.Second / 30
)

func TestRun(t *testing.T) {
	if os.Getenv("IS_CI") == "true" {
		t.Skip("This test are skipped for now on the CI")
	}
	var (
		err     error
		screenW = 288
		screenH = 240
		//players = make(map[string]*store.Player)
	)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ss := &server.Store{}
	sd := flux.NewDispatcher()
	sl := slog.New(slog.NewTextHandler(ioutil.Discard, nil))
	sad := server.NewActionDispatcher(sd, sl, ss)
	rooms := server.NewRoomsStore(sd, ss)
	users := server.NewUsersStore(sd, ss)

	ss.Rooms = rooms
	ss.Users = users

	// Start the Server
	go func() {
		err := server.New(sad, ss, server.Options{
			Port: "5555",
		})
		require.NoError(t, err)
	}()

	copt := client.Options{
		HostURL: "localhost:5555",
		ScreenW: screenW,
		ScreenH: screenH,
	}
	cd := flux.NewDispatcher()
	s := store.NewStore(cd)

	cl := slog.New(slog.NewTextHandler(ioutil.Discard, nil))
	cad := client.NewActionDispatcher(cd, s, cl, copt)

	g := &client.Game{
		Store: s,
	}

	cs := client.NewCameraStore(cd, s, screenW, screenH)
	g.Camera = cs
	g.Units, err = client.NewUnits(g)
	require.NoError(t, err)

	g.Towers, err = client.NewTowers(g)
	require.NoError(t, err)

	g.HUD, err = client.NewHUDStore(cd, g)
	require.NoError(t, err)

	us := client.NewUserStore(cd)
	cls := client.NewStore(s, us)

	ls, err := client.NewLobbyStore(cd, cls)
	require.NoError(t, err)

	wr := client.NewWaitingRoomStore(cd, cls)

	su, err := client.NewSignUpStore(cd, s)
	require.NoError(t, err)

	rs := client.NewRouterStore(cd, su, ls, wr, g)

	// Before starting we give the server
	// some time to start
	wait(time.Second * 2)

	// It's not possible to change a registered EXPECT, so we need to
	// be able to change the content of the expectations.
	// That's where this parameters help, they can change the content
	// of the expectation and what they return
	//var (
	//x, y int

	//mouseButtonJustPressed       ebiten.MouseButton
	//returnMouseButtonJustPressed bool

	//keyJustPressed       ebiten.Key
	//returnKeyJustPressed bool
	//)
	//resetDefault := func() {
	//x, y = 0, 0
	//mouseButtonJustPressed = 0
	//returnMouseButtonJustPressed = false

	//keyJustPressed = 0
	//returnKeyJustPressed = false
	//}
	//i.EXPECT().CursorPosition().DoAndReturn(func() (int, int) {
	//return x, y
	//}).AnyTimes()
	//i.EXPECT().IsMouseButtonJustPressed(gomock.Any()).DoAndReturn(func(button ebiten.MouseButton) bool {
	//if returnMouseButtonJustPressed {
	//assert.Equal(t, mouseButtonJustPressed, button)
	//}
	//return returnMouseButtonJustPressed
	//}).AnyTimes()
	//i.EXPECT().IsKeyJustPressed(gomock.Any()).DoAndReturn(func(key ebiten.Key) bool {
	//if returnKeyJustPressed {
	//assert.Equal(t, keyJustPressed, key)
	//}
	//return returnKeyJustPressed
	//}).AnyTimes()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		err = client.New(ctx, cad, rs, copt)
		require.NoError(t, err)
	}()

	// To run the 2nd client we need to exec it locally
	go func() {
		cmd := exec.CommandContext(ctx, "go", "run", "../cmd/client/")
		err = cmd.Run()
		require.NoError(t, err)
	}()

	//t.Run("Player added to the room", func(t *testing.T) {
	//var (
	//tries int
	//)
	//// Since the second player is initialized via "exec" the times of being ready could be different
	//// between computers so we do 10 tries before failing

	//ros := rooms.GetState().(server.RoomsState)

	//for len(rooms.GetState().(server.RoomsState).Rooms) != 1 || len(ros.Rooms[room].Game.Players.List()) != 2 {
	//if tries == 10 {
	//t.Fatal(t, "Could not initialize the players")
	//}
	//ros = rooms.GetState().(server.RoomsState)
	//time.Sleep(time.Second)
	//tries++
	//}
	//for _, p := range ros.Rooms[room].Game.Players.List() {
	//players[p.Name] = p
	//}

	//lst := l.GetState().(client.LobbyState)
	//x = int(lst.YesBtn.X + 1)
	//y = int(lst.YesBtn.Y + 1)

	//returnMouseButtonJustPressed = true
	//mouseButtonJustPressed = ebiten.MouseButtonLeft

	//wait()
	//resetDefault()
	//wait(serverGameTick)

	//for _, p := range rooms.GetState().(server.RoomsState).Rooms[room].Game.Players.List() {
	//if p.Name == p1n {
	//assert.True(t, p.Ready)
	//}
	//}

	//require.Equal(t, client.UsernameRoute, rs.GetState().(client.RouterState).Route)

	//// We mark 2 players as ready
	//sad.Dispatch(action.NewPlayerReady(players[p2n].ID))
	//for _, p := range rooms.GetState().(server.RoomsState).Rooms[room].Game.Players.List() {
	//assert.True(t, p.Ready)
	//}

	//wait(serverGameTick)
	//// Once the 2 players are ready the clients move to the game route
	//require.Equal(t, client.GameRoute, rs.GetState().(client.RouterState).Route)
	//})
}

// wait waits for the desired first duration if not then for time.Second/30
// which is double of the time of the internal Ebiten TPS(60/s)
func wait(d ...time.Duration) {
	if len(d) == 0 {
		d = []time.Duration{clientTPS}
	}
	runtime.Gosched()
	time.Sleep(d[0])
}
