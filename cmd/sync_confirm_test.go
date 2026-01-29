package cmd

import (
	"os"
	"testing"
)

func TestSyncConfirmHelpers(t *testing.T) {
	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	tests := []struct {
		input  string
		expect bool
	}{
		{"y\n", true},
		{"yes\n", true},
		{"n\n", false},
		{"\n", false},
	}

	for _, tc := range tests {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("failed to create pipe: %v", err)
		}
		if _, err := w.Write([]byte(tc.input)); err != nil {
			t.Fatalf("failed to write pipe: %v", err)
		}
		w.Close()
		os.Stdin = r

		if got := confirm(); got != tc.expect {
			t.Fatalf("confirm(%q) = %v, want %v", tc.input, got, tc.expect)
		}
	}
}

func TestSyncConfirmWithOptions(t *testing.T) {
	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	tests := []struct {
		input  string
		expect string
	}{
		{"y\n", "yes"},
		{"yes\n", "yes"},
		{"a\n", "all"},
		{"all\n", "all"},
		{"q\n", "quit"},
		{"quit\n", "quit"},
		{"n\n", "no"},
		{"\n", "no"},
	}

	for _, tc := range tests {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("failed to create pipe: %v", err)
		}
		if _, err := w.Write([]byte(tc.input)); err != nil {
			t.Fatalf("failed to write pipe: %v", err)
		}
		w.Close()
		os.Stdin = r

		if got := confirmWithOptions(); got != tc.expect {
			t.Fatalf("confirmWithOptions(%q) = %s, want %s", tc.input, got, tc.expect)
		}
	}
}
