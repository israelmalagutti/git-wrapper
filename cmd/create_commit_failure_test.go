package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunCreateCommitFailureRollback(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevMessage := createMessage
	prevAll := createAll
	defer func() {
		createMessage = prevMessage
		createAll = prevAll
	}()

	if err := os.WriteFile(filepath.Join(repo.dir, "fail.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	writeFailingHook(t, repo.dir, "pre-commit")
	createAll = true
	createMessage = "fail commit"

	if err := runCreate(nil, []string{"feat-create-fail"}); err == nil {
		t.Fatalf("expected runCreate commit failure error")
	}

	if repo.repo.BranchExists("feat-create-fail") {
		t.Fatalf("expected failed branch to be rolled back")
	}
}
