# git-wrapper Threat Model (custom — v0.1.0)

This document enumerates the highest-risk threats for the in-scope system and assigns stable IDs (`THR-*`) that must map
to controls in `gov-infra/planning/git-wrapper-controls-matrix.md`.

## Scope (must be explicit)
- **System:** A Go CLI (`gw`) that wraps git to create/manage PR stacks and visualize the stack in the terminal.
- **In-scope data:** source code; local git repository metadata; branch stack metadata under `.gw/`; remote git URLs; optional GitHub auth context via `gh` (tokens managed by `gh`); CI logs and build artifacts.
- **Environments:** local developer machines, CI (GitHub Actions), and release build environment (tagged builds). “Prod-like” means a clean checkout in CI on `ubuntu-latest`.
- **Third parties:** Git (external binary), GitHub (remote hosting + Actions), `actions/*` (CI actions), `golangci-lint`, Go toolchain, (optional) `gh` CLI.
- **Out of scope:** security of the user’s workstation; git hosting provider security beyond what we can verify from this repo.
- **Assurance target:** audit-ready engineering evidence (repeatable checks + anti-drift gates), without claiming any framework certification.

## Assets and Trust Boundaries (high level)
- **Primary assets:** local repository integrity (commit graph + branches), correctness of stack relationships (`.gw/metadata.json`), user trust in CLI output, release binaries.
- **Trust boundaries:**
  - `gw` process ↔ external binaries (`git`, and optionally `gh`).
  - Local filesystem (repo) ↔ remote git provider.
  - CI runner ↔ GitHub Actions marketplace dependencies.
- **Entry points:** CLI arguments; branch names (user-controlled strings); git configuration; environment variables (e.g., `GW_DEBUG`); remote URLs; CI workflow definitions.

## Top Threats (stable IDs)
Threat IDs must be stable over time. When a new class of risk is discovered:
1) add a new `THR-*`,
2) add/adjust controls in the controls matrix,
3) update the rubric/roadmap if a new verifier is required.

| Threat ID | Title | What can go wrong | Primary controls (Control IDs) | Verification (gate) |
| --- | --- | --- | --- | --- |
| THR-1 | Destructive git operations or incorrect restack logic | `gw` issues the wrong git commands (or in the wrong order) causing data loss, unexpected rebases, or history rewrites. | QUA-1, QUA-3, COM-1 | `make test`, `make test-coverage`, `go test -run TestNonExistent ./...` |
| THR-2 | Incorrect branch parent/child relationships | Stack metadata becomes inconsistent with git branches, leading to broken navigation, wrong base branches, or incorrect PR stacking. | QUA-1, QUA-3, COM-1 | `make test`, `make test-coverage`, `go test -run TestNonExistent ./...` |
| THR-3 | Command injection / unsafe shelling-out | Unsanitized branch names or user inputs get interpolated into shell commands, causing unexpected command execution. | CON-2, COM-3, SEC-1 | `golangci-lint run --timeout=5m ./...`, lint config validation, `golangci-lint run --enable gosec ...` |
| THR-4 | Credential/token leakage via logs/config | Debug output or errors include credentials, remote URLs with embedded tokens, or other sensitive values. | COM-6 (planned), SEC-4 (planned) | Not implemented: add deterministic log/redaction tests and logging standards verifier |
| THR-5 | Dependency and CI supply-chain compromise | Malicious upstream dependency, compromised action, or installer script causes execution of attacker-controlled code. | SEC-2, SEC-3, COM-2 | `govulncheck ./...`, `check_supply_chain`, toolchain pin checks |
| THR-6 | Release integrity failure | Release binaries are built with inconsistent toolchains, from untrusted sources, or with tampered inputs. | COM-2, SEC-3 | Toolchain pin checks; Actions integrity pinning |

## Parity Rule (no “named threat without control”)
- Every `THR-*` listed above must appear at least once in the controls matrix “Threat IDs” column.
- The repo must have a deterministic parity check (used by `gov validate`) that fails if any threat is unmapped.

## Notes
- Prefer threats phrased as “failure modes” the repo can actually prevent or detect.
- If GitHub integration via `gh` is implemented, expand THR-4/THR-5 to cover auth context handling and PR metadata caching.
