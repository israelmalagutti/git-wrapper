# git-wrapper: 10/10 Rubric (Quality, Consistency, Completeness, Security, Compliance Readiness, Maintainability, Docs)

This rubric defines what “10/10” means and how category grades are computed. It is designed to prevent goalpost drift and
“green by dilution” by making scoring **versioned, measurable, and repeatable**.

## Versioning (no moving goalposts)
- **Rubric version:** `v0.1.0` (2026-01-28)
- **Comparability rule:** grades are comparable only within the same version.
- **Change rule:** bump the version + changelog entry for any rubric change (what changed + why).

### Changelog
- `v0.1.0`: Initial rubric scaffold for git-wrapper.

## Scoring (deterministic)
- Each category is scored **0–10**.
- Point weights sum to **10** per category.
- Requirements are **pass/fail** (either earn full points or 0).
- A category is **10/10 only if all requirements in that category pass**.

## Verification (commands + deterministic artifacts are the source of truth)
Every rubric item has exactly one verification mechanism:
- a command (`make ...`, `go test ...`, `bash scripts/...`), or
- a deterministic artifact check (required doc exists and matches an agreed format).

Enforcement rule (anti-drift):
- If an item’s verifier is a command/script, it only counts as passing once it runs in CI and produces evidence.

---

## Quality (QUA) — reliable, testable, change-friendly
| ID | Points | Requirement | How to verify |
| --- | ---: | --- | --- |
| QUA-1 | 4 | Unit tests stay green | `make test` |
| QUA-2 | 3 | Integration or contract tests stay green | `Not implemented: add integration/contract test suite (or bump rubric and remove if truly N/A)` |
| QUA-3 | 3 | Coverage ≥ 90% (no denominator games) | `make test-coverage` |

**10/10 definition:** QUA-1 through QUA-3 pass.

## Consistency (CON) — one way to do the important things
| ID | Points | Requirement | How to verify |
| --- | ---: | --- | --- |
| CON-1 | 3 | gofmt/formatter clean (no diffs) | `bash -c 'files=$(gofmt -l .); if [ -n "$files" ]; then echo "$files"; exit 1; fi'` |
| CON-2 | 5 | Lint/static analysis green (pinned version) | `golangci-lint run --timeout=5m ./...` |
| CON-3 | 2 | Public boundary contract parity (if applicable) | `Not implemented: define exported-API contract parity checks (or bump rubric and remove if N/A)` |

**10/10 definition:** CON-1 through CON-3 pass (or document why CON-3 is N/A and remove it with a version bump).

## Completeness (COM) — verify the verifiers (anti-drift)
| ID | Points | Requirement | How to verify |
| --- | ---: | --- | --- |
| COM-1 | 2 | All modules compile (no “mystery meat”) | `go test -run TestNonExistent ./...` |
| COM-2 | 2 | Toolchain pins align to repo (Go/lint/tool versions) | `bash -c 'set -e; mod=$(awk "$1==\"go\"{print $2;exit}" go.mod); mod_minor=$(printf "%s" "$mod" | awk -F. "{print $1\".\"$2}"); echo "go.mod go version: ${mod} (minor=${mod_minor})"; explicit=$(grep -R --line-number -E "^[[:space:]]*go-version:[[:space:]]*[\x27\"]?[0-9]+\.[0-9]+(\.[0-9]+)?[\x27\"]?" .github/workflows 2>/dev/null | grep -v "go-version-file" || true); if [ -n "$explicit" ]; then echo "Explicit Go version pins found (must match go.mod at least major.minor):"; echo "$explicit"; while IFS= read -r line; do v=$(printf "%s" "$line" | sed -E "s/^.*go-version:[[:space:]]*//" | tr -d "\x27\"" | awk "{print $1}"); if [ "$v" != "$mod" ] && [ "$v" != "$mod_minor" ]; then echo "FAIL: workflow go-version pin $v does not match go.mod $mod (or $mod_minor)"; exit 1; fi; done <<< "$explicit"; fi; if grep -n "golangci-lint@latest" Makefile >/dev/null 2>&1; then echo "FAIL: Makefile uses golangci-lint@latest"; exit 1; fi; echo "OK"'` |
| COM-3 | 2 | Lint config schema-valid (no silent skip) | `bash -c 'if ls .golangci.* >/dev/null 2>&1; then echo "golangci config present"; else echo "Not implemented: add .golangci.yml (explicit lint policy)"; exit 2; fi'` |
| COM-4 | 2 | Coverage threshold not diluted (≥ 90%) | `bash -c 'make test-coverage >/dev/null; pct=$(go tool cover -func=coverage.out | awk "/^total:/ {gsub(/%/,\"\",\$3); print \$3}"); echo "coverage=${pct}%"; awk -v p="$pct" "BEGIN{exit (p+0<90)?1:0}"'` |
| COM-5 | 1 | Security scan config not diluted (no excluded high-signal rules) | `Not implemented: add a security scan config (gosec or equivalent) and validate it` |
| COM-6 | 1 | Logging/operational standards enforced (if applicable) | `Not implemented: add deterministic checks for redaction / safe logging (no secrets in logs)` |

**10/10 definition:** COM-1 through COM-6 pass.

## Security (SEC) — abuse-resilient and reviewable
| ID | Points | Requirement | How to verify |
| --- | ---: | --- | --- |
| SEC-1 | 3 | Static security scan green (pinned version) | `golangci-lint run --timeout=5m --enable gosec ./...` |
| SEC-2 | 3 | Dependency vulnerability scan green | `govulncheck ./...` |
| SEC-3 | 2 | Supply-chain verification green | `check_supply_chain` (via `bash gov-infra/verifiers/gov-verify-rubric.sh`) |
| SEC-4 | 2 | Domain-specific P0 regression tests (e.g., destructive operations; redaction) | `Not implemented: add P0 regression tests for high-risk operations` |

