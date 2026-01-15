package cmd

import (
	"errors"
	"fmt"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/israelmalagutti/git-wrapper/internal/config"
	"github.com/israelmalagutti/git-wrapper/internal/git"
	"github.com/israelmalagutti/git-wrapper/internal/stack"
	"github.com/spf13/cobra"
)

var (
	checkoutTrunk        bool
	checkoutShowUntracked bool
	checkoutStack        bool
)

var checkoutCmd = &cobra.Command{
	Use:     "checkout [branch]",
	Aliases: []string{"co"},
	Short:   "Switch to a branch",
	Long: `Switch to a branch in your stack.

If no branch is specified, opens an interactive selector showing
all branches with their stack context.

Example:
  gw checkout feat-1       # Switch to feat-1
  gw co feat-2             # Switch to feat-2 (alias)
  gw checkout              # Interactive branch selector (tracked only)
  gw checkout -t           # Switch to trunk
  gw checkout -s           # Interactive selector (current stack only)
  gw co -u                 # Show untracked branches in selector`,
	RunE: runCheckout,
}

func init() {
	rootCmd.AddCommand(checkoutCmd)
	checkoutCmd.Flags().BoolVarP(&checkoutTrunk, "trunk", "t", false, "Checkout the trunk branch")
	checkoutCmd.Flags().BoolVarP(&checkoutShowUntracked, "show-untracked", "u", false, "Include untracked branches in interactive selection")
	checkoutCmd.Flags().BoolVarP(&checkoutStack, "stack", "s", false, "Only show current stack in interactive selection")
}

func runCheckout(cmd *cobra.Command, args []string) error {
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

	// Build stack
	s, err := stack.BuildStack(repo, cfg, metadata)
	if err != nil {
		return fmt.Errorf("failed to build stack: %w", err)
	}

	// Handle --trunk flag
	if checkoutTrunk {
		targetBranch := cfg.Trunk
		return checkoutBranch(repo, s, targetBranch)
	}

	// Determine which branch to checkout
	var targetBranch string
	if len(args) > 0 {
		targetBranch = args[0]
	} else {
		// Get all branches
		branches, err := repo.ListBranches()
		if err != nil {
			return fmt.Errorf("failed to list branches: %w", err)
		}

		if len(branches) == 0 {
			return fmt.Errorf("no branches found")
		}

		// Filter branches based on flags
		filteredBranches := []string{}
		currentBranch, _ := repo.GetCurrentBranch()

		for _, branch := range branches {
			// Handle --stack flag: only show current stack
			if checkoutStack {
				// Include if it's in the path from trunk to current, or is a descendant
				path := s.FindPath(currentBranch)
				inPath := false
				for _, node := range path {
					if node.Name == branch {
						inPath = true
						break
					}
				}

				// Also include descendants of current branch
				isDescendant := false
				if currentNode := s.GetNode(currentBranch); currentNode != nil {
					isDescendant = isInDescendants(s, currentNode, branch)
				}

				if !inPath && !isDescendant {
					continue
				}
			}

			// Handle --show-untracked flag
			if !checkoutShowUntracked && !metadata.IsTracked(branch) && branch != cfg.Trunk {
				continue
			}

			filteredBranches = append(filteredBranches, branch)
		}

		if len(filteredBranches) == 0 {
			return fmt.Errorf("no branches match the filter criteria")
		}

		branches = filteredBranches

		// Sort branches: trunk first, then tracked, then others
		sort.Slice(branches, func(i, j int) bool {
			bi := branches[i]
			bj := branches[j]

			// Trunk first
			if bi == cfg.Trunk {
				return true
			}
			if bj == cfg.Trunk {
				return false
			}

			// Tracked branches next
			iTracked := metadata.IsTracked(bi)
			jTracked := metadata.IsTracked(bj)
			if iTracked && !jTracked {
				return true
			}
			if !iTracked && jTracked {
				return false
			}

			// Alphabetical
			return bi < bj
		})

		// Create options with context
		options := make([]string, len(branches))
		for i, branch := range branches {
			options[i] = branch
		}

		// Interactive selector
		prompt := &survey.Select{
			Message: "Select branch to checkout:",
			Options: options,
			Description: func(value string, index int) string {
				if value == cfg.Trunk {
					return "(trunk)"
				}

				node := s.GetNode(value)
				if node == nil {
					return "(not tracked)"
				}

				// Show parent info
				if node.Parent != nil {
					return fmt.Sprintf("(parent: %s)", node.Parent.Name)
				}

				return ""
			},
		}

		err = survey.AskOne(prompt, &targetBranch, survey.WithValidator(survey.Required))
		if err != nil {
			// Handle ESC/Ctrl+C gracefully
			if errors.Is(err, terminal.InterruptErr) {
				fmt.Println("Cancelled.")
				return nil
			}
			return fmt.Errorf("failed to get branch selection: %w", err)
		}
	}

	return checkoutBranch(repo, s, targetBranch)
}

// checkoutBranch performs the actual checkout and displays stack context
func checkoutBranch(repo *git.Repo, s *stack.Stack, targetBranch string) error {
	// Verify branch exists
	if !repo.BranchExists(targetBranch) {
		return fmt.Errorf("branch '%s' does not exist", targetBranch)
	}

	// Get current branch to show where we're switching from
	currentBranch, err := repo.GetCurrentBranch()
	if err == nil && currentBranch == targetBranch {
		fmt.Printf("Already on branch '%s'\n", targetBranch)
		return nil
	}

	// Checkout the branch
	if err := repo.CheckoutBranch(targetBranch); err != nil {
		return err
	}

	fmt.Printf("Switched to branch '%s'\n", targetBranch)

	// Show stack context if tracked
	node := s.GetNode(targetBranch)
	if node != nil {
		if node.Parent != nil {
			fmt.Printf("  Parent: %s\n", node.Parent.Name)
		}
		children := s.GetChildren(targetBranch)
		if len(children) > 0 {
			fmt.Printf("  Children: ")
			for i, child := range children {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(child.Name)
			}
			fmt.Println()
		}
	} else {
		fmt.Println("This branch is not tracked by Git-Wrapper")
	}

	return nil
}

// isInDescendants checks if a branch is a descendant of a node
func isInDescendants(s *stack.Stack, node *stack.Node, branchName string) bool {
	for _, child := range node.Children {
		if child.Name == branchName {
			return true
		}
		if isInDescendants(s, child, branchName) {
			return true
		}
	}
	return false
}
