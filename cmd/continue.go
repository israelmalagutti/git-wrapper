package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/israelmalagutti/git-wrapper/internal/colors"
	"github.com/israelmalagutti/git-wrapper/internal/config"
	"github.com/israelmalagutti/git-wrapper/internal/git"
	"github.com/israelmalagutti/git-wrapper/internal/stack"
	"github.com/spf13/cobra"
)

var continueCmd = &cobra.Command{
	Use:   "continue",
	Short: "Continue after resolving conflicts",
	Long: `Continue a rebase operation after resolving conflicts.

This command:
1. Continues the in-progress rebase (git rebase --continue)
2. Restacks remaining children branches

Use this after resolving merge conflicts during a restack operation.

Example:
  # After resolving conflicts:
  git add .
  gw continue`,
	RunE: runContinue,
}

func init() {
	rootCmd.AddCommand(continueCmd)
}

func runContinue(cmd *cobra.Command, args []string) error {
	// Initialize repository
	repo, err := git.NewRepo()
	if err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	// Check if a rebase is in progress
	if !isRebaseInProgress(repo) {
		fmt.Println(colors.Muted("No rebase in progress."))
		return nil
	}

	// Continue the rebase
	fmt.Println(colors.Muted("Continuing rebase..."))
	if _, err := repo.RunGitCommand("rebase", "--continue"); err != nil {
		return fmt.Errorf("rebase --continue failed: resolve conflicts and try again")
	}

	// Get current branch (the one we just finished rebasing)
	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	fmt.Printf("%s Rebased %s\n", colors.Success("✓"), colors.BranchCurrent(currentBranch))

	// Load config and metadata
	cfg, err := config.Load(repo.GetConfigPath())
	if err != nil {
		return err
	}

	metadata, err := config.LoadMetadata(repo.GetMetadataPath())
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	// Build stack
	s, err := stack.BuildStack(repo, cfg, metadata)
	if err != nil {
		return fmt.Errorf("failed to build stack: %w", err)
	}

	// Get current node
	node := s.GetNode(currentBranch)
	if node == nil {
		// Branch not tracked, nothing more to do
		return nil
	}

	// Restack children if any
	if len(node.Children) > 0 {
		fmt.Println()
		fmt.Println(colors.Muted("Restacking children..."))

		if err := continueRestackChildren(repo, s, node); err != nil {
			return err
		}

		// Return to original branch
		if err := repo.CheckoutBranch(currentBranch); err != nil {
			fmt.Printf("%s Could not return to %s: %v\n",
				colors.Warning("⚠"),
				colors.BranchCurrent(currentBranch),
				err)
		}
	}

	fmt.Println()
	fmt.Printf("%s All done!\n", colors.Success("✓"))

	return nil
}

// isRebaseInProgress checks if a rebase is currently in progress
func isRebaseInProgress(repo *git.Repo) bool {
	gitDir := repo.GetGitDir()

	// Check for rebase-merge directory (interactive rebase)
	if _, err := os.Stat(filepath.Join(gitDir, "rebase-merge")); err == nil {
		return true
	}

	// Check for rebase-apply directory (regular rebase)
	if _, err := os.Stat(filepath.Join(gitDir, "rebase-apply")); err == nil {
		return true
	}

	return false
}

// continueRestackChildren rebases children onto parent after a continue
func continueRestackChildren(repo *git.Repo, s *stack.Stack, parent *stack.Node) error {
	for _, child := range parent.Children {
		// Checkout child branch
		if err := repo.CheckoutBranch(child.Name); err != nil {
			return fmt.Errorf("failed to checkout '%s': %w", child.Name, err)
		}

		// Check if needs rebase
		needsRebase, err := childNeedsRebase(repo, child.Name, parent.Name)
		if err != nil {
			return err
		}

		if !needsRebase {
			fmt.Printf("%s %s already up to date\n",
				colors.Success("✓"),
				colors.BranchCurrent(child.Name))
		} else {
			// Perform rebase
			if _, err := repo.RunGitCommand("rebase", parent.Name, child.Name); err != nil {
				fmt.Println()
				fmt.Printf("%s Conflict restacking %s onto %s\n",
					colors.Warning("⚠"),
					colors.BranchCurrent(child.Name),
					colors.BranchParent(parent.Name))
				fmt.Println()
				fmt.Println(colors.Muted("To continue:"))
				fmt.Println(colors.Muted("  1. Resolve conflicts"))
				fmt.Println(colors.Muted("  2. git add ."))
				fmt.Println(colors.Muted("  3. gw continue"))
				fmt.Println()
				fmt.Println(colors.Muted("To abort: git rebase --abort"))
				return fmt.Errorf("rebase conflict")
			}

			fmt.Printf("%s Restacked %s onto %s\n",
				colors.Success("✓"),
				colors.BranchCurrent(child.Name),
				colors.BranchParent(parent.Name))
		}

		// Recursively restack grandchildren
		if len(child.Children) > 0 {
			if err := continueRestackChildren(repo, s, child); err != nil {
				return err
			}
		}
	}

	return nil
}

// childNeedsRebase checks if a child branch needs rebasing onto parent
func childNeedsRebase(repo *git.Repo, child, parent string) (bool, error) {
	mergeBase, err := repo.RunGitCommand("merge-base", child, parent)
	if err != nil {
		return false, fmt.Errorf("failed to get merge base: %w", err)
	}

	parentCommit, err := repo.RunGitCommand("rev-parse", parent)
	if err != nil {
		return false, fmt.Errorf("failed to get parent commit: %w", err)
	}

	return mergeBase != parentCommit, nil
}
