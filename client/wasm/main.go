//go:build js && wasm

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"log/slog"
	"syscall/js"

	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/client"
	"github.com/xescugc/maze-wars/store"
)

func main() {
	js.Global().Set("new_client", NewClient())
	select {}
}

func NewClient() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 || (args[0].String() == "") {
			return fmt.Errorf("requires 1 parameter: host")
		}
		var (
			err     error
			hostURL = args[0].String()
			screenW = 288
			screenH = 240
			opt     = client.Options{
				HostURL: hostURL,
				ScreenW: screenW,
				ScreenH: screenH,
			}
		)

		l := slog.New(slog.NewTextHandler(ioutil.Discard, nil))

		d := flux.NewDispatcher()
		s := store.NewStore(d, l)

		ad := client.NewActionDispatcher(d, s, l, opt)

		g := client.NewGame(s, l)

		// TODO: Change this to pass the specific store needed instead of all the game object
		cs := client.NewCameraStore(d, s, l, screenW, screenH)
		g.Camera = cs
		g.Lines, err = client.NewLines(g)
		if err != nil {
			return fmt.Errorf("failed to initialize Lines: %w", err)
		}

		g.HUD, err = client.NewHUDStore(d, g)
		if err != nil {
			return fmt.Errorf("failed to initialize HUDStore: %w", err)
		}

		g.Map = client.NewMap(g)

		us := client.NewUserStore(d)
		cls := client.NewStore(s, us)

		ls, err := client.NewLobbyStore(d, cls, l)
		if err != nil {
			return fmt.Errorf("failed to initialize LobbyStore: %w", err)
		}

		u, err := client.NewSignUpStore(d, s, l)
		if err != nil {
			return fmt.Errorf("failed to initial SignUpStore: %w", err)
		}

		wr := client.NewWaitingRoomStore(d, cls, l)

		rs := client.NewRouterStore(d, u, ls, wr, g, l)

		ctx := context.Background()
		// We need to run this in a goroutine so when it's compiled to WASM
		// it does not block the main thread https://github.com/golang/go/issues/41310
		go func() {
			err = client.New(ctx, ad, rs, opt)
			if err != nil {
				log.Fatal(err)
			}
		}()
		return nil
	})
}
