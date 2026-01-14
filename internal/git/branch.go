package git

import (
	"fmt"
	"strings"
)

// GetCurrentBranch returns the name of the current branch
func (r *Repo) GetCurrentBranch() (string, error) {
	output, err := r.RunGitCommand("branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(output)
	if branch == "" {
		return "", fmt.Errorf("not on any branch (detached HEAD)")
	}

	return branch, nil
}

// ListBranches returns a list of all local branches
func (r *Repo) ListBranches() ([]string, error) {
	output, err := r.RunGitCommand("branch", "--format=%(refname:short)")
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	if output == "" {
		return []string{}, nil
	}

	branches := strings.Split(output, "\n")
	return branches, nil
}

// BranchExists checks if a branch exists
func (r *Repo) BranchExists(branch string) bool {
	_, err := r.RunGitCommand("rev-parse", "--verify", branch)
	return err == nil
}

// CreateBranch creates a new branch from the current HEAD
func (r *Repo) CreateBranch(name string) error {
	_, err := r.RunGitCommand("branch", name)
	if err != nil {
		return fmt.Errorf("failed to create branch %s: %w", name, err)
	}
	return nil
}

// CheckoutBranch switches to the specified branch
func (r *Repo) CheckoutBranch(branch string) error {
	_, err := r.RunGitCommand("checkout", branch)
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w", branch, err)
	}
	return nil
}

// DeleteBranch deletes a branch (force delete if merged is false)
func (r *Repo) DeleteBranch(branch string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}

	_, err := r.RunGitCommand("branch", flag, branch)
	if err != nil {
		return fmt.Errorf("failed to delete branch %s: %w", branch, err)
	}
	return nil
}

// GetBranchCommit returns the commit SHA of a branch
func (r *Repo) GetBranchCommit(branch string) (string, error) {
	output, err := r.RunGitCommand("rev-parse", branch)
	if err != nil {
		return "", fmt.Errorf("failed to get commit for branch %s: %w", branch, err)
	}
	return output, nil
}
