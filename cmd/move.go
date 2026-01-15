package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/israelmalagutti/git-wrapper/internal/config"
	"github.com/israelmalagutti/git-wrapper/internal/git"
	"github.com/israelmalagutti/git-wrapper/internal/stack"
	"github.com/spf13/cobra"
)

var (
	moveOnto string
)

var moveCmd = &cobra.Command{
	Use:   "move [target]",
	Short: "Rebase current branch onto a different parent",
	Long: `Move the current branch to a different parent by rebasing it onto the target branch.
Automatically restacks all descendants.

If no target is specified, opens an interactive selector to choose the new parent.

Example:
  gw move feat-base        # Move current branch onto feat-base
  gw move                  # Interactive selection
  gw move -o feat-base     # Using --onto flag`,
	Aliases: []string{"mv"},
	RunE:    runMove,
}

func init() {
	moveCmd.Flags().StringVarP(&moveOnto, "onto", "o", "", "Branch to move onto")
	rootCmd.AddCommand(moveCmd)
}

func runMove(cmd *cobra.Command, args []string) error {
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

	// Check if current branch is tracked
	if !metadata.IsTracked(currentBranch) {
		return fmt.Errorf("current branch '%s' is not tracked by gw", currentBranch)
	}

	// Don't move trunk
	if currentBranch == cfg.Trunk {
		return fmt.Errorf("cannot move trunk branch '%s'", cfg.Trunk)
	}

	// Determine target branch
	targetBranch := moveOnto
	if len(args) > 0 {
		targetBranch = args[0]
	}

	// If no target specified, show interactive selector
	if targetBranch == "" {
		s, err := stack.BuildStack(repo, cfg, metadata)
		if err != nil {
			return fmt.Errorf("failed to build stack: %w", err)
		}

		// Get all branches except current
		options := []string{}
		optionsMap := make(map[string]string)

		// Add trunk
		options = append(options, cfg.Trunk)
		optionsMap[cfg.Trunk] = cfg.Trunk

		// Add all tracked branches except current
		for _, node := range s.Nodes {
			if node.Name != currentBranch {
				// Get context for display
				parent := metadata.Branches[node.Name].Parent
				context := fmt.Sprintf("%s (parent: %s)", node.Name, parent)
				options = append(options, context)
				optionsMap[context] = node.Name
			}
		}

		if len(options) == 0 {
			return fmt.Errorf("no other branches available to move onto")
		}

		prompt := &survey.Select{
			Message: "Select target branch to move onto:",
			Options: options,
		}

		var selected string
		if err := survey.AskOne(prompt, &selected); err != nil {
			return fmt.Errorf("selection cancelled: %w", err)
		}

		// Map back to branch name
		if mapped, ok := optionsMap[selected]; ok {
			targetBranch = mapped
		} else {
			targetBranch = selected
		}
	}

	// Validate target branch
	if targetBranch == currentBranch {
		return fmt.Errorf("cannot move branch onto itself")
	}

	if !repo.BranchExists(targetBranch) {
		return fmt.Errorf("target branch '%s' does not exist", targetBranch)
	}

	// Check if target is a descendant of current branch
	s, err := stack.BuildStack(repo, cfg, metadata)
	if err != nil {
		return fmt.Errorf("failed to build stack: %w", err)
	}

	currentNode := s.GetNode(currentBranch)
	if currentNode == nil {
		return fmt.Errorf("current branch not found in stack")
	}

	if isDescendant(currentNode, targetBranch) {
		return fmt.Errorf("cannot move branch onto its descendant '%s'", targetBranch)
	}

	// Update parent in metadata
	oldParent := metadata.Branches[currentBranch].Parent

	fmt.Printf("Moving '%s' from '%s' to '%s'...\n", currentBranch, oldParent, targetBranch)

	// Update metadata with new parent
	if err := metadata.UpdateParent(currentBranch, targetBranch); err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	// Save metadata
	if err := metadata.Save(repo.GetMetadataPath()); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	// Rebase onto new parent
	fmt.Printf("Rebasing onto '%s'...\n", targetBranch)
	if _, err := repo.RunGitCommand("rebase", targetBranch); err != nil {
		// Rebase failed, restore old parent
		metadata.UpdateParent(currentBranch, oldParent)
		metadata.Save(repo.GetMetadataPath())

		return fmt.Errorf("rebase failed: %w\nMetadata restored to original state", err)
	}

	fmt.Println("✓ Rebased successfully")

	// Rebuild stack with new structure
	s, err = stack.BuildStack(repo, cfg, metadata)
	if err != nil {
		return fmt.Errorf("failed to rebuild stack: %w", err)
	}

	currentNode = s.GetNode(currentBranch)
	if currentNode == nil {
		return fmt.Errorf("current branch not found in stack")
	}

	// Restack children if any
	if len(currentNode.Children) > 0 {
		fmt.Println("\nRestacking children...")
		if err := restackChildren(repo, s, currentNode); err != nil {
			return fmt.Errorf("failed to restack children: %w", err)
		}
		fmt.Println("✓ Children restacked")
	}

	fmt.Printf("\n✓ Moved '%s' onto '%s'\n", currentBranch, targetBranch)

	return nil
}

// isDescendant checks if targetBranch is a descendant of node
func isDescendant(node *stack.Node, targetBranch string) bool {
	for _, child := range node.Children {
		if child.Name == targetBranch {
			return true
		}
		if isDescendant(child, targetBranch) {
			return true
		}
	}
	return false
}
