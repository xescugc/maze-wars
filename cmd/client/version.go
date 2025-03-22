package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xescugc/maze-wars/client"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Prints the current build version",
		Run: func(cmd *cobra.Command, args []string) {
			if client.Version != "" {
				fmt.Printf("The current version is: %s\n", client.Version)
			} else {
				fmt.Printf("No version defined\n")
			}
		},
	}
)
