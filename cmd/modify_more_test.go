package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunModifyPatchWithChild(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-mod", "main")
	repo.commitFile(t, "mod.txt", "base", "base commit")
	repo.createBranch(t, "feat-mod-child", "feat-mod")

	if err := repo.repo.CheckoutBranch("feat-mod"); err != nil {
		t.Fatalf("failed to checkout feat-mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo.dir, "mod.txt"), []byte("update"), 0644); err != nil {
		t.Fatalf("failed to write mod file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "mod.txt"); err != nil {
		t.Fatalf("failed to stage mod file: %v", err)
	}

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

	t.Setenv("GW_TEST_AUTO_STAGE", "1")
	modifyCommit = true
	modifyPatch = true
	modifyMessage = "modify commit"

	if err := runModify(nil, nil); err != nil {
		t.Fatalf("runModify patch failed: %v", err)
	}
}

func TestRunModifyUnstagedError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-unstaged", "main")
	if err := os.WriteFile(filepath.Join(repo.dir, "unstaged.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write unstaged file: %v", err)
	}

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

	modifyCommit = false
	modifyAll = false
	modifyPatch = false
	modifyMessage = ""

	if err := runModify(nil, nil); err == nil {
		t.Fatalf("expected unstaged changes error")
	}
}
