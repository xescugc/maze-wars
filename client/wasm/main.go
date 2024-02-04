//go:build js && wasm

package main

import (
	"context"
	"fmt"
	"log"
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

		d := flux.NewDispatcher()
		s := store.NewStore(d)

		ad := client.NewActionDispatcher(d, s, opt)

		g := &client.Game{
			Store: s,
		}

		// TODO: Change this to pass the specific store needed instead of all the game object
		cs := client.NewCameraStore(d, s, screenW, screenH)
		g.Camera = cs
		g.Units, err = client.NewUnits(g)
		if err != nil {
			return fmt.Errorf("failed to initialize Units: %w", err)
		}

		g.Towers, err = client.NewTowers(g)
		if err != nil {
			return fmt.Errorf("failed to initialize Towers: %w", err)
		}

		g.HUD, err = client.NewHUDStore(d, g)
		if err != nil {
			return fmt.Errorf("failed to initialize HUDStore: %w", err)
		}

		g.Map = client.NewMap(g)

		us := client.NewUserStore(d)
		cls := client.NewStore(s, us)

		l, err := client.NewLobbyStore(d, cls)
		if err != nil {
			return fmt.Errorf("failed to initialize LobbyStore: %w", err)
		}

		u, err := client.NewSignUpStore(d, s)
		if err != nil {
			return fmt.Errorf("failed to initial SignUpStore: %w", err)
		}

		wr := client.NewWaitingRoomStore(d, cls)

		rs := client.NewRouterStore(d, u, l, wr, g)

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
