# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `gw split` command with three modes: by-commit, by-hunk, by-file
- `gw up` / `gw down` commands for stack navigation
- `gw top` / `gw bottom` commands to jump to stack ends
- `gw move` command with `--source` and `--target` flags
- `gw fold` command to fold branch into parent
- `gw delete` command to delete branches from stack
- `gw modify` command for amending commits
- `gw stack restack` command for rebasing stacks
- `gw sync` command with cycle detection
- Graphite-style restack messaging
- Version injection from git tags
- Cross-platform builds (Linux, macOS, Windows)
- GitHub Actions CI/CD workflows
- Installation script for downloading releases

### Fixed
- Handle trunk branch properly in all commands
- Fix nil pointer in move command interactive mode
- Handle Ctrl+C cancellation gracefully

### Changed
- Silence usage/help output on errors (show only with `-h`)

## [0.1.0] - Initial Release

### Added
- `gw init` - Initialize gw in a repository
- `gw create` - Create stacked branches
- `gw track` - Track existing branches
- `gw checkout` - Smart branch switching with aliases
- `gw log` - Visualize stack structure
- `gw info` - Show branch details
- `gw parent` / `gw children` - Show relationships
- Configuration and metadata storage
- Interactive prompts with survey library
