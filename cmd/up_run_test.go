package cmd

import (
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
)

func TestRunUpErrorsAndPrompt(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	// Invalid step
	if err := runUp(nil, []string{"0"}); err == nil {
		t.Fatalf("expected invalid step error")
	}

	// Already at top (no children)
	if err := runUp(nil, nil); err == nil {
		t.Fatalf("expected already at top error")
	}

	// Multiple children prompt
	repo.createBranch(t, "feat-a", "main")
	repo.createBranch(t, "feat-b", "main")
	if err := repo.repo.CheckoutBranch("main"); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}
	withAskOne(t, []interface{}{"feat-a"}, func() {
		if err := runUp(nil, nil); err != nil {
			t.Fatalf("runUp prompt failed: %v", err)
		}
	})
}

func TestRunUpCancelled(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-a", "main")
	repo.createBranch(t, "feat-b", "main")
	if err := repo.repo.CheckoutBranch("main"); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}

	withAskOneError(t, terminal.InterruptErr, func() {
		if err := runUp(nil, nil); err != nil {
			t.Fatalf("runUp cancel failed: %v", err)
		}
	})
}

func TestRunUpUntrackedBranch(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := repo.repo.CreateBranch("untracked-up"); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}
	if err := repo.repo.CheckoutBranch("untracked-up"); err != nil {
		t.Fatalf("failed to checkout branch: %v", err)
	}

	if err := runUp(nil, nil); err == nil {
		t.Fatalf("expected runUp untracked error")
	}
}
