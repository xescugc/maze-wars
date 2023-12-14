package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/server"
)

var (
	serverCmd = &cobra.Command{
		Use: "server",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := flux.NewDispatcher()
			ad := server.NewActionDispatcher(d)
			rooms := server.NewRoomsStore(d)

			err := server.New(ad, rooms, server.Options{
				Port: viper.GetString("port"),
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
}

func main() {
	if err := serverCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)

	}
}
