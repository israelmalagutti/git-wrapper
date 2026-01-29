package cmd

import "testing"

func TestRunMove(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-move", "main")
	repo.createBranch(t, "feat-target", "main")

	prevMoveSource := moveSource
	prevMoveOnto := moveOnto
	defer func() {
		moveSource = prevMoveSource
		moveOnto = prevMoveOnto
	}()

	moveSource = "feat-move"
	moveOnto = "feat-target"

	if err := runMove(nil, nil); err != nil {
		t.Fatalf("runMove failed: %v", err)
	}
}

func TestRunFold(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-fold", "main")
	repo.commitFile(t, "fold.txt", "data", "fold change")

	prevFoldForce := foldForce
	prevFoldKeep := foldKeep
	defer func() {
		foldForce = prevFoldForce
		foldKeep = prevFoldKeep
	}()

	foldForce = true
	foldKeep = false

	if err := runFold(nil, nil); err != nil {
		t.Fatalf("runFold failed: %v", err)
	}
}

func TestRunStackRestack(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-restack", "main")

	if err := runStackRestack(nil, nil); err != nil {
		t.Fatalf("runStackRestack failed: %v", err)
	}
}
