package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
)

type assertedError struct{}

func (assertedError) Error() string { return "asserted" }

func TestDoCommitError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := doCommit(repo.repo, "should fail", false); err == nil {
		t.Fatalf("expected doCommit to fail with no staged changes")
	}
}

func TestRunCommitAllNoMessage(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevAll := commitAll
	prevMessage := commitMessage
	defer func() {
		commitAll = prevAll
		commitMessage = prevMessage
	}()

	if err := os.WriteFile(filepath.Join(repo.dir, "all-nomsg.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	commitAll = true
	commitMessage = ""

	withAskOne(t, []interface{}{"Stage all changes and commit (--all)", "msg"}, func() {
		if err := runCommit(nil, nil); err != nil {
			t.Fatalf("runCommit all no message failed: %v", err)
		}
	})
}

func TestRunCommitPromptError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := os.WriteFile(filepath.Join(repo.dir, "prompt-error.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	withAskOneError(t, terminal.InterruptErr, func() {
		if err := runCommit(nil, nil); err != nil {
			t.Fatalf("runCommit prompt interrupt failed: %v", err)
		}
	})
}

func TestRunCommitPromptCommitMessageCancelled(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := os.WriteFile(filepath.Join(repo.dir, "staged-cancel.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "staged-cancel.txt"); err != nil {
		t.Fatalf("failed to stage file: %v", err)
	}

	withAskOneSequence(t, []interface{}{
		"Commit staged changes",
		terminal.InterruptErr,
	}, func() {
		if err := runCommit(nil, nil); err != nil {
			t.Fatalf("runCommit prompt message cancel failed: %v", err)
		}
	})
}

func TestRunCommitPromptCommitMessageError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := os.WriteFile(filepath.Join(repo.dir, "staged-error.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "staged-error.txt"); err != nil {
		t.Fatalf("failed to stage file: %v", err)
	}

	withAskOneSequence(t, []interface{}{
		"Commit staged changes",
		assertedError{},
	}, func() {
		if err := runCommit(nil, nil); err == nil {
			t.Fatalf("expected commit message error")
		}
	})
}

func TestRunCommitPromptActionError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := os.WriteFile(filepath.Join(repo.dir, "prompt-error2.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	withAskOneSequence(t, []interface{}{assertedError{}}, func() {
		if err := runCommit(nil, nil); err == nil {
			t.Fatalf("expected runCommit prompt error")
		}
	})
}

func TestRunCommitMessagePatch(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := commitMessage
	prevPatch := commitPatch
	defer func() {
		commitMessage = prevMessage
		commitPatch = prevPatch
	}()

	if err := os.WriteFile(filepath.Join(repo.dir, "msgpatch.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	t.Setenv("GW_TEST_AUTO_STAGE", "1")
	commitMessage = "patch message"
	commitPatch = true

	if err := runCommit(nil, nil); err != nil {
		t.Fatalf("runCommit message patch failed: %v", err)
	}
}

func TestRunCommitPatchNoChanges(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := os.WriteFile(filepath.Join(repo.dir, "patch-none.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	withAskOne(t, []interface{}{"Select changes to commit (--patch)", false}, func() {
		if err := runCommit(nil, nil); err != nil {
			t.Fatalf("runCommit patch no changes failed: %v", err)
		}
	})
}

func TestRunCommitPatchPromptError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := os.WriteFile(filepath.Join(repo.dir, "patch-error.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	withAskOneSequence(t, []interface{}{
		"Select changes to commit (--patch)",
		assertedError{},
	}, func() {
		if err := runCommit(nil, nil); err == nil {
			t.Fatalf("expected runCommit patch prompt error")
		}
	})
}

func TestRunCommitStageAllError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevAll := commitAll
	prevMessage := commitMessage
	defer func() {
		commitAll = prevAll
		commitMessage = prevMessage
	}()

	if err := os.WriteFile(filepath.Join(repo.dir, "stage-error.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	t.Setenv("GIT_INDEX_FILE", repo.dir) // invalid index path to force git add failure
	commitAll = true
	commitMessage = "stage error"

	if err := runCommit(nil, nil); err == nil {
		t.Fatalf("expected runCommit stage error")
	}
}

func TestRunCommitMessageStageError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := commitMessage
	defer func() { commitMessage = prevMessage }()

	if err := os.WriteFile(filepath.Join(repo.dir, "unstaged-stage-error.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	gitDir := filepath.Join(repo.dir, ".git")
	if err := os.Chmod(gitDir, 0500); err != nil {
		t.Fatalf("failed to chmod .git: %v", err)
	}
	defer func() {
		_ = os.Chmod(gitDir, 0755)
	}()

	commitMessage = "stage error"
	if err := runCommit(nil, nil); err == nil {
		t.Fatalf("expected stage error in commitMessage path")
	}
}

func TestRunCommitPatchNoChangesUntracked(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := os.WriteFile(filepath.Join(repo.dir, "untracked.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	withAskOneSequence(t, []interface{}{
		"Select changes to commit (--patch)",
		false,
	}, func() {
		if err := runCommit(nil, nil); err != nil {
			t.Fatalf("runCommit patch no changes untracked failed: %v", err)
		}
	})
}

func TestRunCommitAllPromptMessageError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := os.WriteFile(filepath.Join(repo.dir, "all-msg-error.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	withAskOneSequence(t, []interface{}{
		"Stage all changes and commit (--all)",
		assertedError{},
	}, func() {
		if err := runCommit(nil, nil); err == nil {
			t.Fatalf("expected commit message error for all action")
		}
	})
}
