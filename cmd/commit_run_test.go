package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
)

func TestRunCommitPaths(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := commitMessage
	prevAll := commitAll
	prevPatch := commitPatch
	defer func() {
		commitMessage = prevMessage
		commitAll = prevAll
		commitPatch = prevPatch
	}()

	// commitMessage with staged changes
	file := filepath.Join(repo.dir, "commit-staged.txt")
	if err := os.WriteFile(file, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "commit-staged.txt"); err != nil {
		t.Fatalf("failed to stage file: %v", err)
	}
	commitMessage = "staged commit"
	if err := runCommit(nil, nil); err != nil {
		t.Fatalf("runCommit staged failed: %v", err)
	}

	// commitMessage with unstaged changes (auto-stage)
	if err := os.WriteFile(filepath.Join(repo.dir, "commit-unstaged.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	commitMessage = "unstaged commit"
	if err := runCommit(nil, nil); err != nil {
		t.Fatalf("runCommit unstaged failed: %v", err)
	}

	// commitMessage with no changes
	commitMessage = "no changes"
	if err := runCommit(nil, nil); err != nil {
		t.Fatalf("runCommit no changes failed: %v", err)
	}

	// No message and no changes
	commitMessage = ""
	if err := runCommit(nil, nil); err != nil {
		t.Fatalf("runCommit no message no changes failed: %v", err)
	}

	// No message, with changes, prompt action
	if err := os.WriteFile(filepath.Join(repo.dir, "commit-prompt.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	withAskOne(t, []interface{}{"Stage all changes and commit (--all)", "prompt msg"}, func() {
		if err := runCommit(nil, nil); err != nil {
			t.Fatalf("runCommit prompt failed: %v", err)
		}
	})

	// Abort path
	if err := os.WriteFile(filepath.Join(repo.dir, "commit-abort.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	withAskOne(t, []interface{}{"Abort"}, func() {
		if err := runCommit(nil, nil); err != nil {
			t.Fatalf("runCommit abort failed: %v", err)
		}
	})

	// No message, staged changes path
	if err := os.WriteFile(filepath.Join(repo.dir, "commit-staged2.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "commit-staged2.txt"); err != nil {
		t.Fatalf("failed to stage file: %v", err)
	}
	withAskOne(t, []interface{}{"Commit staged changes", "staged msg"}, func() {
		if err := runCommit(nil, nil); err != nil {
			t.Fatalf("runCommit staged prompt failed: %v", err)
		}
	})

	// No message, patch path with tracked selection
	t.Setenv("GW_TEST_AUTO_STAGE", "1")
	if err := os.WriteFile(filepath.Join(repo.dir, "commit-patch.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	withAskOne(t, []interface{}{
		"Select changes to commit (--patch)",
		true,
		[]string{"commit-patch.txt"},
		"patch msg",
	}, func() {
		if err := runCommit(nil, nil); err != nil {
			t.Fatalf("runCommit patch failed: %v", err)
		}
	})

	// Cancelled prompt
	if err := os.WriteFile(filepath.Join(repo.dir, "commit-cancel.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	withAskOneError(t, terminal.InterruptErr, func() {
		if err := runCommit(nil, nil); err != nil {
			t.Fatalf("runCommit cancel failed: %v", err)
		}
	})

	// commitAll flag path
	if err := os.WriteFile(filepath.Join(repo.dir, "commit-all.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	commitAll = true
	commitMessage = "commit all"
	if err := runCommit(nil, nil); err != nil {
		t.Fatalf("runCommit commitAll failed: %v", err)
	}
}
