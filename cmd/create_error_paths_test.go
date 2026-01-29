package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
)

func TestRunCreateOutsideRepo(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(origDir)
	}()

	if err := runCreate(nil, nil); err == nil {
		t.Fatalf("expected runCreate to fail outside repo")
	}
}

func TestRunCreateDetachedHead(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	head, err := repo.repo.RunGitCommand("rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("failed to get head: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("checkout", head); err != nil {
		t.Fatalf("failed to detach head: %v", err)
	}

	if err := runCreate(nil, []string{"feat-detached"}); err == nil {
		t.Fatalf("expected runCreate detached head error")
	}
}

func TestRunCreateMetadataSaveError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	metaPath := repo.repo.GetMetadataPath()
	if err := os.Chmod(metaPath, 0400); err != nil {
		t.Fatalf("failed to chmod metadata: %v", err)
	}
	defer func() { _ = os.Chmod(metaPath, 0644) }()

	if err := runCreate(nil, []string{"feat-meta-fail"}); err == nil {
		t.Fatalf("expected runCreate metadata save error")
	}
}

func TestRunCreatePromptNoStagedInterrupt(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	defer func() { createMessage = prevMessage }()

	if err := os.WriteFile(filepath.Join(repo.dir, "unstaged.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	createMessage = "msg"
	withAskOneError(t, terminal.InterruptErr, func() {
		if err := runCreate(nil, []string{"feat-create-interrupt"}); err != nil {
			t.Fatalf("expected interrupt to return nil, got %v", err)
		}
	})

	if repo.repo.BranchExists("feat-create-interrupt") {
		t.Fatalf("expected branch to be rolled back on interrupt")
	}
}

func TestRunCreatePromptNoStagedAllStageError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	defer func() { createMessage = prevMessage }()

	if err := os.WriteFile(filepath.Join(repo.dir, "unstaged.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	createMessage = "msg"
	t.Setenv("GIT_INDEX_FILE", repo.dir)

	withAskOne(t, []interface{}{"Commit all file changes (--all)"}, func() {
		if err := runCreate(nil, []string{"feat-create-stage-fail"}); err == nil {
			t.Fatalf("expected stage all error")
		}
	})
}

func TestRunCreatePromptNoStagedAllCommitError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	defer func() { createMessage = prevMessage }()

	if err := os.WriteFile(filepath.Join(repo.dir, "unstaged.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	createMessage = "msg"
	writeFailingHook(t, repo.dir, "pre-commit")

	withAskOne(t, []interface{}{"Commit all file changes (--all)"}, func() {
		if err := runCreate(nil, []string{"feat-create-commit-fail"}); err == nil {
			t.Fatalf("expected commit failure")
		}
	})
}

func TestRunCreatePromptNoStagedPatchInterrupt(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	defer func() { createMessage = prevMessage }()

	if err := os.WriteFile(filepath.Join(repo.dir, "unstaged.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	createMessage = "msg"
	withAskOneSequence(t, []interface{}{
		"Select changes to commit (--patch)",
		terminal.InterruptErr,
	}, func() {
		if err := runCreate(nil, []string{"feat-create-patch-interrupt"}); err != nil {
			t.Fatalf("expected patch interrupt to return nil, got %v", err)
		}
	})

	if repo.repo.BranchExists("feat-create-patch-interrupt") {
		t.Fatalf("expected branch to be rolled back on interrupt")
	}
}

func TestRunCreateNoMessagePromptInterrupt(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	defer func() { createMessage = prevMessage }()

	if err := os.WriteFile(filepath.Join(repo.dir, "unstaged.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	createMessage = ""
	withAskOneError(t, terminal.InterruptErr, func() {
		if err := runCreate(nil, []string{"feat-create-cancel"}); err != nil {
			t.Fatalf("expected prompt interrupt to return nil, got %v", err)
		}
	})
}

func TestRunCreateNoMessageAllStageError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	defer func() { createMessage = prevMessage }()

	if err := os.WriteFile(filepath.Join(repo.dir, "unstaged.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	createMessage = ""
	t.Setenv("GIT_INDEX_FILE", repo.dir)

	withAskOne(t, []interface{}{"Commit all file changes (--all)"}, func() {
		if err := runCreate(nil, []string{"feat-create-all-stage-fail"}); err == nil {
			t.Fatalf("expected stage all error")
		}
	})
}

func TestRunCreateNoMessagePatchErrNoChanges(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	defer func() { createMessage = prevMessage }()

	if err := os.WriteFile(filepath.Join(repo.dir, "untracked.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	createMessage = ""
	withAskOneSequence(t, []interface{}{
		"Select changes to commit (--patch)",
		false,
	}, func() {
		if err := runCreate(nil, []string{"feat-create-nochanges"}); err != nil {
			t.Fatalf("expected errNoChangesToCommit path to return nil, got %v", err)
		}
	})
}

func TestRunCreateNoMessageStagedCommitError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	defer func() { createMessage = prevMessage }()

	if err := os.WriteFile(filepath.Join(repo.dir, "staged.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "staged.txt"); err != nil {
		t.Fatalf("failed to stage file: %v", err)
	}

	createMessage = ""
	writeFailingHook(t, repo.dir, "pre-commit")

	withAskOneSequence(t, []interface{}{
		"Commit staged changes",
		"msg",
	}, func() {
		if err := runCreate(nil, []string{"feat-create-staged-fail"}); err == nil {
			t.Fatalf("expected staged commit failure")
		}
	})
}
