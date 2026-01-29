package cmd

import (
	"testing"
)

func TestSplitByHunkModeCherryPickConflict(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-split-conflict", "main")
	repo.commitFile(t, "conflict.txt", "child", "child commit")

	if err := repo.repo.CheckoutBranch("main"); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}
	repo.commitFile(t, "conflict.txt", "main", "main commit")

	if err := splitByHunkMode(repo.repo, repo.cfg, repo.metadata, "feat-split-conflict", "main", "feat-split-base"); err == nil {
		t.Fatalf("expected splitByHunkMode cherry-pick conflict error")
	}
}

func TestSplitByFileModeCherryPickConflict(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-split-conflict", "main")
	repo.commitFile(t, "conflict.txt", "child", "child commit")

	if err := repo.repo.CheckoutBranch("main"); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}
	repo.commitFile(t, "conflict.txt", "main", "main commit")

	if err := splitByFileMode(repo.repo, repo.cfg, repo.metadata, "feat-split-conflict", "main", "feat-split-base", []string{"conflict.txt"}); err == nil {
		t.Fatalf("expected splitByFileMode cherry-pick conflict error")
	}
}

func TestSplitByHunkModeCommitHookFailure(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-split-hook", "main")
	repo.commitFile(t, "hunk.txt", "data", "hunk commit")

	writeFailingHook(t, repo.dir, "pre-commit")
	t.Setenv("GW_TEST_AUTO_STAGE", "1")

	if err := splitByHunkMode(repo.repo, repo.cfg, repo.metadata, "feat-split-hook", "main", "feat-split-base"); err == nil {
		t.Fatalf("expected splitByHunkMode commit hook error")
	}
}

func TestSplitByFileModeCommitHookFailure(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-split-hook", "main")
	repo.commitFile(t, "move.txt", "data", "move commit")

	writeFailingHook(t, repo.dir, "pre-commit")

	if err := splitByFileMode(repo.repo, repo.cfg, repo.metadata, "feat-split-hook", "main", "feat-split-base", []string{"move.txt"}); err == nil {
		t.Fatalf("expected splitByFileMode commit hook error")
	}
}
