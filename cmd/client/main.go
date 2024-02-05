package main

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/sagikazarmark/slog-shim"
	"github.com/spf13/cobra"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/client"
	"github.com/xescugc/maze-wars/store"
)

var (
	defaultHost = "http://localhost:5555"
	logFile     = path.Join(xdg.CacheHome, "maze-wars", "client.log")

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
			s := store.NewStore(d)

			lvl := slog.LevelInfo
			if verbose {
				lvl = slog.LevelDebug
			}
			err = os.MkdirAll(path.Dir(logFile), 0700)
			if err != nil {
				return err
			}
			f, err := os.OpenFile(logFile, os.O_APPEND|os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				return err
			}

			defer f.Close()

			l := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{
				Level: lvl,
			}))
			ad := client.NewActionDispatcher(d, s, l, opt)

			g := &client.Game{
				Store: s,
			}

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

			ls, err := client.NewLobbyStore(d, cls)
			if err != nil {
				return fmt.Errorf("failed to initialize LobbyStore: %w", err)
			}

			su, err := client.NewSignUpStore(d, s)
			if err != nil {
				return fmt.Errorf("failed to initial SignUpStore: %w", err)
			}
			wr := client.NewWaitingRoomStore(d, cls)

			rs := client.NewRouterStore(d, su, ls, wr, g)
			ctx := context.Background()

			err = client.New(ctx, ad, rs, opt)

			return err
		},
	}
)

func init() {
	clientCmd.Flags().StringVar(&hostURL, "host", defaultHost, "The URL of the server")
	clientCmd.Flags().IntVar(&screenW, "screenw", 288, "The default width of the screen when not full screen")
	clientCmd.Flags().IntVar(&screenH, "screenh", 240, "The default height of the screen when not full screen")
	clientCmd.Flags().BoolVar(&verbose, "verbose", false, fmt.Sprintf("If all the logs are gonna be printed to %s", logFile))
}

func main() {
	if err := clientCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)

	}
}
