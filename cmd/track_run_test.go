package cmd

import "testing"

func TestRunTrackPaths(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	// Create branch to track
	if err := repo.repo.CreateBranch("feat-track"); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	// Prompt to select parent
	withAskOne(t, []interface{}{"main"}, func() {
		if err := runTrack(nil, []string{"feat-track"}); err != nil {
			t.Fatalf("runTrack failed: %v", err)
		}
	})

	// Already tracked should error
	if err := runTrack(nil, []string{"feat-track"}); err == nil {
		t.Fatalf("expected already tracked error")
	}

	// Missing branch should error
	if err := runTrack(nil, []string{"missing"}); err == nil {
		t.Fatalf("expected missing branch error")
	}
}
