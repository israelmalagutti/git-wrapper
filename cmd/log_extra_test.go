package cmd

import "testing"

func TestRunLogLong(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevShort := logShort
	prevLong := logLong
	defer func() {
		logShort = prevShort
		logLong = prevLong
	}()

	logShort = false
	logLong = true
	if err := runLog(nil, nil); err != nil {
		t.Fatalf("runLog long failed: %v", err)
	}
}

func TestRunLogMissingConfig(t *testing.T) {
	_, cleanup := setupRawRepo(t)
	defer cleanup()

	if err := runLog(nil, nil); err == nil {
		t.Fatalf("expected runLog config error")
	}
}
