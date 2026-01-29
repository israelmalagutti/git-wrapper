package cmd

import "testing"

func TestRunSplitErrors(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevCommit := splitByCommit
	prevHunk := splitByHunk
	prevFile := splitByFile
	prevName := splitName
	defer func() {
		splitByCommit = prevCommit
		splitByHunk = prevHunk
		splitByFile = prevFile
		splitName = prevName
	}()

	// Cannot split trunk
	if err := runSplit(nil, nil); err == nil {
		t.Fatalf("expected trunk split error")
	}

	// Branch not tracked
	if err := repo.repo.CreateBranch("untracked"); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}
	if err := repo.repo.CheckoutBranch("untracked"); err != nil {
		t.Fatalf("failed to checkout: %v", err)
	}
	if err := runSplit(nil, nil); err == nil {
		t.Fatalf("expected not tracked error")
	}

	// Tracked branch with no commits
	repo.createBranch(t, "feat-empty", "main")
	if err := repo.repo.CheckoutBranch("feat-empty"); err != nil {
		t.Fatalf("failed to checkout: %v", err)
	}
	splitByCommit = true
	if err := runSplit(nil, nil); err == nil {
		t.Fatalf("expected no commits error")
	}

	// Multiple modes specified
	repo.commitFile(t, "split.txt", "data", "split commit")
	splitByCommit = true
	splitByHunk = true
	if err := runSplit(nil, nil); err == nil {
		t.Fatalf("expected multiple mode error")
	}
}
