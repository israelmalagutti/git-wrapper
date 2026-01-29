package cmd

import "testing"

func TestRunParentAndChildrenPaths(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-parent", "main")
	repo.createBranch(t, "feat-child", "feat-parent")

	// parent of tracked branch
	if err := runParent(nil, []string{"feat-child"}); err != nil {
		t.Fatalf("runParent failed: %v", err)
	}

	// trunk has no parent
	if err := runParent(nil, []string{"main"}); err == nil {
		t.Fatalf("expected trunk parent error")
	}

	// untracked branch error
	if err := repo.repo.CreateBranch("untracked-parent"); err != nil {
		t.Fatalf("failed to create untracked branch: %v", err)
	}
	if err := runParent(nil, []string{"untracked-parent"}); err == nil {
		t.Fatalf("expected untracked parent error")
	}

	// children via args
	if err := runChildren(nil, []string{"feat-parent"}); err != nil {
		t.Fatalf("runChildren failed: %v", err)
	}

	// children when none
	if err := runChildren(nil, []string{"feat-child"}); err != nil {
		t.Fatalf("runChildren none failed: %v", err)
	}
}

func TestRunParentChildrenMissingConfig(t *testing.T) {
	_, cleanup := setupRawRepo(t)
	defer cleanup()

	if err := runParent(nil, nil); err == nil {
		t.Fatalf("expected runParent config error")
	}
	if err := runChildren(nil, nil); err == nil {
		t.Fatalf("expected runChildren config error")
	}
}

func TestRunChildrenDetachedHead(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if _, err := repo.repo.RunGitCommand("checkout", "--detach"); err != nil {
		t.Fatalf("failed to detach: %v", err)
	}
	if err := runChildren(nil, nil); err == nil {
		t.Fatalf("expected runChildren detached head error")
	}
}

func TestRunParentDetachedHead(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if _, err := repo.repo.RunGitCommand("checkout", "--detach"); err != nil {
		t.Fatalf("failed to detach: %v", err)
	}
	if err := runParent(nil, nil); err == nil {
		t.Fatalf("expected runParent detached head error")
	}
}
