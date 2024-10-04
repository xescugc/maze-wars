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
	"github.com/xescugc/maze-wars/client/game"
	"github.com/xescugc/maze-wars/store"
)

const isOnServer = true

func main() {
	js.Global().Set("new_client", NewClient())
	select {}
}

func NewClient() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 2 || (args[0].String() == "") || (args[1].String() == "") {
			return fmt.Errorf("requires 2 parameter: host, version")
		}
		var (
			err     error
			hostURL = args[0].String()
			version = args[1].String()
			screenW = 550
			screenH = 500
			opt     = client.Options{
				HostURL: hostURL,
				ScreenW: screenW,
				ScreenH: screenH,
				Version: version,
			}
		)

		l := slog.New(slog.NewTextHandler(ioutil.Discard, nil))

		d := flux.NewDispatcher()
		s := store.NewStore(d, l, !isOnServer)

		ad := client.NewActionDispatcher(d, s, l, opt)

		g := client.NewGame(s, d, l)

		cs := game.NewCameraStore(d, s, l, screenW, screenH)
		g.Game.Camera = cs
		g.Game.Lines, err = game.NewLines(g.Game)
		if err != nil {
			return fmt.Errorf("failed to initialize Lines: %w", err)
		}

		g.Game.HUD, err = game.NewHUDStore(d, g.Game)
		if err != nil {
			return fmt.Errorf("failed to initialize HUDStore: %w", err)
		}

		g.Game.Map = game.NewMap(g.Game)

		us := client.NewUserStore(d)
		cls := client.NewStore(s, us)

		ros, err := client.NewRootStore(d, cls, l)
		if err != nil {
			return fmt.Errorf("failed to initialize RootStore: %w", err)
		}

		u, err := client.NewSignUpStore(d, s, "", "", l)
		if err != nil {
			return fmt.Errorf("failed to initial SignUpStore: %w", err)
		}

		rs := client.NewRouterStore(d, u, ros, g, l)

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
