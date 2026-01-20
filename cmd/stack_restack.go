package cmd

import (
	"fmt"

	"github.com/israelmalagutti/git-wrapper/internal/config"
	"github.com/israelmalagutti/git-wrapper/internal/git"
	"github.com/israelmalagutti/git-wrapper/internal/stack"
	"github.com/spf13/cobra"
)

var stackRestackCmd = &cobra.Command{
	Use:     "restack",
	Aliases: []string{"r", "fix", "f"},
	Short:   "Rebase stack to maintain parent-child relationships",
	Long: `Ensure each branch in the current stack is based on its parent, rebasing if necessary.

This command:
- Checks if the current branch needs rebasing onto its parent
- Performs the rebase if needed
- Recursively restacks all children branches
- Handles conflicts interactively

Example:
  gw stack restack    # Restack current branch and children
  gw stack r          # Short alias
  gw stack fix        # Alternative alias`,
	RunE: runStackRestack,
}

func init() {
	stackCmd.AddCommand(stackRestackCmd)
}

func runStackRestack(cmd *cobra.Command, args []string) error {
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

	// Get current branch
	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Build stack
	s, err := stack.BuildStack(repo, cfg, metadata)
	if err != nil {
		return fmt.Errorf("failed to build stack: %w", err)
	}

	// Handle trunk specially - restack all children of trunk
	if currentBranch == cfg.Trunk {
		trunkNode := s.GetNode(cfg.Trunk)
		if trunkNode == nil {
			return fmt.Errorf("trunk '%s' not found in stack", cfg.Trunk)
		}

		if len(trunkNode.Children) == 0 {
			fmt.Println("No branches to restack from trunk.")
			return nil
		}

		if err := restackChildren(repo, s, trunkNode); err != nil {
			return err
		}

		// Return to trunk
		if err := repo.CheckoutBranch(cfg.Trunk); err != nil {
			fmt.Printf("Warning: could not return to trunk: %v\n", err)
		}

		return nil
	}

	// Check if current branch is tracked
	if !metadata.IsTracked(currentBranch) {
		return fmt.Errorf("current branch '%s' is not tracked by gw", currentBranch)
	}

	// Get current branch node
	node := s.GetNode(currentBranch)
	if node == nil {
		return fmt.Errorf("branch '%s' not found in stack", currentBranch)
	}

	if node.Parent == nil {
		return fmt.Errorf("branch '%s' has no parent", currentBranch)
	}

	// Restack current branch
	if err := restackBranch(repo, currentBranch, node.Parent.Name); err != nil {
		return err
	}

	// Recursively restack children
	if len(node.Children) > 0 {
		if err := restackChildren(repo, s, node); err != nil {
			return err
		}
	}

	return nil
}

// restackBranch rebases a branch onto its parent
func restackBranch(repo *git.Repo, branch, parent string) error {
	// Check if branch needs rebasing
	needsRebase, err := needsRebase(repo, branch, parent)
	if err != nil {
		return err
	}

	if !needsRebase {
		fmt.Printf("%s does not need to be restacked on %s.\n", branch, parent)
		return nil
	}

	// Perform rebase
	_, err = repo.RunGitCommand("rebase", parent, branch)
	if err != nil {
		fmt.Printf("\nHit conflict restacking %s on %s.\n", branch, parent)
		fmt.Println("\nTo fix and continue:")
		fmt.Println("  (1) resolve the merge conflicts")
		fmt.Println("  (2) stage changes with: git add .")
		fmt.Println("  (3) continue rebase: git rebase --continue")
		fmt.Println("  (4) restack remaining: gw stack restack")
		fmt.Println("\nTo abort: git rebase --abort")
		return fmt.Errorf("rebase conflict")
	}

	fmt.Printf("Restacked %s on %s.\n", branch, parent)
	return nil
}

// restackChildren recursively restacks all children of a node
func restackChildren(repo *git.Repo, s *stack.Stack, parent *stack.Node) error {
	for _, child := range parent.Children {
		// Checkout child branch
		if err := repo.CheckoutBranch(child.Name); err != nil {
			return fmt.Errorf("failed to checkout '%s': %w", child.Name, err)
		}

		// Restack this child
		if err := restackBranch(repo, child.Name, parent.Name); err != nil {
			return err
		}

		// Recursively restack its children
		if len(child.Children) > 0 {
			if err := restackChildren(repo, s, child); err != nil {
				return err
			}
		}
	}

	return nil
}

// needsRebase checks if a branch needs to be rebased onto its parent
func needsRebase(repo *git.Repo, branch, parent string) (bool, error) {
	// Get merge base between branch and parent
	mergeBase, err := repo.RunGitCommand("merge-base", branch, parent)
	if err != nil {
		return false, fmt.Errorf("failed to get merge base: %w", err)
	}

	// Get parent's current commit
	parentCommit, err := repo.RunGitCommand("rev-parse", parent)
	if err != nil {
		return false, fmt.Errorf("failed to get parent commit: %w", err)
	}

	// If merge base != parent commit, branch needs rebasing
	return mergeBase != parentCommit, nil
}
