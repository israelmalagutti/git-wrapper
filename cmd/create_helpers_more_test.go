package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateHelperFunctions(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	// sanitize/generate/resolve
	if got := sanitizeBranchName("Feature: New*Thing"); got == "" {
		t.Fatalf("expected sanitized branch name")
	}
	if got := generateBranchName("Line one\nLine two"); got == "" {
		t.Fatalf("expected generated branch name")
	}
	if got := resolveBranchName([]string{"My Branch"}, ""); got == "" {
		t.Fatalf("expected resolveBranchName from args")
	}
	if got := resolveBranchName(nil, "A very long message that should be truncated to fit within the expected limit for branch names"); got == "" {
		t.Fatalf("expected resolveBranchName from message")
	}

	// staged changes
	if err := os.WriteFile(filepath.Join(repo.dir, "staged.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write staged file: %v", err)
	}
	if _, err := repo.repo.RunGitCommand("add", "staged.txt"); err != nil {
		t.Fatalf("failed to stage file: %v", err)
	}
	if !detectStagedChanges(repo.repo) {
		t.Fatalf("expected detectStagedChanges true")
	}
	if !hasTrackedChanges(repo.repo) {
		t.Fatalf("expected hasTrackedChanges true")
	}

	// unstaged changes
	if err := os.WriteFile(filepath.Join(repo.dir, "unstaged.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write unstaged file: %v", err)
	}
	if !detectUnstagedChanges(repo.repo) {
		t.Fatalf("expected detectUnstagedChanges true")
	}

	// promptTrackUntrackedFiles: no untracked but tracked changes present
	if err := os.Remove(filepath.Join(repo.dir, "unstaged.txt")); err != nil {
		t.Fatalf("failed to remove file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo.dir, "README.md"), []byte("update"), 0644); err != nil {
		t.Fatalf("failed to update tracked file: %v", err)
	}
	if err := promptTrackUntrackedFiles(repo.repo); err != nil {
		t.Fatalf("promptTrackUntrackedFiles unexpected error: %v", err)
	}

	// printNoChangesInfo with untracked files
	if err := os.WriteFile(filepath.Join(repo.dir, "untracked-info.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write untracked file: %v", err)
	}
	printNoChangesInfo(repo.repo)

	// promptTrackUntrackedFiles: untracked present, user selects none but tracked changes exist
	if err := os.WriteFile(filepath.Join(repo.dir, "untracked-extra.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write untracked file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo.dir, "README.md"), []byte("tracked update"), 0644); err != nil {
		t.Fatalf("failed to update tracked file: %v", err)
	}
	withAskOne(t, []interface{}{true, []string{}}, func() {
		if err := promptTrackUntrackedFiles(repo.repo); err != nil {
			t.Fatalf("promptTrackUntrackedFiles expected nil, got %v", err)
		}
	})
}
