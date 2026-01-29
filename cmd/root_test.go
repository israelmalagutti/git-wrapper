package cmd

import "testing"

func TestGetVersionInfo(t *testing.T) {
	Version = "v1.2.3"
	Commit = "abc123"
	BuildDate = "2026-01-28"

	info := GetVersionInfo()
	if info == "" {
		t.Fatalf("expected version info")
	}
}

func TestExecuteDoesNotError(t *testing.T) {
	Execute()
}
