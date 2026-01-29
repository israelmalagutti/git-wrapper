package cmd

import (
	"os"
	"testing"
)

func TestRunContinueRestacksChildren(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	t.Setenv("GIT_EDITOR", "true")
	t.Setenv("GIT_SEQUENCE_EDITOR", "true")

	repo.createBranch(t, "feat-parent", "main")
	repo.commitFile(t, "parent.txt", "parent", "parent commit")
	repo.createBranch(t, "feat-child", "feat-parent")
	repo.commitFile(t, "child.txt", "child", "child commit")

	if err := repo.repo.CheckoutBranch("main"); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}
	if err := os.WriteFile(repo.dir+"/conflict.txt", []byte("main"), 0644); err != nil {
		t.Fatalf("failed to write main conflict: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "conflict.txt"); err != nil {
		t.Fatalf("failed to add main conflict: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("commit", "-m", "main change"); err != nil {
		t.Fatalf("failed to commit main change: %v", err)
	}

	if err := repo.repo.CheckoutBranch("feat-parent"); err != nil {
		t.Fatalf("failed to checkout parent: %v", err)
	}
	if err := os.WriteFile(repo.dir+"/conflict.txt", []byte("parent"), 0644); err != nil {
		t.Fatalf("failed to write parent conflict: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "conflict.txt"); err != nil {
		t.Fatalf("failed to add parent conflict: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("commit", "-m", "parent change"); err != nil {
		t.Fatalf("failed to commit parent change: %v", err)
	}

	if _, err := repo.repo.RunGitCommand("rebase", "main"); err == nil {
		t.Fatalf("expected rebase conflict")
	}

	if err := os.WriteFile(repo.dir+"/conflict.txt", []byte("resolved"), 0644); err != nil {
		t.Fatalf("failed to resolve conflict: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "conflict.txt"); err != nil {
		t.Fatalf("failed to add resolved conflict: %v", err)
	}

	if err := runContinue(nil, nil); err != nil {
		t.Fatalf("runContinue restack children failed: %v", err)
	}
}
