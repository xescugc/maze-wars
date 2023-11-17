package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/ltw/server"
)

var (
	serverPort string

	serverCmd = &cobra.Command{
		Use: "server",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := flux.NewDispatcher()
			ad := server.NewActionDispatcher(d)
			rooms := server.NewRoomsStore(d)

			err := server.New(ad, rooms, server.Options{
				Port: serverPort,
			})
			if err != nil {
				return fmt.Errorf("server error: %w", err)
			}

			return nil
		},
	}
)

func init() {
	serverCmd.Flags().StringVar(&serverPort, "port", ":5555", "The port in which the sever is open")
}
