package cmd

import (
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
)

func TestRunCheckoutInteractiveAndFlags(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-1", "main")
	repo.createBranch(t, "feat-2", "feat-1")
	if err := repo.repo.CreateBranch("untracked"); err != nil {
		t.Fatalf("failed to create untracked: %v", err)
	}

	prevTrunk := checkoutTrunk
	prevShow := checkoutShowUntracked
	prevStack := checkoutStack
	defer func() {
		checkoutTrunk = prevTrunk
		checkoutShowUntracked = prevShow
		checkoutStack = prevStack
	}()

	// Trunk flag path
	checkoutTrunk = true
	if err := runCheckout(nil, nil); err != nil {
		t.Fatalf("runCheckout trunk failed: %v", err)
	}
	checkoutTrunk = false

	// Interactive selection, tracked branches only
	withAskOne(t, []interface{}{"feat-1"}, func() {
		if err := runCheckout(nil, nil); err != nil {
			t.Fatalf("runCheckout interactive failed: %v", err)
		}
	})

	// Direct checkout with args
	if err := runCheckout(nil, []string{"feat-2"}); err != nil {
		t.Fatalf("runCheckout direct failed: %v", err)
	}

	// Missing branch error
	if err := runCheckout(nil, []string{"missing"}); err == nil {
		t.Fatalf("expected missing branch error")
	}

	// Show untracked branches and select one
	checkoutShowUntracked = true
	withAskOne(t, []interface{}{"untracked"}, func() {
		if err := runCheckout(nil, nil); err != nil {
			t.Fatalf("runCheckout untracked failed: %v", err)
		}
	})
	checkoutShowUntracked = false

	// Stack-only filter
	checkoutStack = true
	if err := repo.repo.CheckoutBranch("feat-2"); err != nil {
		t.Fatalf("failed to checkout feat-2: %v", err)
	}
	withAskOne(t, []interface{}{"main"}, func() {
		if err := runCheckout(nil, nil); err != nil {
			t.Fatalf("runCheckout stack failed: %v", err)
		}
	})
	checkoutStack = false
}

func TestRunCheckoutNoBranchesMatch(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	if err := repo.repo.CreateBranch("untracked-only"); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}
	if err := repo.repo.CheckoutBranch("untracked-only"); err != nil {
		t.Fatalf("failed to checkout: %v", err)
	}

	prevShow := checkoutShowUntracked
	prevStack := checkoutStack
	defer func() {
		checkoutShowUntracked = prevShow
		checkoutStack = prevStack
	}()

	// Not tracked, stack-only filter should result in no branches
	checkoutShowUntracked = false
	checkoutStack = true
	if err := runCheckout(nil, nil); err == nil {
		t.Fatalf("expected no branches match error")
	}
}

func TestRunCheckoutCancelled(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	withAskOneError(t, terminal.InterruptErr, func() {
		if err := runCheckout(nil, nil); err != nil {
			t.Fatalf("runCheckout cancel failed: %v", err)
		}
	})
}

func TestRunCheckoutPromptError(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	withAskOneSequence(t, []interface{}{assertedError{}}, func() {
		if err := runCheckout(nil, nil); err == nil {
			t.Fatalf("expected runCheckout prompt error")
		}
	})
}
