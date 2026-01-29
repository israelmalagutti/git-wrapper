package cmd

import "testing"

func TestRunFoldKeepBranch(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-fold", "main")
	repo.commitFile(t, "fold.txt", "data", "fold commit")

	prevKeep := foldKeep
	prevForce := foldForce
	defer func() {
		foldKeep = prevKeep
		foldForce = prevForce
	}()

	foldKeep = true
	foldForce = true
	if err := runFold(nil, nil); err != nil {
		t.Fatalf("runFold keep failed: %v", err)
	}
}

func TestRunFoldCancel(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-fold-cancel", "main")

	prevKeep := foldKeep
	prevForce := foldForce
	defer func() {
		foldKeep = prevKeep
		foldForce = prevForce
	}()

	foldKeep = false
	foldForce = false
	withAskOne(t, []interface{}{false}, func() {
		if err := runFold(nil, nil); err != nil {
			t.Fatalf("runFold cancel failed: %v", err)
		}
	})
}
