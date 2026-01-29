package cmd

import (
	"os"
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
)

func TestRunFoldOutsideRepo(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(origDir)
	}()

	if err := runFold(nil, nil); err == nil {
		t.Fatalf("expected runFold to fail outside repo")
	}
}

func TestRunFoldConfirmError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevForce := foldForce
	defer func() { foldForce = prevForce }()

	repo.createBranch(t, "feat-fold-confirm", "main")
	if err := repo.repo.CheckoutBranch("feat-fold-confirm"); err != nil {
		t.Fatalf("failed to checkout: %v", err)
	}

	foldForce = false
	withAskOneError(t, terminal.InterruptErr, func() {
		if err := runFold(nil, nil); err == nil {
			t.Fatalf("expected confirmation error")
		}
	})
}

func TestRunFoldMergeError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevForce := foldForce
	defer func() { foldForce = prevForce }()

	repo.createBranch(t, "feat-fold-conflict", "main")
	repo.commitFile(t, "conflict.txt", "child", "child commit")

	if err := repo.repo.CheckoutBranch("main"); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}
	repo.commitFile(t, "conflict.txt", "main", "main commit")

	if err := repo.repo.CheckoutBranch("feat-fold-conflict"); err != nil {
		t.Fatalf("failed to checkout branch: %v", err)
	}

	foldForce = true
	if err := runFold(nil, nil); err == nil {
		t.Fatalf("expected merge error")
	}
}

func TestRunFoldCommitError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevForce := foldForce
	defer func() { foldForce = prevForce }()

	repo.createBranch(t, "feat-fold-commit", "main")
	repo.commitFile(t, "fold.txt", "data", "fold commit")

	if err := repo.repo.CheckoutBranch("feat-fold-commit"); err != nil {
		t.Fatalf("failed to checkout branch: %v", err)
	}

	writeFailingHook(t, repo.dir, "pre-commit")
	foldForce = true

	if err := runFold(nil, nil); err == nil {
		t.Fatalf("expected commit error")
	}
}
