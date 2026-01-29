package cmd

import "testing"

func TestRunBottomAndDownPaths(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	// Already at trunk
	if err := runBottom(nil, nil); err != nil {
		t.Fatalf("runBottom trunk failed: %v", err)
	}

	repo.createBranch(t, "feat-nav", "main")
	if err := runBottom(nil, nil); err != nil {
		t.Fatalf("runBottom switch failed: %v", err)
	}

	// down invalid step
	if err := runDown(nil, []string{"0"}); err == nil {
		t.Fatalf("expected down invalid step error")
	}

	// down from trunk error
	if err := runDown(nil, nil); err == nil {
		t.Fatalf("expected down already trunk error")
	}

	if err := repo.repo.CheckoutBranch("feat-nav"); err != nil {
		t.Fatalf("failed to checkout feat-nav: %v", err)
	}

	// down steps > 1 reaches trunk
	if err := runDown(nil, []string{"2"}); err != nil {
		t.Fatalf("runDown steps failed: %v", err)
	}
}

func TestRunBottomDetachedHeadError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if _, err := repo.repo.RunGitCommand("checkout", "--detach"); err != nil {
		t.Fatalf("failed to detach: %v", err)
	}
	if err := runBottom(nil, nil); err == nil {
		t.Fatalf("expected runBottom detached head error")
	}
}

func TestRunDownUntrackedBranch(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := repo.repo.CreateBranch("untracked-down"); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}
	if err := repo.repo.CheckoutBranch("untracked-down"); err != nil {
		t.Fatalf("failed to checkout branch: %v", err)
	}

	if err := runDown(nil, nil); err == nil {
		t.Fatalf("expected runDown untracked error")
	}
}

func TestRunBottomMissingConfig(t *testing.T) {
	_, cleanup := setupRawRepo(t)
	defer cleanup()

	if err := runBottom(nil, nil); err == nil {
		t.Fatalf("expected runBottom config error")
	}
}
