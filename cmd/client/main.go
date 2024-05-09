package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/client"
	"github.com/xescugc/maze-wars/client/game"
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
				Version: version,
			}

			d := flux.NewDispatcher()

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

			s := store.NewStore(d, l)
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

			su, err := client.NewSignUpStore(d, s, l)
			if err != nil {
				return fmt.Errorf("failed to initial SignUpStore: %w", err)
			}

			wr := client.NewWaitingRoomStore(d, cls, l)

			lv := client.NewLobbiesView(cls, l)
			nlv := client.NewNewLobbyView(cls, l)
			slv := client.NewShowLobbyView(cls, l)

			rs := client.NewRouterStore(d, su, ros, wr, g, lv, nlv, slv, l)
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

	clientCmd.AddCommand(versionCmd)
}

func main() {
	if err := clientCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)

	}
}
