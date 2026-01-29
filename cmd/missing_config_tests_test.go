package cmd

import "testing"

func TestRunCommitMissingConfig(t *testing.T) {
	_, cleanup := setupRawRepo(t)
	defer cleanup()

	if err := runCommit(nil, nil); err == nil {
		t.Fatalf("expected runCommit config error")
	}
}

func TestRunCreateMissingConfig(t *testing.T) {
	_, cleanup := setupRawRepo(t)
	defer cleanup()

	if err := runCreate(nil, []string{"feat-missing-config"}); err == nil {
		t.Fatalf("expected runCreate config error")
	}
}

func TestRunMoveMissingConfig(t *testing.T) {
	_, cleanup := setupRawRepo(t)
	defer cleanup()

	if err := runMove(nil, nil); err == nil {
		t.Fatalf("expected runMove config error")
	}
}

func TestRunSyncMissingConfig(t *testing.T) {
	_, cleanup := setupRawRepo(t)
	defer cleanup()

	if err := runSync(nil, nil); err == nil {
		t.Fatalf("expected runSync config error")
	}
}

func TestRunSplitMissingConfig(t *testing.T) {
	_, cleanup := setupRawRepo(t)
	defer cleanup()

	if err := runSplit(nil, nil); err == nil {
		t.Fatalf("expected runSplit config error")
	}
}
