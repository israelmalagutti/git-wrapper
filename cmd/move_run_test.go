package cmd

import "testing"

func TestRunMoveErrorsAndInteractive(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-parent", "main")
	repo.createBranch(t, "feat-child", "feat-parent")

	prevSource := moveSource
	prevOnto := moveOnto
	defer func() {
		moveSource = prevSource
		moveOnto = prevOnto
	}()

	// Cannot move trunk
	moveSource = "main"
	moveOnto = "feat-parent"
	if err := runMove(nil, nil); err == nil {
		t.Fatalf("expected trunk move error")
	}

	// Source not tracked
	if err := repo.repo.CreateBranch("untracked"); err != nil {
		t.Fatalf("failed to create untracked: %v", err)
	}
	moveSource = "untracked"
	moveOnto = "feat-parent"
	if err := runMove(nil, nil); err == nil {
		t.Fatalf("expected untracked source error")
	}

	// Target is same as source
	moveSource = "feat-parent"
	moveOnto = "feat-parent"
	if err := runMove(nil, nil); err == nil {
		t.Fatalf("expected self target error")
	}

	// Target missing
	moveSource = "feat-parent"
	moveOnto = "missing"
	if err := runMove(nil, nil); err == nil {
		t.Fatalf("expected missing target error")
	}

	// Target is descendant
	moveSource = "feat-parent"
	moveOnto = "feat-child"
	if err := runMove(nil, nil); err == nil {
		t.Fatalf("expected descendant target error")
	}

	// Interactive target selection
	moveSource = "feat-parent"
	moveOnto = ""
	withAskOne(t, []interface{}{"main"}, func() {
		if err := runMove(nil, nil); err != nil {
			t.Fatalf("runMove interactive failed: %v", err)
		}
	})

	// Interactive selection without mapping key
	moveSource = "feat-child"
	moveOnto = ""
	if err := repo.repo.CreateBranch("other"); err != nil {
		t.Fatalf("failed to create other: %v", err)
	}
	withAskOne(t, []interface{}{"other"}, func() {
		if err := runMove(nil, nil); err != nil {
			t.Fatalf("runMove other selection failed: %v", err)
		}
	})
}
