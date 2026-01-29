package cmd

import (
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
)

func TestRunTopPaths(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	// Already at top (no children)
	if err := runTop(nil, nil); err == nil {
		t.Fatalf("expected already at top error")
	}

	// Multiple leaves prompt
	repo.createBranch(t, "feat-a", "main")
	repo.createBranch(t, "feat-b", "main")
	if err := repo.repo.CheckoutBranch("main"); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}

	withAskOne(t, []interface{}{"feat-a"}, func() {
		if err := runTop(nil, nil); err != nil {
			t.Fatalf("runTop prompt failed: %v", err)
		}
	})
}

func TestRunTopCancelled(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-a", "main")
	repo.createBranch(t, "feat-b", "main")
	if err := repo.repo.CheckoutBranch("main"); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}

	withAskOneError(t, terminal.InterruptErr, func() {
		if err := runTop(nil, nil); err != nil {
			t.Fatalf("runTop cancel failed: %v", err)
		}
	})
}

func TestRunTopUntrackedBranch(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := repo.repo.CreateBranch("untracked-top"); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}
	if err := repo.repo.CheckoutBranch("untracked-top"); err != nil {
		t.Fatalf("failed to checkout branch: %v", err)
	}

	if err := runTop(nil, nil); err == nil {
		t.Fatalf("expected untracked branch error")
	}
}
