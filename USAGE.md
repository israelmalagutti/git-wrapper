# Git Wrapper (gw) - Usage Guide

## Interactive Mode Commands

Many `gw` commands use interactive prompts to help you select branches, enter names, or make decisions. Here are the keyboard shortcuts available in interactive mode:

### Navigation & Selection
- **Arrow Keys (↑/↓)** - Navigate up and down through options
- **Tab** - Move to next option (in Select prompts)
- **Enter** - Confirm selection or submit input
- **Type to filter** - Start typing to filter options (in Select prompts)

### Vim Mode
- **ESC** - Toggle vim mode on/off
- **j** - Move down (when vim mode is enabled)
- **k** - Move up (when vim mode is enabled)

### Editing (Input prompts)
- **Backspace/Delete** - Remove characters
- **Ctrl+W** - Delete word
- **Ctrl+U** - Delete line

### Control
- **Ctrl+C** - Cancel and exit the current prompt

## Command Reference

### Initialization

#### `gw init`
Initialize git-wrapper in your repository. This sets up the configuration and identifies your trunk branch (usually `main` or `master`).

```bash
gw init
```

### Branch Management

#### `gw create [name]`
Create a new branch stacked on top of the current branch. The new branch is automatically tracked with the current branch as its parent.

```bash
# Interactive mode - prompts for branch name
gw create

# Direct mode - specify branch name
gw create feat-auth
```

If you have staged changes, you'll be prompted to commit them to the new branch.

#### `gw track`
Start tracking an existing branch. You'll be prompted to select a parent branch from your stack.

```bash
gw track
```

#### `gw checkout [options]`
Smart branch checkout with interactive selection. Shows stack context for each branch.

```bash
# Interactive mode - select from list of branches
gw co

# Quick checkout to trunk
gw co -t
gw co --trunk

# Show untracked branches in selection
gw co -u
gw co --show-untracked

# Only show branches in current stack
gw co -s
gw co --stack
```

**Aliases:** `co`, `checkout`, `switch`

### Visualization

#### `gw log [options]`
Display a visual tree representation of your branch stack.

```bash
# Compact view
gw log
gw log --short

# Detailed view with commit messages
gw log --long
```

Output format:
```
● *main (trunk) [fe9d15f]
├── feat-1 [a1cb412]
└── feat-2 [a1cb412]
```

Legend:
- `●` - Current branch (filled circle)
- `○` - Other branches (hollow circle)
- `*` - Indicator for current branch name
- `[hash]` - Commit SHA

#### `gw info`
Show detailed information about the current branch, including parent, children, depth in stack, and path to trunk.

```bash
gw info
```

#### `gw parent`
Show the parent branch of the current branch.

```bash
gw parent
```

#### `gw children`
Show all child branches of the current branch.

```bash
gw children
```

### Stack Maintenance

#### `gw stack restack`
Ensure each branch in the current stack is based on its parent, rebasing if necessary. This command recursively restacks all children branches.

```bash
# Restack current branch and all its children
gw stack restack

# Short aliases
gw stack r
gw stack fix
gw stack f
```

**What it does:**
- Checks if the current branch needs rebasing onto its parent
- Performs the rebase if the parent has moved forward
- Recursively restacks all children branches
- Handles conflicts interactively (prompts you to resolve and continue)

**When to use:**
- After making changes to a parent branch
- When trunk has moved forward and you want to update your stack
- To fix "out of sync" branches in your stack

**Aliases:** `r`, `fix`, `f`

#### `gw sync`
Clean up metadata and validate stack structure.

```bash
# Interactive cleanup
gw sync

# Force cleanup without prompts
gw sync -f
```

**What it does:**
- Removes metadata for branches that no longer exist in git
- Validates trunk branch has no parent
- Detects cycles in branch relationships
- Ensures stack structure is valid

#### `gw modify`
Modify the current branch by amending its commit or creating a new commit. Automatically restacks descendants.

```bash
# Amend current commit
gw modify

# Amend with message
gw modify -m "Updated commit message"

# Stage all changes and amend
gw modify -a

# Stage interactively and amend
gw modify -p

# Create new commit instead of amending
gw modify -c -m "New commit message"

# Short alias
gw m -a
```

**What it does:**
- Stages changes if requested (with `-a` or `-p`)
- Amends the current commit (default) or creates a new commit (with `-c`)
- Automatically restacks all children branches after the change
- Handles conflicts during restacking interactively

**Flags:**
- `-a, --all` - Stage all changes before committing
- `-p, --patch` - Interactively stage changes (prompts for each hunk)
- `-c, --commit` - Create a new commit instead of amending
- `-m, --message` - Specify commit message

**When to use:**
- When you want to make changes to the current branch's commit
- After code review feedback on a stacked branch
- To add forgotten changes to the current commit
- To split changes into a new commit

**Alias:** `m`

## Workflow Examples

### Creating a Stack of Features

```bash
# Start from trunk
git checkout main

# Create first feature branch
gw create feat-database
# ... make changes, commit ...

# Create second feature stacked on first
gw create feat-api
# ... make changes, commit ...

# Create third feature stacked on second
gw create feat-ui
# ... make changes, commit ...

# View your stack
gw log
```

### Navigating Your Stack

```bash
# View the stack
gw log

# Quickly jump to trunk
gw co -t

# Interactively select a branch to checkout
gw co

# Only see branches in current stack
gw co -s
```

### Tracking Existing Branches

```bash
# Checkout an existing branch
git checkout feat-existing

# Track it in gw
gw track
# Select parent branch when prompted

# Verify it's tracked
gw info
```

## Tips

1. **Use vim mode** for faster navigation if you're comfortable with j/k keys. Press ESC to toggle.
2. **Type to filter** in Select prompts to quickly find branches in large stacks.
3. **Use `gw co -t`** as a quick way to return to trunk from anywhere.
4. **Press Ctrl+C** anytime to safely cancel an operation.
5. **Check `gw log`** frequently to visualize your stack structure.
