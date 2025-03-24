package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xescugc/maze-wars/server"
)

var (
	// version is the value of the current version, this
	// is set via -ldflags
	version string = "dev"

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Prints the current build version",
		Run: func(cmd *cobra.Command, args []string) {
			if server.Version != "" {
				fmt.Printf("The current version is: %s\n", server.Version)
			} else {
				fmt.Printf("No version defined\n")
			}
		},
	}
)
