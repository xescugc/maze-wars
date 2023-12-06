//go:build js && wasm

package main

import (
	"context"
	"fmt"
	"log"
	"syscall/js"

	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/client"
	"github.com/xescugc/ltw/inputer"
	"github.com/xescugc/ltw/store"
)

func main() {
	js.Global().Set("new_client", NewClient())
	select {}
}

func NewClient() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 3 || (args[0].String() == "" || args[1].String() == "" || args[2].String() == "") {
			return fmt.Errorf("requires 3 parameters: host, room and name")
		}
		var (
			err     error
			hostURL = args[0].String()
			room    = args[1].String()
			name    = args[2].String()
			screenW = 288
			screenH = 240
		)

		d := flux.NewDispatcher()
		ad := client.NewActionDispatcher(d)

		s := store.NewStore(d)

		g := &client.Game{
			Store: s,
		}

		i := inputer.NewEbiten()

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

		g.HUD, err = client.NewHUDStore(d, i, g)
		if err != nil {
			return fmt.Errorf("failed to initialize HUDStore: %w", err)
		}

		l, err := client.NewLobbyStore(d, i, s, cs)
		if err != nil {
			return fmt.Errorf("failed to initialize LobbyStore: %w", err)
		}
		rs := client.NewRouterStore(d, g, l)

		ctx := context.Background()
		// We need to run this in a goroutine so when it's compiled to WASM
		// it does not block the main thread https://github.com/golang/go/issues/41310
		go func() {
			err = client.New(ctx, ad, rs, client.Options{
				HostURL: hostURL,
				Room:    room,
				Name:    name,
				ScreenW: screenW,
				ScreenH: screenH,
			})
			if err != nil {
				log.Fatal(err)
			}
		}()
		return nil
	})
}
