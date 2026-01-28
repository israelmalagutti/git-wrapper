package cmd

import (
	"github.com/spf13/cobra"
)

var restackCmd = &cobra.Command{
	Use:     "restack",
	Aliases: []string{"rs"},
	Short:   "Restack current branch and its children",
	Long: `Rebase the current branch onto its parent and recursively restack all children.

This is an alias for 'gw stack restack'.

This command:
- Checks if the current branch needs rebasing onto its parent
- Performs the rebase if needed
- Recursively restacks all children branches
- Handles conflicts interactively

Example:
  gw restack    # Restack current branch and children
  gw rs         # Short alias`,
	RunE: runStackRestack,
}

func init() {
	rootCmd.AddCommand(restackCmd)
}
