package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBranchNameHelpers(t *testing.T) {
	t.Run("sanitizeBranchName normalizes input", func(t *testing.T) {
		got := sanitizeBranchName("Feature: Add Login*")
		if got != "feature-add-login" {
			t.Fatalf("unexpected sanitized name: %q", got)
		}
	})

	t.Run("generateBranchName trims and truncates", func(t *testing.T) {
		message := "Add login flow\n\nExtra details"
		got := generateBranchName(message)
		if strings.Contains(got, "\n") {
			t.Fatalf("expected single-line branch name, got %q", got)
		}
		if got != "add-login-flow" {
			t.Fatalf("unexpected branch name: %q", got)
		}
	})

	t.Run("resolveBranchName prefers args", func(t *testing.T) {
		got := resolveBranchName([]string{"Feat_One"}, "ignored")
		if got != "feat_one" {
			t.Fatalf("unexpected resolved name: %q", got)
		}
	})
}

func TestChangeDetectionHelpers(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	t.Run("detects staged and unstaged changes", func(t *testing.T) {
		if detectStagedChanges(repo.repo) {
			t.Fatalf("expected no staged changes")
		}
		if detectUnstagedChanges(repo.repo) {
			t.Fatalf("expected no unstaged changes")
		}

		file := filepath.Join(repo.dir, "change.txt")
		if err := os.WriteFile(file, []byte("change"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		if !detectUnstagedChanges(repo.repo) {
			t.Fatalf("expected unstaged changes")
		}

		if _, err := repo.repo.RunGitCommand("add", "change.txt"); err != nil {
			t.Fatalf("failed to add file: %v", err)
		}
		if !detectStagedChanges(repo.repo) {
			t.Fatalf("expected staged changes")
		}
	})

	t.Run("getUntrackedFiles returns list", func(t *testing.T) {
		file := filepath.Join(repo.dir, "untracked.txt")
		if err := os.WriteFile(file, []byte("data"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		files := getUntrackedFiles(repo.repo)
		if len(files) == 0 {
			t.Fatalf("expected untracked files")
		}
	})
}

func TestCommitHelpers(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := os.WriteFile(filepath.Join(repo.dir, "foo.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "foo.txt"); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	if err := commitChanges(repo.repo, "Add foo", false); err != nil {
		t.Fatalf("commitChanges failed: %v", err)
	}

	if hasTrackedChanges(repo.repo) {
		t.Fatalf("expected no tracked changes after commit")
	}
}
