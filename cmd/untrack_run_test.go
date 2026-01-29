package cmd

import "testing"

func TestRunUntrackConfirmCancel(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-parent", "main")
	repo.createBranch(t, "feat-child", "feat-parent")

	prevForce := untrackForce
	defer func() { untrackForce = prevForce }()

	untrackForce = false
	withAskOne(t, []interface{}{false}, func() {
		if err := runUntrack(nil, []string{"feat-parent"}); err != nil {
			t.Fatalf("runUntrack failed: %v", err)
		}
	})
}

func TestRunUntrackErrors(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := runUntrack(nil, []string{"main"}); err == nil {
		t.Fatalf("expected error untracking trunk")
	}

	if err := runUntrack(nil, []string{"missing"}); err != nil {
		t.Fatalf("runUntrack missing branch should be no-op: %v", err)
	}
}
