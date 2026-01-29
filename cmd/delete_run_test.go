package cmd

import (
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
)

func TestRunDeleteWithChildren(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-parent", "main")
	repo.createBranch(t, "feat-child", "feat-parent")

	// Ensure current branch is the one being deleted
	if err := repo.repo.CheckoutBranch("feat-parent"); err != nil {
		t.Fatalf("failed to checkout parent: %v", err)
	}

	prevForce := deleteForce
	defer func() { deleteForce = prevForce }()

	deleteForce = false
	withAskOne(t, []interface{}{true}, func() {
		if err := runDelete(nil, []string{"feat-parent"}); err != nil {
			t.Fatalf("runDelete failed: %v", err)
		}
	})
}

func TestRunDeleteInteractiveCancel(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-del", "main")

	prevForce := deleteForce
	defer func() { deleteForce = prevForce }()
	deleteForce = false

	// Interactive selection then cancel confirmation
	withAskOne(t, []interface{}{"feat-del (current, parent: main)", false}, func() {
		if err := runDelete(nil, nil); err != nil {
			t.Fatalf("runDelete interactive failed: %v", err)
		}
	})
}

func TestRunDeleteErrors(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	// Cannot delete trunk
	if err := runDelete(nil, []string{"main"}); err == nil {
		t.Fatalf("expected trunk delete error")
	}

	// Branch does not exist
	if err := runDelete(nil, []string{"missing"}); err == nil {
		t.Fatalf("expected missing branch error")
	}

	// Branch not tracked
	if err := repo.repo.CreateBranch("untracked-del"); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}
	if err := runDelete(nil, []string{"untracked-del"}); err == nil {
		t.Fatalf("expected not tracked error")
	}
}

func TestRunDeletePromptError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-del-error", "main")

	prevForce := deleteForce
	defer func() { deleteForce = prevForce }()
	deleteForce = false

	withAskOneError(t, terminal.InterruptErr, func() {
		if err := runDelete(nil, []string{"feat-del-error"}); err == nil {
			t.Fatalf("expected delete confirmation error")
		}
	})
}
