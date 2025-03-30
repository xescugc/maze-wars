package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path"
	"time"

	_ "net/http/pprof"

	"github.com/adrg/xdg"
	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/client"
	"github.com/xescugc/maze-wars/client/game"
	cutils "github.com/xescugc/maze-wars/client/utils"
	"github.com/xescugc/maze-wars/store"
)

const isOnServer = true

type config struct {
	Username string `json:"username"`
	ImageKey string `json:"image_key"`
}

var (
	logFile = path.Join(xdg.CacheHome, "maze-wars", "client.log")

	hostURL string
	screenW int
	screenH int
	verbose bool

	clientCmd = &cobra.Command{
		Use: "client",
		RunE: func(cmd *cobra.Command, args []string) error {
			go func() {
				log.Println(http.ListenAndServe("localhost:6061", nil))
			}()
			var err error
			opt := client.Options{
				ScreenW: screenW,
				ScreenH: screenH,
				Version: client.Version,
				HostURL: client.Host,
			}

			configFilePath, err := xdg.ConfigFile("maze-wars/user.json")
			if err != nil {
				return err
			}
			b, err := os.ReadFile(configFilePath)
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			var cfg cutils.Config
			err = json.Unmarshal(b, &cfg)

			d := flux.NewDispatcher[*action.Action]()

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

			su, err := client.NewSignUpStore(d, s, cfg.Username, cfg.ImageKey, l)
			if err != nil {
				return fmt.Errorf("failed to initial SignUpStore: %w", err)
			}

			rs := client.NewRouterStore(d, su, ros, g, l)
			ctx := context.Background()

			err = client.New(ctx, ad, rs, opt)

			return err
		},
	}
)

func init() {
	clientCmd.Flags().StringVar(&hostURL, "host", client.Host, "The URL of the server")
	clientCmd.Flags().IntVar(&screenW, "screenw", 550, "The default width of the screen when not full screen")
	clientCmd.Flags().IntVar(&screenH, "screenh", 500, "The default height of the screen when not full screen")
	clientCmd.Flags().BoolVar(&verbose, "verbose", false, fmt.Sprintf("If all the logs are gonna be printed to %s", logFile))

	clientCmd.AddCommand(versionCmd)
}

func main() {
	err := sentry.Init(sentry.ClientOptions{
		// Either set your DSN here or set the SENTRY_DSN environment variable.
		Dsn: "https://a2778d36fdcdf66eea5ca37b403ddc5f@o4509005974667264.ingest.de.sentry.io/4509018830864464",
		// Enable printing of SDK debug messages.
		// Useful when getting started or trying to figure something out.
		EnableTracing:    true,
		Release:          client.Version,
		AttachStacktrace: true,
		Environment:      client.Environment,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	// Flush buffered events before the program terminates.
	// Set the timeout to the maximum duration the program can afford to wait.
	defer func() {
		err := recover()

		if err != nil {
			sentry.CurrentHub().Recover(err)
			sentry.Flush(time.Second * 5)
			if client.Environment == "dev" {
				panic(err)
			}
		}
	}()

	if err := clientCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
