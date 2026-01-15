package cmd

import (
	"fmt"

	"github.com/israelmalagutti/git-wrapper/internal/config"
	"github.com/israelmalagutti/git-wrapper/internal/git"
	"github.com/spf13/cobra"
)

var (
	syncForce bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync metadata with git branches",
	Long: `Sync the gw metadata with actual git branches.

This command:
- Removes metadata for branches that no longer exist
- Validates stack structure (detects cycles)
- Ensures trunk branch has no parent
- Fixes any inconsistencies in the stack

Example:
  gw sync         # Clean up stale metadata
  gw sync -f      # Force sync without prompting`,
	RunE: runSync,
}

func init() {
	syncCmd.Flags().BoolVarP(&syncForce, "force", "f", false, "Don't prompt for confirmation")
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	// Initialize repository
	repo, err := git.NewRepo()
	if err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	// Load config
	cfg, err := config.Load(repo.GetConfigPath())
	if err != nil {
		return err
	}

	// Load metadata
	metadata, err := config.LoadMetadata(repo.GetMetadataPath())
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	var staleBranches []string
	var modified bool

	// Find branches in metadata that don't exist in git
	for branch := range metadata.Branches {
		if !repo.BranchExists(branch) {
			staleBranches = append(staleBranches, branch)
		}
	}

	if len(staleBranches) > 0 {
		fmt.Printf("Found %d stale branch(es) in metadata:\n", len(staleBranches))
		for _, branch := range staleBranches {
			fmt.Printf("  - %s (deleted from git)\n", branch)
		}

		if !syncForce {
			fmt.Print("\nRemove these branches from metadata? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Sync cancelled.")
				return nil
			}
		}

		// Remove stale branches
		for _, branch := range staleBranches {
			metadata.UntrackBranch(branch)
			modified = true
		}

		fmt.Printf("✓ Removed %d stale branch(es) from metadata\n", len(staleBranches))
	}

	// Validate trunk branch
	if !repo.BranchExists(cfg.Trunk) {
		return fmt.Errorf("trunk branch '%s' does not exist", cfg.Trunk)
	}

	// Ensure trunk has no parent in metadata
	if trunkMeta, exists := metadata.Branches[cfg.Trunk]; exists && trunkMeta.Parent != "" {
		fmt.Printf("⚠ Warning: trunk branch '%s' has parent '%s' in metadata\n", cfg.Trunk, trunkMeta.Parent)
		if !syncForce {
			fmt.Print("Fix this by removing trunk's parent? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Sync cancelled.")
				return nil
			}
		}
		metadata.UntrackBranch(cfg.Trunk)
		modified = true
		fmt.Printf("✓ Fixed trunk branch '%s' (removed invalid parent)\n", cfg.Trunk)
	}

	// Detect cycles in stack
	if err := detectCycles(metadata); err != nil {
		return fmt.Errorf("cycle detected in stack: %w", err)
	}

	// Save metadata if modified
	if modified {
		if err := metadata.Save(repo.GetMetadataPath()); err != nil {
			return fmt.Errorf("failed to save metadata: %w", err)
		}
	}

	if !modified && len(staleBranches) == 0 {
		fmt.Println("✓ Everything is in sync")
	}

	return nil
}

// detectCycles detects cycles in the branch parent-child relationships
func detectCycles(metadata *config.Metadata) error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var detectCycle func(branch string) error
	detectCycle = func(branch string) error {
		visited[branch] = true
		recStack[branch] = true

		meta, exists := metadata.Branches[branch]
		if !exists || meta.Parent == "" {
			recStack[branch] = false
			return nil
		}

		if recStack[meta.Parent] {
			return fmt.Errorf("cycle detected: %s -> %s", branch, meta.Parent)
		}

		if !visited[meta.Parent] {
			if err := detectCycle(meta.Parent); err != nil {
				return err
			}
		}

		recStack[branch] = false
		return nil
	}

	for branch := range metadata.Branches {
		if !visited[branch] {
			if err := detectCycle(branch); err != nil {
				return err
			}
		}
	}

	return nil
}
