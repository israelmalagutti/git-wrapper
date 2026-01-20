package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gw",
	Short: "gw - blazing fast git stack management",
	Long: `gw (git-wrapper) is a fast, simple git stack management tool.

It helps you work with stacked diffs (stacked PRs) efficiently,
maintaining parent-child relationships between branches.`,
	Version:       "0.1.0",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
	// rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug output")
}
