package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunModifyPaths(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevCommit := modifyCommit
	prevAll := modifyAll
	prevPatch := modifyPatch
	prevMessage := modifyMessage
	defer func() {
		modifyCommit = prevCommit
		modifyAll = prevAll
		modifyPatch = prevPatch
		modifyMessage = prevMessage
	}()

	// Untracked branch error
	if err := repo.repo.CreateBranch("untracked-modify"); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}
	if err := repo.repo.CheckoutBranch("untracked-modify"); err != nil {
		t.Fatalf("failed to checkout: %v", err)
	}
	if err := runModify(nil, nil); err == nil {
		t.Fatalf("expected error for untracked branch")
	}

	// Tracked branch with no commits -> create commit
	repo.createBranch(t, "feat-nocommit", "main")
	if err := repo.repo.CheckoutBranch("feat-nocommit"); err != nil {
		t.Fatalf("failed to checkout: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo.dir, "nocommit.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	modifyCommit = false
	modifyAll = true
	modifyMessage = "first commit"
	if err := runModify(nil, nil); err != nil {
		t.Fatalf("runModify nocommit failed: %v", err)
	}

	// Unstaged tracked changes error when no flags
	if err := os.WriteFile(filepath.Join(repo.dir, "nocommit.txt"), []byte("updated"), 0644); err != nil {
		t.Fatalf("failed to update file: %v", err)
	}
	modifyCommit = false
	modifyAll = false
	modifyPatch = false
	modifyMessage = ""
	if err := runModify(nil, nil); err == nil {
		t.Fatalf("expected unstaged changes error")
	}

	// Amend path with staged changes
	if _, err := repo.repo.RunGitCommand("add", "nocommit.txt"); err != nil {
		t.Fatalf("failed to stage: %v", err)
	}
	modifyCommit = false
	modifyAll = false
	modifyMessage = ""
	if err := runModify(nil, nil); err != nil {
		t.Fatalf("runModify amend failed: %v", err)
	}
}
