package cmd

import (
	"os"
	"testing"
)

func TestDoCommitAndPromptAction(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := repo.repo.CheckoutBranch("main"); err != nil {
		t.Fatalf("checkout failed: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("commit", "--allow-empty", "-m", "empty"); err != nil {
		t.Fatalf("failed to create commit: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("commit", "--allow-empty", "-m", "empty2"); err != nil {
		t.Fatalf("failed to create commit: %v", err)
	}

	withAskOne(t, []interface{}{"Stage all changes and commit (--all)"}, func() {
		action, err := promptCommitActionNoMessage(false)
		if err != nil || action != "all" {
			t.Fatalf("promptCommitActionNoMessage returned %q, %v", action, err)
		}
	})
	withAskOne(t, []interface{}{"Select changes to commit (--patch)"}, func() {
		action, err := promptCommitActionNoMessage(false)
		if err != nil || action != "patch" {
			t.Fatalf("promptCommitActionNoMessage returned %q, %v", action, err)
		}
	})
	withAskOne(t, []interface{}{"Commit staged changes"}, func() {
		action, err := promptCommitActionNoMessage(true)
		if err != nil || action != "staged" {
			t.Fatalf("promptCommitActionNoMessage returned %q, %v", action, err)
		}
	})

	if err := os.WriteFile(repo.dir+"/commit.txt", []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "commit.txt"); err != nil {
		t.Fatalf("failed to stage file: %v", err)
	}
	if err := doCommit(repo.repo, "commit message", false); err != nil {
		t.Fatalf("doCommit failed: %v", err)
	}
}
