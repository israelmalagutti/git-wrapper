package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/israelmalagutti/git-wrapper/internal/colors"
	"github.com/israelmalagutti/git-wrapper/internal/config"
	"github.com/israelmalagutti/git-wrapper/internal/git"
	"github.com/spf13/cobra"
)

var (
	createMessage string
	createAll     bool
	createPatch   bool
)

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new branch stacked on the current branch",
	Long: `Create a new branch stacked on top of the current branch.

The new branch will be automatically tracked with the current branch as its parent.

Examples:
  gw create feat-auth                    # Create empty branch
  gw create feat-auth -m "Add login"     # Create and commit staged changes
  gw create feat-auth -am "Add login"    # Stage all changes and commit
  gw create feat-auth -pm "Add login"    # Interactive patch mode
  gw create -m "Add login"               # Auto-generate branch name from message`,
	Aliases: []string{"c"},
	RunE:    runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&createMessage, "message", "m", "", "Commit staged changes with this message")
	createCmd.Flags().BoolVarP(&createAll, "all", "a", false, "Stage all unstaged changes before committing")
	createCmd.Flags().BoolVarP(&createPatch, "patch", "p", false, "Interactively select hunks to stage")
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

	// Determine branch name
	branchName := resolveBranchName(args, createMessage)
	if branchName == "" {
		// Prompt for branch name if not provided
		prompt := &survey.Input{
			Message: "Branch name:",
		}
		err = survey.AskOne(prompt, &branchName, survey.WithValidator(survey.Required))
		if err != nil {
			if errors.Is(err, terminal.InterruptErr) {
				fmt.Println(colors.Muted("Cancelled."))
				return nil
			}
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

	// Create and checkout the new branch
	if err := repo.CreateBranch(branchName); err != nil {
		return err
	}

	if err := repo.CheckoutBranch(branchName); err != nil {
		return err
	}

	// Track the branch in metadata
	metadata, err := config.LoadMetadata(repo.GetMetadataPath())
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	metadata.TrackBranch(branchName, currentBranch)

	if err := metadata.Save(repo.GetMetadataPath()); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	colors.PrintCreated(branchName, currentBranch)

	// Stage all changes if --all flag is set
	if createAll {
		if _, err := repo.RunGitCommand("add", "-A"); err != nil {
			return fmt.Errorf("failed to stage changes: %w", err)
		}
	}

	// Check for changes
	hasStaged := detectStagedChanges(repo)
	hasUnstaged := detectUnstagedChanges(repo)

	// Handle commit if we have a message
	if createMessage != "" {
		if !hasStaged && !createAll {
			// No staged changes and not using --all
			if hasUnstaged {
				fmt.Println(colors.Muted("tip: ") + "There are unstaged changes. Use " + colors.Info("-a") + " to stage all changes.")
			}
			fmt.Println(colors.Muted("No staged changes; created branch with no commit."))
			return nil
		}

		// Commit the changes
		if err := commitChanges(repo, createMessage, createPatch && !createAll); err != nil {
			// Rollback: delete the branch on commit failure
			_ = repo.CheckoutBranch(currentBranch)
			_ = repo.DeleteBranch(branchName, true)
			metadata.UntrackBranch(branchName)
			_ = metadata.Save(repo.GetMetadataPath())
			return fmt.Errorf("failed to commit: %w", err)
		}

		fmt.Println(colors.Success("✓") + " Committed changes")
	} else {
		// No message provided - just created the branch
		if hasStaged {
			fmt.Println(colors.Muted("tip: ") + "You have staged changes. Use " + colors.Info("-m \"message\"") + " to commit them.")
		} else if hasUnstaged {
			fmt.Println(colors.Muted("tip: ") + "You have unstaged changes. Use " + colors.Info("-am \"message\"") + " to stage and commit.")
		} else {
			fmt.Println()
			fmt.Println(colors.Muted("Next steps:"))
			fmt.Println(colors.Muted("  Make your changes, then:"))
			fmt.Println(colors.Muted("  - gw modify -m \"message\"  to commit"))
			fmt.Println(colors.Muted("  - gw log                   to see your stack"))
		}
	}

	return nil
}

// resolveBranchName determines the branch name from args or message
func resolveBranchName(args []string, message string) string {
	if len(args) > 0 {
		return sanitizeBranchName(args[0])
	}

	// Auto-generate from commit message if provided
	if message != "" {
		return generateBranchName(message)
	}

	return ""
}

// sanitizeBranchName cleans up a branch name
func sanitizeBranchName(name string) string {
	// Replace spaces and special chars with dashes
	replacer := strings.NewReplacer(
		" ", "-",
		":", "-",
		"~", "-",
		"^", "-",
		"?", "-",
		"*", "-",
		"[", "-",
		"]", "-",
		"\\", "-",
	)
	name = replacer.Replace(name)

	// Remove leading/trailing dashes and dots
	name = strings.Trim(name, "-.")

	// Collapse multiple dashes
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}

	return strings.ToLower(name)
}

// generateBranchName creates a branch name from a commit message
func generateBranchName(message string) string {
	// Take first line only
	if idx := strings.Index(message, "\n"); idx != -1 {
		message = message[:idx]
	}

	// Truncate long messages
	if len(message) > 50 {
		message = message[:50]
	}

	return sanitizeBranchName(message)
}

// detectStagedChanges checks if there are staged changes
func detectStagedChanges(repo *git.Repo) bool {
	output, err := repo.RunGitCommand("diff", "--cached", "--shortstat")
	if err != nil {
		return false
	}
	return strings.TrimSpace(output) != ""
}

// hasStagedChanges is an alias for detectStagedChanges (used by other commands)
func hasStagedChanges(repo *git.Repo) (bool, error) {
	return detectStagedChanges(repo), nil
}

// detectUnstagedChanges checks if there are unstaged changes (modified or untracked)
func detectUnstagedChanges(repo *git.Repo) bool {
	// Check for modified tracked files
	output, err := repo.RunGitCommand("diff", "--shortstat")
	if err == nil && strings.TrimSpace(output) != "" {
		return true
	}

	// Check for untracked files
	output, err = repo.RunGitCommand("ls-files", "--others", "--exclude-standard")
	if err == nil && strings.TrimSpace(output) != "" {
		return true
	}

	return false
}

// commitChanges commits with the given message, optionally using patch mode
func commitChanges(repo *git.Repo, message string, patch bool) error {
	args := []string{"commit"}

	if patch {
		args = append(args, "-p")
	}

	args = append(args, "-m", message)

	_, err := repo.RunGitCommand(args...)
	return err
}
