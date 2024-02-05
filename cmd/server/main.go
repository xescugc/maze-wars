package main

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/server"
)

var (
	logFile = path.Join(xdg.CacheHome, "maze-wars", "server.log")

	serverCmd = &cobra.Command{
		Use: "server",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := flux.NewDispatcher()
			out := os.Stdout
			lvl := slog.LevelInfo
			if viper.GetBool("verbose") {
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
			ss := server.NewStore(d, l)
			ad := server.NewActionDispatcher(d, l, ss)

			err := server.New(ad, ss, server.Options{
				Port:    viper.GetString("port"),
				Verbose: viper.GetBool("verbose"),
			})
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

	serverCmd.Flags().Bool("verbose", false, fmt.Sprintf("If all the logs are gonna be printed to %s", logFile))
	viper.BindPFlag("verbose", serverCmd.Flags().Lookup("verbose"))
}

func main() {
	if err := serverCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)

	}
}
