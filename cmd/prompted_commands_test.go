package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AlecAivazis/survey/v2"
)

func withAskOne(t *testing.T, responses []interface{}, fn func()) {
	t.Helper()
	prev := askOne
	idx := 0
	askOne = func(prompt survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		if idx >= len(responses) {
			t.Fatalf("no more responses for prompt %T", prompt)
		}
		switch v := responses[idx].(type) {
		case string:
			ptr, ok := response.(*string)
			if !ok {
				t.Fatalf("expected *string response for %T", prompt)
			}
			*ptr = v
		case bool:
			ptr, ok := response.(*bool)
			if !ok {
				t.Fatalf("expected *bool response for %T", prompt)
			}
			*ptr = v
		case []string:
			ptr, ok := response.(*[]string)
			if !ok {
				t.Fatalf("expected *[]string response for %T", prompt)
			}
			*ptr = v
		default:
			t.Fatalf("unsupported response type %T", v)
		}
		idx++
		return nil
	}
	defer func() { askOne = prev }()
	fn()
}

func withAskOneError(t *testing.T, err error, fn func()) {
	t.Helper()
	prev := askOne
	askOne = func(prompt survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		return err
	}
	defer func() { askOne = prev }()
	fn()
}

func withAskOneSequence(t *testing.T, steps []interface{}, fn func()) {
	t.Helper()
	prev := askOne
	idx := 0
	askOne = func(prompt survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		if idx >= len(steps) {
			t.Fatalf("no more steps for prompt %T", prompt)
		}
		step := steps[idx]
		idx++

		if err, ok := step.(error); ok {
			return err
		}

		switch v := step.(type) {
		case string:
			ptr, ok := response.(*string)
			if !ok {
				t.Fatalf("expected *string response for %T", prompt)
			}
			*ptr = v
		case bool:
			ptr, ok := response.(*bool)
			if !ok {
				t.Fatalf("expected *bool response for %T", prompt)
			}
			*ptr = v
		case []string:
			ptr, ok := response.(*[]string)
			if !ok {
				t.Fatalf("expected *[]string response for %T", prompt)
			}
			*ptr = v
		default:
			t.Fatalf("unsupported step type %T", v)
		}
		return nil
	}
	defer func() { askOne = prev }()
	fn()
}

func setupRawRepo(t *testing.T) (string, func()) {
	t.Helper()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "config", "user.email", "test@test.com").Run(); err != nil {
		t.Fatalf("failed to config email: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "config", "user.name", "Test User").Run(); err != nil {
		t.Fatalf("failed to config name: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "config", "commit.gpgsign", "false").Run(); err != nil {
		t.Fatalf("failed to config gpgsign: %v", err)
	}
	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("failed to write README: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "add", ".").Run(); err != nil {
		t.Fatalf("failed to add: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "commit", "-m", "Initial").Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "branch", "-M", "main").Run(); err != nil {
		t.Fatalf("failed to set main: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	cleanup := func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("failed to restore cwd: %v", err)
		}
	}
	return dir, cleanup
}

func TestRunInitAndTrackWithPrompts(t *testing.T) {
	_, cleanup := setupRawRepo(t)
	defer cleanup()

	withAskOne(t, []interface{}{"main"}, func() {
		if err := runInit(nil, nil); err != nil {
			t.Fatalf("runInit failed: %v", err)
		}
	})

	// Create a new branch to track
	if err := exec.Command("git", "checkout", "-b", "feat-track").Run(); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	withAskOne(t, []interface{}{"main"}, func() {
		if err := runTrack(nil, []string{"feat-track"}); err != nil {
			t.Fatalf("runTrack failed: %v", err)
		}
	})
}

func TestPromptHelpers(t *testing.T) {
	withAskOne(t, []interface{}{"Commit all file changes (--all)"}, func() {
		if got, err := promptNoStagedChanges(); err != nil || got != "all" {
			t.Fatalf("promptNoStagedChanges returned %q, %v", got, err)
		}
	})
	withAskOne(t, []interface{}{"Abort this operation"}, func() {
		if got, err := promptNoStagedChanges(); err != nil || got != "abort" {
			t.Fatalf("promptNoStagedChanges abort returned %q, %v", got, err)
		}
	})

	withAskOne(t, []interface{}{"Commit staged changes"}, func() {
		if got, err := promptHasChanges(true); err != nil || got != "staged" {
			t.Fatalf("promptHasChanges returned %q, %v", got, err)
		}
	})
	withAskOne(t, []interface{}{"Abort this operation"}, func() {
		if got, err := promptHasChanges(true); err != nil || got != "abort" {
			t.Fatalf("promptHasChanges abort returned %q, %v", got, err)
		}
	})

	withAskOne(t, []interface{}{"Commit all file changes (--all)"}, func() {
		if got, err := promptHasChanges(false); err != nil || got != "all" {
			t.Fatalf("promptHasChanges returned %q, %v", got, err)
		}
	})
	withAskOne(t, []interface{}{"Abort this operation"}, func() {
		if got, err := promptHasChanges(false); err != nil || got != "abort" {
			t.Fatalf("promptHasChanges abort returned %q, %v", got, err)
		}
	})

	withAskOne(t, []interface{}{"Abort"}, func() {
		if got, err := promptCommitActionNoMessage(false); err != nil || got != "abort" {
			t.Fatalf("promptCommitActionNoMessage abort returned %q, %v", got, err)
		}
	})

	withAskOne(t, []interface{}{"commit msg"}, func() {
		if got, err := promptCommitMessage(); err != nil || got != "commit msg" {
			t.Fatalf("promptCommitMessage returned %q, %v", got, err)
		}
	})

	repo := setupCmdTestRepo(t)
	if err := os.WriteFile(filepath.Join(repo.dir, "untracked.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write untracked file: %v", err)
	}

	withAskOne(t, []interface{}{true, []string{"untracked.txt"}}, func() {
		if err := promptTrackUntrackedFiles(repo.repo); err != nil {
			t.Fatalf("promptTrackUntrackedFiles failed: %v", err)
		}
	})
	repo.cleanup()

	// User declines tracking with no tracked changes -> errNoChangesToCommit
	repo2 := setupCmdTestRepo(t)
	if err := os.WriteFile(filepath.Join(repo2.dir, "untracked2.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write untracked file: %v", err)
	}
	withAskOne(t, []interface{}{false}, func() {
		if err := promptTrackUntrackedFiles(repo2.repo); err != errNoChangesToCommit {
			t.Fatalf("expected errNoChangesToCommit, got %v", err)
		}
	})
	repo2.cleanup()

	// User selects none when no tracked changes -> errNoChangesToCommit
	repo3 := setupCmdTestRepo(t)
	withAskOne(t, []interface{}{true, []string{}}, func() {
		if err := promptTrackUntrackedFiles(repo3.repo); err != errNoChangesToCommit {
			t.Fatalf("expected errNoChangesToCommit, got %v", err)
		}
	})
	repo3.cleanup()
}
