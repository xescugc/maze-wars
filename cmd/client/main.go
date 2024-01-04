package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/client"
	"github.com/xescugc/maze-wars/inputer"
	"github.com/xescugc/maze-wars/store"
)

var (
	hostURL string
	screenW int
	screenH int
	verbose bool

	clientCmd = &cobra.Command{
		Use: "client",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			opt := client.Options{
				HostURL: hostURL,
				ScreenW: screenW,
				ScreenH: screenH,
			}

			d := flux.NewDispatcher()
			if verbose {
				client.NewLoggerStore(d)
			}
			ad := client.NewActionDispatcher(d, opt)

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

			us := client.NewUserStore(d)
			cls := client.NewStore(s, us)

			l, err := client.NewLobbyStore(d, i, cls)
			if err != nil {
				return fmt.Errorf("failed to initialize LobbyStore: %w", err)
			}

			su, err := client.NewSignUpStore(d, i, s)
			if err != nil {
				return fmt.Errorf("failed to initial SignUpStore: %w", err)
			}
			wr := client.NewWaitingRoomStore(d, cls)

			rs := client.NewRouterStore(d, su, l, wr, g)
			ctx := context.Background()

			err = client.New(ctx, ad, rs, opt)

			return err
		},
	}
)

func init() {
	clientCmd.Flags().StringVar(&hostURL, "port", "localhost:5555", "The URL of the server")
	clientCmd.Flags().IntVar(&screenW, "screenw", 288, "The default width of the screen when not full screen")
	clientCmd.Flags().IntVar(&screenH, "screenh", 240, "The default height of the screen when not full screen")
	clientCmd.Flags().BoolVar(&verbose, "verbose", false, "Logs information of the running client")
}

func main() {
	if err := clientCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)

	}
}
