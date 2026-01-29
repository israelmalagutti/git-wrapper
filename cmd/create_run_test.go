package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
)

func TestRunCreatePaths(t *testing.T) {
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

	// Message with staged changes
	if err := os.WriteFile(filepath.Join(repo.dir, "create-staged.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "create-staged.txt"); err != nil {
		t.Fatalf("failed to stage file: %v", err)
	}
	createMessage = "create staged"
	if err := runCreate(nil, []string{"feat-create-staged"}); err != nil {
		t.Fatalf("runCreate staged failed: %v", err)
	}

	// Message with unstaged changes -> prompt no staged changes
	if err := os.WriteFile(filepath.Join(repo.dir, "create-unstaged.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	createMessage = "create unstaged"
	withAskOne(t, []interface{}{"Commit all file changes (--all)"}, func() {
		if err := runCreate(nil, []string{"feat-create-unstaged"}); err != nil {
			t.Fatalf("runCreate unstaged failed: %v", err)
		}
	})

	// Message with unstaged changes -> no-commit
	if err := os.WriteFile(filepath.Join(repo.dir, "create-nocommit.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	createMessage = "create no commit"
	withAskOne(t, []interface{}{"Create a branch with no commit"}, func() {
		if err := runCreate(nil, []string{"feat-create-nocommit"}); err != nil {
			t.Fatalf("runCreate no-commit failed: %v", err)
		}
	})

	// Message with unstaged changes -> abort
	if err := os.WriteFile(filepath.Join(repo.dir, "create-abort.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	createMessage = "create abort"
	withAskOne(t, []interface{}{"Abort this operation"}, func() {
		if err := runCreate(nil, []string{"feat-create-abort"}); err != nil {
			t.Fatalf("runCreate abort failed: %v", err)
		}
	})

	// No message, with staged changes -> commit staged
	if err := os.WriteFile(filepath.Join(repo.dir, "create-staged2.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "create-staged2.txt"); err != nil {
		t.Fatalf("failed to stage file: %v", err)
	}
	createMessage = ""
	withAskOne(t, []interface{}{"Commit staged changes", "staged msg"}, func() {
		if err := runCreate(nil, []string{"feat-create-staged2"}); err != nil {
			t.Fatalf("runCreate staged prompt failed: %v", err)
		}
	})

	// No message, with changes -> abort
	if err := os.WriteFile(filepath.Join(repo.dir, "create-abort2.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	createMessage = ""
	withAskOne(t, []interface{}{"Abort this operation"}, func() {
		if err := runCreate(nil, []string{"feat-create-abort2"}); err != nil {
			t.Fatalf("runCreate abort2 failed: %v", err)
		}
	})

	// createAll flag path
	if err := os.WriteFile(filepath.Join(repo.dir, "create-all.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	createMessage = "create all"
	createAll = true
	if err := runCreate(nil, []string{"feat-create-all"}); err != nil {
		t.Fatalf("runCreate all failed: %v", err)
	}
	createAll = false

	// No changes at all -> next steps output
	createMessage = ""
	if err := runCreate(nil, []string{"feat-create-clean"}); err != nil {
		t.Fatalf("runCreate clean failed: %v", err)
	}

	// Prompt branch name when no args/message
	createMessage = ""
	withAskOne(t, []interface{}{"Feature Branch"}, func() {
		if err := runCreate(nil, nil); err != nil {
			t.Fatalf("runCreate prompt name failed: %v", err)
		}
	})

	withAskOneError(t, terminal.InterruptErr, func() {
		if err := runCreate(nil, nil); err != nil {
			t.Fatalf("runCreate prompt cancel failed: %v", err)
		}
	})

	// Branch exists error
	if err := runCreate(nil, []string{"feat-create-clean"}); err == nil {
		t.Fatalf("expected branch exists error")
	}

	// Invalid branch name error
	if err := runCreate(nil, []string{"."}); err == nil {
		t.Fatalf("expected invalid branch name error")
	}

	// Message with staged changes and patch flag
	t.Setenv("GW_TEST_AUTO_STAGE", "1")
	if err := os.WriteFile(filepath.Join(repo.dir, "create-patch.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "create-patch.txt"); err != nil {
		t.Fatalf("failed to stage file: %v", err)
	}
	createMessage = "create patch"
	createPatch = true
	if err := runCreate(nil, []string{"feat-create-patch"}); err != nil {
		t.Fatalf("runCreate patch failed: %v", err)
	}
	createPatch = false
}