**10/10 definition:** SEC-1 through SEC-4 pass.

## Compliance Readiness (CMP) — auditability and evidence
| ID | Points | Requirement | How to verify |
| --- | ---: | --- | --- |
| CMP-1 | 4 | Controls matrix exists and is current | File exists: `gov-infra/planning/git-wrapper-controls-matrix.md` |
| CMP-2 | 3 | Evidence plan exists and is reproducible | File exists: `gov-infra/planning/git-wrapper-evidence-plan.md` |
| CMP-3 | 3 | Threat model exists and is current | File exists: `gov-infra/planning/git-wrapper-threat-model.md` |

**10/10 definition:** CMP-1 through CMP-3 pass.

## Maintainability (MAI) — convergent codebase (recommended for AI-heavy repos)
| ID | Points | Requirement | How to verify |
| --- | ---: | --- | --- |
| MAI-1 | 3 | File-size/complexity budgets enforced | `bash -c 'set -e; max=600; bad=$(find . -type f -name "*.go" -not -path "./gov-infra/*" -not -path "./dist/*" -not -path "./bin/*" -not -path "./node_modules/*" -print0 | xargs -0 -I{} sh -c "n=$(wc -l < \"{}\"); if [ \"$n\" -gt ${max} ]; then echo \"${n} {}\"; fi" | sort -nr || true); if [ -n "$bad" ]; then echo "FAIL: oversized Go files (>${max} lines):"; echo "$bad"; exit 1; fi; echo "OK"'` |
| MAI-2 | 2 | Maintainability roadmap current | `bash -c 'set -e; f=gov-infra/planning/git-wrapper-10of10-roadmap.md; test -f "$f"; o="{"; c="}"; pat="${o}${o}[A-Z0-9_][A-Z0-9_]*${c}${c}"; if grep -nE "$pat" "$f"; then exit 1; fi; echo "OK"'` |
| MAI-3 | 2 | Canonical implementations (no duplicate semantics) | `Not implemented: add a deterministic duplicate-semantics check for critical abstractions` |
| MAI-4 | 3 | CI runs `bash gov-infra/verifiers/gov-verify-rubric.sh` and fails on non-PASS | Built-in (CI config scan) |

**10/10 definition:** MAI-1 through MAI-4 pass.

## Docs (DOC) — integrity and parity
| ID | Points | Requirement | How to verify |
| --- | ---: | --- | --- |
| DOC-1 | 2 | Threat model present | File exists: `gov-infra/planning/git-wrapper-threat-model.md` |
| DOC-2 | 2 | Evidence plan present | File exists: `gov-infra/planning/git-wrapper-evidence-plan.md` |
| DOC-3 | 2 | Rubric + roadmap present | File exists: `gov-infra/planning/git-wrapper-10of10-rubric.md` |
| DOC-4 | 2 | Doc integrity (links, version claims) | `bash -c 'set -e; o="{"; c="}"; pat="${o}${o}[A-Z0-9_][A-Z0-9_]*${c}${c}"; if grep -R --line-number -E "$pat" gov-infra; then exit 1; fi; echo "OK"'` |
| DOC-5 | 2 | Threat ↔ controls parity | Built-in parity check (verifier) |

**10/10 definition:** DOC-1 through DOC-5 pass.


## Maintaining 10/10 (recommended CI surface)
List the minimal command set CI must run (no `latest` tools; pinned versions only). Example:
```bash
bash -c 'files=$(gofmt -l .); if [ -n "$files" ]; then echo "$files"; exit 1; fi'

golangci-lint run --timeout=5m ./...

make test
make test-coverage

go test -run TestNonExistent ./...

bash -c 'set -e; mod=$(awk "$1==\"go\"{print $2;exit}" go.mod); mod_minor=$(printf "%s" "$mod" | awk -F. "{print $1\".\"$2}"); echo "go.mod go version: ${mod} (minor=${mod_minor})"; explicit=$(grep -R --line-number -E "^[[:space:]]*go-version:[[:space:]]*[\x27\"]?[0-9]+\.[0-9]+(\.[0-9]+)?[\x27\"]?" .github/workflows 2>/dev/null | grep -v "go-version-file" || true); if [ -n "$explicit" ]; then echo "Explicit Go version pins found (must match go.mod at least major.minor):"; echo "$explicit"; while IFS= read -r line; do v=$(printf "%s" "$line" | sed -E "s/^.*go-version:[[:space:]]*//" | tr -d "\x27\"" | awk "{print $1}"); if [ "$v" != "$mod" ] && [ "$v" != "$mod_minor" ]; then echo "FAIL: workflow go-version pin $v does not match go.mod $mod (or $mod_minor)"; exit 1; fi; done <<< "$explicit"; fi; if grep -n "golangci-lint@latest" Makefile >/dev/null 2>&1; then echo "FAIL: Makefile uses golangci-lint@latest"; exit 1; fi; echo "OK"'

bash -c 'make test-coverage >/dev/null; pct=$(go tool cover -func=coverage.out | awk "/^total:/ {gsub(/%/,\"\",\$3); print \$3}"); echo "coverage=${pct}%"; awk -v p="$pct" "BEGIN{exit (p+0<90)?1:0}"'

check_supply_chain
```
