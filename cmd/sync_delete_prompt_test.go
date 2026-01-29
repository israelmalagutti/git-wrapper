package cmd

import (
	"os"
	"testing"
)

func TestDeleteMergedBranchesPromptAll(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-merged-1", "main")
	if err := repo.repo.CheckoutBranch("feat-merged-1"); err != nil {
		t.Fatalf("failed to checkout feat-merged-1: %v", err)
	}
	repo.commitFile(t, "m1.txt", "data", "feat1 commit")
	if err := repo.repo.CheckoutBranch("main"); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("merge", "feat-merged-1"); err != nil {
		t.Fatalf("failed to merge feat-merged-1: %v", err)
	}

	repo.createBranch(t, "feat-merged-2", "main")
	if err := repo.repo.CheckoutBranch("feat-merged-2"); err != nil {
		t.Fatalf("failed to checkout feat-merged-2: %v", err)
	}
	repo.commitFile(t, "m2.txt", "data", "feat2 commit")
	if err := repo.repo.CheckoutBranch("main"); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("merge", "feat-merged-2"); err != nil {
		t.Fatalf("failed to merge feat-merged-2: %v", err)
	}

	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	if _, err := w.Write([]byte("a\n")); err != nil {
		t.Fatalf("failed to write pipe: %v", err)
	}
	w.Close()
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	if err := deleteMergedBranches(repo.repo, repo.metadata, "main", false); err != nil {
		t.Fatalf("deleteMergedBranches prompt failed: %v", err)
	}
}
