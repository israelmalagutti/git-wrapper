package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
)

func TestRunCreateMessageGeneratedBranch(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	prevAll := createAll
	prevPatch := createPatch
	defer func() {
		createMessage = prevMessage
		createAll = prevAll
		createPatch = prevPatch
	}()

	createMessage = "Add login flow"
	if err := runCreate(nil, nil); err != nil {
		t.Fatalf("runCreate generated name failed: %v", err)
	}
	if !repo.repo.BranchExists("add-login-flow") {
		t.Fatalf("expected generated branch name to exist")
	}
}

func TestRunCreateMessagePatchPath(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	prevAll := createAll
	prevPatch := createPatch
	defer func() {
		createMessage = prevMessage
		createAll = prevAll
		createPatch = prevPatch
	}()

	if err := os.WriteFile(filepath.Join(repo.dir, "patch.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	t.Setenv("GW_TEST_AUTO_STAGE", "1")
	createMessage = "patch commit"
	withAskOne(t, []interface{}{
		"Select changes to commit (--patch)",
		true,
		[]string{"patch.txt"},
	}, func() {
		if err := runCreate(nil, []string{"feat-create-patch-msg"}); err != nil {
			t.Fatalf("runCreate patch message failed: %v", err)
		}
	})
}

func TestRunCreateNoMessageActions(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	prevAll := createAll
	prevPatch := createPatch
	defer func() {
		createMessage = prevMessage
		createAll = prevAll
		createPatch = prevPatch
	}()

	// Action: all
	if err := os.WriteFile(filepath.Join(repo.dir, "all.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	createMessage = ""
	withAskOne(t, []interface{}{"Commit all file changes (--all)", "all commit"}, func() {
		if err := runCreate(nil, []string{"feat-create-all-msg"}); err != nil {
			t.Fatalf("runCreate all action failed: %v", err)
		}
	})

	// Action: patch
	if err := os.WriteFile(filepath.Join(repo.dir, "patch2.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	t.Setenv("GW_TEST_AUTO_STAGE", "1")
	withAskOne(t, []interface{}{
		"Select changes to commit (--patch)",
		true,
		[]string{"patch2.txt"},
		"patch commit",
	}, func() {
		if err := runCreate(nil, []string{"feat-create-patch2"}); err != nil {
			t.Fatalf("runCreate patch action failed: %v", err)
		}
	})

	// Action: no-commit
	if err := os.WriteFile(filepath.Join(repo.dir, "nocommit.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	withAskOne(t, []interface{}{"Create a branch with no commit"}, func() {
		if err := runCreate(nil, []string{"feat-create-nocommit2"}); err != nil {
			t.Fatalf("runCreate no-commit action failed: %v", err)
		}
	})
}

func TestRunCreatePatchNoChanges(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	defer func() { createMessage = prevMessage }()

	if err := os.WriteFile(filepath.Join(repo.dir, "patch-none.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	createMessage = "patch none"
	withAskOne(t, []interface{}{"Select changes to commit (--patch)", false}, func() {
		if err := runCreate(nil, []string{"feat-create-patch-none"}); err != nil {
			t.Fatalf("runCreate patch none failed: %v", err)
		}
	})
}

func TestRunCreatePromptErrors(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	prevAll := createAll
	prevPatch := createPatch
	defer func() {
		createMessage = prevMessage
		createAll = prevAll
		createPatch = prevPatch
	}()

	// Branch name prompt error
	createMessage = ""
	withAskOneSequence(t, []interface{}{assertedError{}}, func() {
		if err := runCreate(nil, nil); err == nil {
			t.Fatalf("expected branch name prompt error")
		}
	})

	// No staged changes prompt error with message
	if err := os.WriteFile(filepath.Join(repo.dir, "unstaged.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	createMessage = "msg"
	withAskOneSequence(t, []interface{}{assertedError{}}, func() {
		if err := runCreate(nil, []string{"feat-create-error"}); err == nil {
			t.Fatalf("expected promptNoStagedChanges error")
		}
	})

	// promptHasChanges error
	if err := os.WriteFile(filepath.Join(repo.dir, "unstaged2.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	createMessage = ""
	withAskOneSequence(t, []interface{}{assertedError{}}, func() {
		if err := runCreate(nil, []string{"feat-create-error2"}); err == nil {
			t.Fatalf("expected promptHasChanges error")
		}
	})

	// promptCommitMessage interrupt
	if _, err := repo.repo.RunGitCommand("add", "unstaged2.txt"); err != nil {
		t.Fatalf("failed to stage file: %v", err)
	}
	withAskOneSequence(t, []interface{}{"Commit staged changes", terminal.InterruptErr}, func() {
		if err := runCreate(nil, []string{"feat-create-cancel"}); err != nil {
			t.Fatalf("runCreate commit message cancel failed: %v", err)
		}
	})
}

func TestRunCreatePromptEmptyName(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	defer func() { createMessage = prevMessage }()

	createMessage = ""
	withAskOneSequence(t, []interface{}{"   "}, func() {
		if err := runCreate(nil, nil); err == nil {
			t.Fatalf("expected empty branch name error")
		}
	})
}

func TestRunCreateMissingMetadata(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := os.Remove(repo.repo.GetMetadataPath()); err != nil {
		t.Fatalf("failed to remove metadata: %v", err)
	}

	if err := runCreate(nil, []string{"feat-create-missing-meta"}); err != nil {
		t.Fatalf("runCreate missing metadata failed: %v", err)
	}
}

func TestRunCreatePatchNoChangesUntracked(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	defer func() { createMessage = prevMessage }()

	if err := os.WriteFile(filepath.Join(repo.dir, "untracked.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	createMessage = "patch untracked"
	withAskOneSequence(t, []interface{}{
		"Select changes to commit (--patch)",
		false,
	}, func() {
		if err := runCreate(nil, []string{"feat-create-patch-untracked"}); err != nil {
			t.Fatalf("runCreate patch untracked failed: %v", err)
		}
	})
}

func TestRunCreateAllPromptMessageError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	defer func() { createMessage = prevMessage }()

	if err := os.WriteFile(filepath.Join(repo.dir, "all-msg-error.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	createMessage = ""
	withAskOneSequence(t, []interface{}{
		"Commit all file changes (--all)",
		assertedError{},
	}, func() {
		if err := runCreate(nil, []string{"feat-create-all-msg-error"}); err == nil {
			t.Fatalf("expected promptCommitMessage error")
		}
	})
}

func TestRunCreateStageAllError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	prevAll := createAll
	defer func() {
		createMessage = prevMessage
		createAll = prevAll
	}()

	if err := os.WriteFile(filepath.Join(repo.dir, "stage-error.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	t.Setenv("GIT_INDEX_FILE", repo.dir)
	createAll = true
	createMessage = "stage error"
	if err := runCreate(nil, []string{"feat-create-stage-error"}); err == nil {
		t.Fatalf("expected runCreate stage all error")
	}
}
