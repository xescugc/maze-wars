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
			ss := &server.Store{}
			d := flux.NewDispatcher()
			ad := server.NewActionDispatcher(d, ss)
			rooms := server.NewRoomsStore(d, ss)
			users := server.NewUsersStore(d, ss)

			ss.Rooms = rooms
			ss.Users = users

			err := server.New(ad, ss, server.Options{
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
