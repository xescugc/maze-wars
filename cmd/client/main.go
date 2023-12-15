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
	room    string
	name    string
	hostURL string
	screenW int
	screenH int
	verbose bool

	clientCmd = &cobra.Command{
		Use: "client",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			d := flux.NewDispatcher()
			if verbose {
				client.NewLoggerStore(d)
			}
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

			err = client.New(ctx, ad, rs, client.Options{
				HostURL: hostURL,
				Room:    room,
				Name:    name,
				ScreenW: screenW,
				ScreenH: screenH,
			})

			return err
		},
	}
)

func init() {
	clientCmd.Flags().StringVar(&hostURL, "port", "localhost:5555", "The URL of the server")
	clientCmd.Flags().StringVar(&room, "room", "room", "The room name to join")
	clientCmd.Flags().StringVar(&name, "name", "john doe", "The name of the client")
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