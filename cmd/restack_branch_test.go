package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRestackBranchSuccess(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-rebase", "main")
	if err := repo.repo.CheckoutBranch("main"); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo.dir, "main.txt"), []byte("main"), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "main.txt"); err != nil {
		t.Fatalf("failed to add main file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("commit", "-m", "main commit"); err != nil {
		t.Fatalf("failed to commit main: %v", err)
	}

	if err := restackBranch(repo.repo, "feat-rebase", "main"); err != nil {
		t.Fatalf("restackBranch failed: %v", err)
	}
}

func TestRestackBranchConflict(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-conflict", "main")
	if err := repo.repo.CheckoutBranch("feat-conflict"); err != nil {
		t.Fatalf("failed to checkout feat: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo.dir, "conflict.txt"), []byte("feat"), 0644); err != nil {
		t.Fatalf("failed to write feat file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "conflict.txt"); err != nil {
		t.Fatalf("failed to add feat file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("commit", "-m", "feat commit"); err != nil {
		t.Fatalf("failed to commit feat: %v", err)
	}

	if err := repo.repo.CheckoutBranch("main"); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo.dir, "conflict.txt"), []byte("main"), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "conflict.txt"); err != nil {
		t.Fatalf("failed to add main file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("commit", "-m", "main commit"); err != nil {
		t.Fatalf("failed to commit main: %v", err)
	}

	if err := restackBranch(repo.repo, "feat-conflict", "main"); err == nil {
		t.Fatalf("expected restack conflict error")
	}
	_, _ = repo.repo.RunGitCommand("rebase", "--abort")
}
