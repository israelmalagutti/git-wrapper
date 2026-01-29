package cmd

import (
	"os"
	"testing"
)

func TestRunNavigationCommands(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-1", "main")
	repo.createBranch(t, "feat-2", "feat-1")

	if err := runBottom(nil, nil); err != nil {
		t.Fatalf("runBottom failed: %v", err)
	}
	if err := runTop(nil, nil); err != nil {
		t.Fatalf("runTop failed: %v", err)
	}
	if err := runDown(nil, nil); err != nil {
		t.Fatalf("runDown failed: %v", err)
	}
	if err := runUp(nil, nil); err != nil {
		t.Fatalf("runUp failed: %v", err)
	}
}

func TestRunInfoLogParentChildren(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-info", "main")

	if err := runParent(nil, []string{"feat-info"}); err != nil {
		t.Fatalf("runParent failed: %v", err)
	}
	if err := runChildren(nil, []string{"main"}); err != nil {
		t.Fatalf("runChildren failed: %v", err)
	}

	logShort = true
	logLong = false
	if err := runLog(nil, nil); err != nil {
		t.Fatalf("runLog failed: %v", err)
	}
	logShort = false
	logLong = false

	if err := runInfo(nil, nil); err != nil {
		t.Fatalf("runInfo failed: %v", err)
	}
}

func TestRunCheckoutAndRename(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	repo.createBranch(t, "feat-check", "main")

	if err := runCheckout(nil, []string{"feat-check"}); err != nil {
		t.Fatalf("runCheckout failed: %v", err)
	}

	if err := runRename(nil, []string{"feat-renamed"}); err != nil {
		t.Fatalf("runRename failed: %v", err)
	}
}

func TestRunCreateCommitModifyDeleteUntrack(t *testing.T) {
	repo := setupCmdTestRepo(t)
	defer repo.cleanup()

	prevCreateMessage := createMessage
	prevCreateAll := createAll
	prevCreatePatch := createPatch
	prevCommitMessage := commitMessage
	prevCommitAll := commitAll
	prevCommitPatch := commitPatch
	prevModifyCommit := modifyCommit
	prevModifyAll := modifyAll
	prevModifyPatch := modifyPatch
	prevModifyMessage := modifyMessage
	prevUntrackForce := untrackForce
	prevDeleteForce := deleteForce
	defer func() {
		createMessage = prevCreateMessage
		createAll = prevCreateAll
		createPatch = prevCreatePatch
		commitMessage = prevCommitMessage
		commitAll = prevCommitAll
		commitPatch = prevCommitPatch
		modifyCommit = prevModifyCommit
		modifyAll = prevModifyAll
		modifyPatch = prevModifyPatch
		modifyMessage = prevModifyMessage
		untrackForce = prevUntrackForce
		deleteForce = prevDeleteForce
	}()

	// Create a new branch with no changes
	createMessage = ""
	createAll = false
	createPatch = false
	if err := runCreate(nil, []string{"feat-create"}); err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}

	// Commit a change on the new branch
	if err := repo.repo.CheckoutBranch("feat-create"); err != nil {
		t.Fatalf("checkout failed: %v", err)
	}
	repo.commitFile(t, "change.txt", "data", "change")
	commitMessage = "commit change"
	commitAll = false
	commitPatch = false
	if err := runCommit(nil, nil); err != nil {
		t.Fatalf("runCommit failed: %v", err)
	}

	// Modify with a new commit
	if err := os.WriteFile(repo.dir+"/change2.txt", []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write change2: %v", err)
	}
	modifyCommit = true
	modifyAll = true
	modifyPatch = false
	modifyMessage = "modify commit"
	if err := runModify(nil, nil); err != nil {
		t.Fatalf("runModify failed: %v", err)
	}

	// Create another branch to untrack
	createMessage = ""
	createAll = false
	createPatch = false
	if err := runCreate(nil, []string{"feat-untrack"}); err != nil {
		t.Fatalf("runCreate (untrack branch) failed: %v", err)
	}

	// Untrack the new branch
	untrackForce = true
	if err := runUntrack(nil, []string{"feat-untrack"}); err != nil {
		t.Fatalf("runUntrack failed: %v", err)
	}
	untrackForce = false

	// Delete the branch
	deleteForce = true
	if err := runDelete(nil, []string{"feat-create"}); err != nil {
		t.Fatalf("runDelete failed: %v", err)
	}
	deleteForce = false
}
