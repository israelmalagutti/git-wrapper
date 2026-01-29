package cmd

import (
	"os"
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
)

func TestRunMoveOutsideRepo(t *testing.T) {
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

	if err := runMove(nil, nil); err == nil {
		t.Fatalf("expected runMove to fail outside repo")
	}
}

func TestRunMoveDetachedHead(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	head, err := repo.repo.RunGitCommand("rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("failed to get head: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("checkout", head); err != nil {
		t.Fatalf("failed to detach head: %v", err)
	}

	if err := runMove(nil, nil); err == nil {
		t.Fatalf("expected runMove detached head error")
	}
}

func TestRunMoveSelectionCancel(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-move", "main")

	prevSource := moveSource
	prevOnto := moveOnto
	defer func() {
		moveSource = prevSource
		moveOnto = prevOnto
	}()

	moveSource = "feat-move"
	moveOnto = ""
	withAskOneError(t, terminal.InterruptErr, func() {
		if err := runMove(nil, nil); err != nil {
			t.Fatalf("expected cancel to return nil, got %v", err)
		}
	})
}

func TestRunMoveSelectionError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-move", "main")

	prevSource := moveSource
	prevOnto := moveOnto
	defer func() {
		moveSource = prevSource
		moveOnto = prevOnto
	}()

	moveSource = "feat-move"
	moveOnto = ""
	withAskOneError(t, assertedError{}, func() {
		if err := runMove(nil, nil); err == nil {
			t.Fatalf("expected selection error")
		}
	})
}

func TestRunMoveMetadataSaveError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-move", "main")

	metaPath := repo.repo.GetMetadataPath()
	if err := os.Chmod(metaPath, 0400); err != nil {
		t.Fatalf("failed to chmod metadata: %v", err)
	}
	defer func() { _ = os.Chmod(metaPath, 0644) }()

	prevSource := moveSource
	prevOnto := moveOnto
	defer func() {
		moveSource = prevSource
		moveOnto = prevOnto
	}()

	moveSource = "feat-move"
	moveOnto = "main"
	if err := runMove(nil, nil); err == nil {
		t.Fatalf("expected metadata save error")
	}
}

func TestRunMoveRebaseFailureCurrentBranch(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-child", "main")
	repo.commitFile(t, "conflict.txt", "child", "child commit")

	if err := repo.repo.CheckoutBranch("main"); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}
	repo.commitFile(t, "conflict.txt", "main", "main commit")

	if err := repo.repo.CheckoutBranch("feat-child"); err != nil {
		t.Fatalf("failed to checkout child: %v", err)
	}

	prevSource := moveSource
	prevOnto := moveOnto
	defer func() {
		moveSource = prevSource
		moveOnto = prevOnto
	}()

	moveSource = "feat-child"
	moveOnto = "main"

	if err := runMove(nil, nil); err == nil {
		t.Fatalf("expected rebase failure")
	}
}
