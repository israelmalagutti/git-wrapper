package cmd

import (
	"os"
	"os/exec"
	"testing"
)

func TestExecuteError(t *testing.T) {
	if os.Getenv("GW_TEST_EXECUTE") == "1" {
		rootCmd.SetArgs([]string{"no-such-command"})
		Execute()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExecuteError")
	cmd.Env = append(os.Environ(), "GW_TEST_EXECUTE=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected Execute to exit non-zero")
	}
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 0 {
		t.Fatalf("expected non-zero exit code")
	}
}
