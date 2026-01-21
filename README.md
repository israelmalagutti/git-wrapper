# gw - Git Wrapper

A fast, simple git stack management CLI tool for working with stacked diffs (stacked PRs).

## Features

- **Stack Management** - Create and manage parent-child relationships between branches
- **Smart Navigation** - Move up/down the stack, jump to top/bottom
- **Rebase Operations** - Automatic restacking when parent branches change
- **Interactive UI** - Prompts for branch selection, conflict resolution guidance
- **Platform Agnostic** - Works with any git hosting (GitHub, GitLab, etc.)

## Installation

### From Release (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/israelmalagutti/git-wrapper/main/scripts/install.sh | bash
```

Or specify a version:

```bash
GW_VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/israelmalagutti/git-wrapper/main/scripts/install.sh | bash
```

### From Source

```bash
git clone https://github.com/israelmalagutti/git-wrapper.git
cd git-wrapper
make install
```

### Manual Download

Download the appropriate binary from the [releases page](https://github.com/israelmalagutti/git-wrapper/releases) and add it to your PATH.

## Quick Start

```bash
# Initialize gw in your repository
gw init

# Create a new stacked branch
gw create feat-auth

# Make changes and commit
git add . && git commit -m "Add authentication"

# Create another branch on top
gw create feat-auth-ui

# View the stack
gw log

# Navigate the stack
gw up          # Move to child branch
gw down        # Move to parent branch
gw top         # Jump to top of stack
gw bottom      # Jump to trunk

# Restack after parent changes
gw stack restack
```

## Commands

### Core Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `gw init` | | Initialize gw in a repository |
| `gw create <name>` | | Create a new stacked branch |
| `gw track [branch]` | | Track an existing branch |
| `gw checkout <branch>` | `co`, `switch` | Switch to a branch |
| `gw log` | | Visualize the stack structure |
| `gw info` | | Show current branch details |

### Navigation

| Command | Alias | Description |
|---------|-------|-------------|
| `gw up [n]` | `u` | Move up toward leaves |
| `gw down [n]` | `dn` | Move down toward trunk |
| `gw top` | `t` | Jump to top of stack |
| `gw bottom` | `b` | Jump to trunk |
| `gw parent` | | Show parent branch |
| `gw children` | | Show child branches |

### Stack Operations

| Command | Alias | Description |
|---------|-------|-------------|
| `gw stack restack` | | Rebase stack to maintain relationships |
| `gw modify` | `m` | Amend commit and restack children |
| `gw move [target]` | `mv` | Move branch to different parent |
| `gw fold` | | Fold current branch into parent |
| `gw delete [branch]` | `rm` | Delete branch from stack |
| `gw split` | | Split branch into multiple branches |
| `gw sync` | | Sync metadata with git branches |

### Split Modes

```bash
gw split -c              # Split by selecting commits
gw split -u              # Interactive hunk selection
gw split -f "*.json"     # Split files matching pattern
gw split -n base         # Specify new branch name
```

## Development

### Prerequisites

- Go 1.21+
- Make

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Create release archives with checksums
make release
```

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Build binary with version injection |
| `make build-all` | Cross-platform builds (Linux/macOS/Windows) |
| `make release` | Build all + create archives + checksums |
| `make install` | Build and install to /usr/local/bin |
| `make uninstall` | Remove from /usr/local/bin |
| `make clean` | Remove build artifacts |
| `make test` | Run tests |
| `make test-coverage` | Run tests with HTML coverage report |
| `make lint` | Run golangci-lint |
| `make version` | Show version info |

### Versioning

Version is injected at build time from git tags:

```bash
# Shows commit hash if no tag
gw --version
# gw version a1b2c3d

# After tagging
git tag v0.1.0
make build
gw --version
# gw version v0.1.0
```

### Creating a Release

1. Update `CHANGELOG.md`
2. Commit changes
3. Create and push a tag:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
4. GitHub Actions will automatically build and publish the release

## Configuration

gw stores configuration in `.gw/config.json` at the repository root:

```json
{
  "trunk": "main",
  "version": "1.0.0"
}
```

Branch metadata is stored in `.gw/metadata.json`:

```json
{
  "branches": {
    "feat-auth": {
      "parent": "main"
    },
    "feat-auth-ui": {
      "parent": "feat-auth"
    }
  }
}
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
