package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/israelmalagutti/git-wrapper/internal/config"
	"github.com/israelmalagutti/git-wrapper/internal/git"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new branch stacked on the current branch",
	Long: `Create a new branch stacked on top of the current branch.

The new branch will be automatically tracked with the current branch as its parent.
If you have staged changes, you'll be prompted to commit them to the new branch.

Example:
  gw create feat-auth       # Create feat-auth stacked on current branch
  gw create                 # Prompt for branch name`,
	RunE: runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Initialize repository
	repo, err := git.NewRepo()
	if err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	// Check if gw is initialized
	_, err = config.Load(repo.GetConfigPath())
	if err != nil {
		return err
	}

	// Get current branch (parent of new branch)
	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Get branch name
	var branchName string
	if len(args) > 0 {
		branchName = args[0]
	} else {
		// Prompt for branch name
		prompt := &survey.Input{
			Message: "Branch name:",
		}
		if err := survey.AskOne(prompt, &branchName, survey.WithValidator(survey.Required)); err != nil {
			return fmt.Errorf("failed to get branch name: %w", err)
		}
	}

	// Validate branch name
	if branchName == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// Check if branch already exists
	if repo.BranchExists(branchName) {
		return fmt.Errorf("branch '%s' already exists", branchName)
	}

	// Check for staged changes
	hasStagedChanges, err := hasStagedChanges(repo)
	if err != nil {
		return fmt.Errorf("failed to check for staged changes: %w", err)
	}

	// Create the branch
	if err := repo.CreateBranch(branchName); err != nil {
		return err
	}

	// Checkout the new branch
	if err := repo.CheckoutBranch(branchName); err != nil {
		return err
	}

	// Load metadata
	metadata, err := config.LoadMetadata(repo.GetMetadataPath())
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	// Track the branch
	metadata.TrackBranch(branchName, currentBranch)

	// Save metadata
	if err := metadata.Save(repo.GetMetadataPath()); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	fmt.Printf("✓ Created branch '%s' stacked on '%s'\n", branchName, currentBranch)
	fmt.Printf("✓ Switched to branch '%s'\n", branchName)

	// If there are staged changes, prompt to commit
	if hasStagedChanges {
		var commitNow bool
		prompt := &survey.Confirm{
			Message: "You have staged changes. Commit them now?",
			Default: true,
		}
		if err := survey.AskOne(prompt, &commitNow); err != nil {
			return fmt.Errorf("failed to get commit confirmation: %w", err)
		}

		if commitNow {
			// Prompt for commit message
			var commitMsg string
			msgPrompt := &survey.Input{
				Message: "Commit message:",
			}
			if err := survey.AskOne(msgPrompt, &commitMsg, survey.WithValidator(survey.Required)); err != nil {
				return fmt.Errorf("failed to get commit message: %w", err)
			}

			// Commit the changes
			if _, err := repo.RunGitCommand("commit", "-m", commitMsg); err != nil {
				return fmt.Errorf("failed to commit changes: %w", err)
			}

			fmt.Printf("✓ Committed staged changes\n")
		}
	}

	// Show next steps
	if !hasStagedChanges {
		fmt.Println("\nNext steps:")
		fmt.Println("  - Make your changes")
		fmt.Println("  - Use 'gw modify' to amend or create commits")
		fmt.Println("  - Use 'gw log' to see your stack")
	}

	return nil
}

// hasStagedChanges checks if there are any staged changes
func hasStagedChanges(repo *git.Repo) (bool, error) {
	// Check if index has changes
	output, err := repo.RunGitCommand("diff", "--cached", "--quiet")
	if err != nil {
		// Exit code 1 means there are differences (staged changes)
		// This is expected and means we have staged changes
		return true, nil
	}
	// Exit code 0 means no differences
	return output != "", nil
}
