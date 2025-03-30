package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/adrg/xdg"
	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/server"
)

var (
	logFile = path.Join(xdg.CacheHome, "maze-wars", "server.log")

	serverCmd = &cobra.Command{
		Use: "server",
		RunE: func(cmd *cobra.Command, args []string) error {

			go func() {
				log.Println(http.ListenAndServe("localhost:6060", nil))
			}()

			opt := server.Options{
				Port:             viper.GetString("port"),
				Verbose:          viper.GetBool("verbose"),
				Version:          server.Version,
				DiscordBotToken:  viper.GetString("discord-bot-token"),
				DiscordChannelID: viper.GetString("discord-channel-id"),
			}
			d := flux.NewDispatcher[*action.Action]()

			out := os.Stdout
			lvl := slog.LevelInfo
			if opt.Verbose {
				lvl = slog.LevelDebug
				err := os.MkdirAll(path.Dir(logFile), 0700)
				if err != nil {
					return err
				}
				f, err := os.OpenFile(logFile, os.O_APPEND|os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
				if err != nil {
					return err
				}
				out = f

				defer f.Close()
			}
			l := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{
				Level: lvl,
			}))

			var (
				dgo *discordgo.Session
				err error
			)
			if server.Environment == "prod" {
				dgo, err := discordgo.New("Bot " + opt.DiscordBotToken)
				if err != nil {
					return err
				}

				err = dgo.Open()
				if err != nil {
					return err
				}
				defer dgo.Close()
			}

			ws := server.NewWS()
			ss := server.NewStore(d, ws, dgo, opt, l)
			ad := server.NewActionDispatcher(d, l, ss, ws)

			err = server.New(ad, ss, opt)
			if err != nil {
				return fmt.Errorf("server error: %w", err)
			}

			return nil
		},
	}
)

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	serverCmd.Flags().String("port", "5555", "The port in which the sever is open")
	viper.BindPFlag("port", serverCmd.Flags().Lookup("port"))

	serverCmd.Flags().String("discord-bot-token", "", "The TOKEN to send Discord notifications")
	viper.BindPFlag("discord-bot-token", serverCmd.Flags().Lookup("discord-bot-token"))

	serverCmd.Flags().String("discord-channel-id", "", "The ID for the Channel to send messages")
	viper.BindPFlag("discord-channel-id", serverCmd.Flags().Lookup("discord-channel-id"))

	serverCmd.Flags().Bool("verbose", false, fmt.Sprintf("If all the logs are gonna be printed to %s", logFile))
	viper.BindPFlag("verbose", serverCmd.Flags().Lookup("verbose"))

	serverCmd.AddCommand(versionCmd)
}

func main() {
	err := sentry.Init(sentry.ClientOptions{
		// Either set your DSN here or set the SENTRY_DSN environment variable.
		Dsn: "https://e3d184a556b90e462a15294b172a336e@o4509005974667264.ingest.de.sentry.io/4509006005338192",
		// Enable printing of SDK debug messages.
		// Useful when getting started or trying to figure something out.
		EnableTracing:    true,
		Release:          server.Version,
		AttachStacktrace: true,
		Environment:      server.Environment,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	defer func() {
		err := recover()

		if err != nil {
			sentry.CurrentHub().Recover(err)
			sentry.Flush(time.Second * 5)
			if server.Environment == "dev" {
				panic(err)
			}
		}
	}()

	if err := serverCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)

	}
}
