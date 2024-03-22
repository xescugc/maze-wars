package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// version is the value of the current version, this
	// is set via -ldflags
	version string

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Prints the current build version",
		Run: func(cmd *cobra.Command, args []string) {
			if version != "" {
				fmt.Printf("The current version is: %s\n", version)
			} else {
				fmt.Printf("No version defined\n")
			}
		},
	}
)
