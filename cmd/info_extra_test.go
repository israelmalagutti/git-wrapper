package cmd

import "testing"

func TestRunInfoUntrackedAndMissing(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := repo.repo.CreateBranch("untracked-info"); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}
	if err := runInfo(nil, []string{"untracked-info"}); err != nil {
		t.Fatalf("runInfo untracked failed: %v", err)
	}

	if err := runInfo(nil, []string{"missing-info"}); err == nil {
		t.Fatalf("expected missing branch error")
	}
}

func TestRunInfoMissingConfig(t *testing.T) {
	_, cleanup := setupRawRepo(t)
	defer cleanup()

	if err := runInfo(nil, nil); err == nil {
		t.Fatalf("expected runInfo config error")
	}
}
