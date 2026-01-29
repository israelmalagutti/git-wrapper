# git-wrapper Controls Matrix (custom — v0.1.0)

This matrix is the “requirements → controls → verifiers → evidence” backbone for git-wrapper. It is intentionally
engineering-focused: it does not claim compliance, but it makes security/quality assertions traceable and repeatable.

## Scope
- **System:** A Go CLI (`gw`) that wraps git to create/manage PR stacks and visualize the stack in the terminal.
- **In-scope data:** source code; local git repository metadata; branch stack metadata under `.gw/`; remote git URLs; optional GitHub auth context via `gh` (tokens managed by `gh`); CI logs and build artifacts.
- **Environments:** local developer machines, CI (GitHub Actions), and release build environment. “Prod-like” means a clean checkout in CI on `ubuntu-latest`.
- **Third parties:** Git (external binary), GitHub (remote hosting + Actions), `actions/*` (CI actions), `golangci-lint`, Go toolchain, (optional) `gh` CLI for future GitHub integration.
- **Out of scope:** security of the user’s workstation; git hosting provider security beyond what we can verify from this repo; network perimeter controls; organization-level SSO policy.
- **Assurance target:** audit-ready engineering evidence (repeatable checks + anti-drift gates), without claiming any framework certification.

## Threats (reference IDs)
- Threats are enumerated as stable IDs (`THR-*`) in `gov-infra/planning/git-wrapper-threat-model.md`.
- Each `THR-*` must map to ≥1 row in the controls table below (validated by a deterministic parity check).

## Status (evidence-driven)
If you track implementation status, treat it as evidence-driven:
- `unknown`: no verifier/evidence yet
- `partial`: some controls exist but coverage/evidence is incomplete
- `implemented`: verifier exists and evidence path is repeatable

## Engineering Controls (Threat → Control → Verifier → Evidence)
This table is the canonical mapping used by the rubric/roadmap/evidence plan.

| Area | Threat IDs | Control ID | Requirement | Control (what we implement) | Verification (command/gate) | Evidence (artifact/location) |
| --- | --- | --- | --- | --- | --- | --- |
| Quality | THR-1, THR-2 | QUA-1 | Unit tests prevent regressions | Unit tests for stack/branch logic and metadata persistence run in CI and locally. | `make test` | `gov-infra/evidence/QUA-1-output.log` |
| Quality | THR-1, THR-2 | QUA-3 | Coverage threshold is enforced (no dilution) | Coverage is measured from a clean checkout and compared to a fixed floor. | `make test-coverage` | `gov-infra/evidence/QUA-3-output.log` |
| Consistency | — | CON-1 | Formatting is clean (no diffs) | All `.go` files are gofmt clean. | `bash -c 'files=$(gofmt -l .); if [ -n "$files" ]; then echo "$files"; exit 1; fi'` | `gov-infra/evidence/CON-1-output.log` |
| Consistency | THR-3 | CON-2 | Lint/static analysis is enforced (pinned toolchain) | Lint runs with a pinned `golangci-lint` version and fails the build on findings. | `golangci-lint run --timeout=5m ./...` | `gov-infra/evidence/CON-2-output.log` |
| Completeness | THR-1, THR-2 | COM-1 | CI/build checks cover all packages | All packages compile in a clean environment (no “only main builds”). | `go test -run TestNonExistent ./...` | `gov-infra/evidence/COM-1-output.log` |
| Completeness | THR-5, THR-6 | COM-2 | Toolchain pins align to repo expectations | Workflows and repo pins align; avoid `latest` installs; fail on explicit Go version mismatch. | `bash -c 'set -e; mod=$(awk "$1==\"go\"{print $2;exit}" go.mod); mod_minor=$(printf "%s" "$mod" | awk -F. "{print $1\".\"$2}"); echo "go.mod go version: ${mod} (minor=${mod_minor})"; explicit=$(grep -R --line-number -E "^[[:space:]]*go-version:[[:space:]]*[\x27\"]?[0-9]+\.[0-9]+(\.[0-9]+)?[\x27\"]?" .github/workflows 2>/dev/null | grep -v "go-version-file" || true); if [ -n "$explicit" ]; then echo "Explicit Go version pins found (must match go.mod at least major.minor):"; echo "$explicit"; while IFS= read -r line; do v=$(printf "%s" "$line" | sed -E "s/^.*go-version:[[:space:]]*//" | tr -d "\x27\"" | awk "{print $1}"); if [ "$v" != "$mod" ] && [ "$v" != "$mod_minor" ]; then echo "FAIL: workflow go-version pin $v does not match go.mod $mod (or $mod_minor)"; exit 1; fi; done <<< "$explicit"; fi; if grep -n "golangci-lint@latest" Makefile >/dev/null 2>&1; then echo "FAIL: Makefile uses golangci-lint@latest"; exit 1; fi; echo "OK"'` | `gov-infra/evidence/COM-2-output.log` |
| Completeness | THR-3, THR-5 | COM-3 | Lint config schema-valid (no silent skip) | Lint policy is explicit via `.golangci.*` config and is validated. | `bash -c 'if ls .golangci.* >/dev/null 2>&1; then echo "golangci config present"; else echo "Not implemented: add .golangci.yml (explicit lint policy)"; exit 2; fi'` | `gov-infra/evidence/COM-3-output.log` |
| Security | THR-3 | SEC-1 | Baseline SAST stays green | Run a security-focused static analysis pass (e.g., `gosec` via golangci-lint) using pinned tooling. | `golangci-lint run --timeout=5m --enable gosec ./...` | `gov-infra/evidence/SEC-1-output.log` |
| Security | THR-5 | SEC-2 | Dependency vulnerability scan stays green | Run `govulncheck` using a pinned version and fail on findings. | `govulncheck ./...` | `gov-infra/evidence/SEC-2-output.log` |
| Security | THR-5, THR-6 | SEC-3 | Supply-chain verification green | Enforce integrity pins for CI actions and scan dependencies for common supply-chain risk signals. | `check_supply_chain` (via `bash gov-infra/verifiers/gov-verify-rubric.sh`) | `gov-infra/evidence/SEC-3-output.log` |
| Docs | THR-1, THR-2, THR-3, THR-4, THR-5, THR-6 | DOC-5 | Threat model ↔ controls parity (no unmapped threats) | Threat IDs in the threat model must appear in this matrix. | Built-in parity check (verifier) | `gov-infra/evidence/DOC-5-parity.log` |

> Add rows as needed for anti-drift (coverage/security config floors, multi-module health, CI rubric enforcement),
> for supply-chain/release integrity, and for domain-specific P0 gates.

## Framework Mapping (Optional)
No compliance framework is assumed for this repository. If one is adopted later, store only requirement IDs + short titles
and reference a KB path/env var (no licensed text in-repo).

## Notes
- Prefer deterministic verifiers (tests, static analysis, build assertions) over manual checklists.
- Treat this matrix as “source material”: the rubric/roadmap/evidence plan must stay consistent with Control IDs here.
