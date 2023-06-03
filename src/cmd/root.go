package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "takeout",
		Short: "A google takeout command helper toolset",
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
