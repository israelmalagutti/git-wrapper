# git-wrapper: 10/10 Roadmap (Rubric v0.1.0)

This roadmap maps milestones directly to rubric IDs with measurable acceptance criteria and verification commands.

## Current scorecard (Rubric v0.1.0)
Scoring note: a check is only treated as “passing” if it is both green **and** enforced by a trustworthy verifier
(pinned tooling, schema-valid configs, and no “green by dilution” shortcuts). Completeness failures invalidate “green by
drift”.

Status note: this pack does not assume the repo is currently compliant. Treat the following as the expected baseline
until you run the verifier and record evidence:
- Run: `bash gov-infra/verifiers/gov-verify-rubric.sh`
- Read: `gov-infra/evidence/gov-rubric-report.json`

| Category | Grade | Blocking rubric items |
| --- | ---: | --- |
| Quality | UNKNOWN | QUA-2, QUA-3 |
| Consistency | UNKNOWN | CON-2, CON-3 |
| Completeness | UNKNOWN | COM-2, COM-3, COM-4, COM-5, COM-6 |
| Security | UNKNOWN | SEC-2, SEC-3, SEC-4 |
| Compliance Readiness | 10/10 | (none; planning artifacts exist) |
| Maintainability | UNKNOWN | MAI-3, MAI-4 |
| Docs | UNKNOWN | DOC-4, DOC-5 |

Evidence (refresh whenever behavior changes):
- `make test`
- `Not implemented: add integration/contract test suite (or bump rubric and remove if truly N/A)`
- `make test-coverage`
- `bash -c 'files=$(gofmt -l .); if [ -n "$files" ]; then echo "$files"; exit 1; fi'`
- `golangci-lint run --timeout=5m ./...`
- `go test -run TestNonExistent ./...`
- `bash -c 'set -e; mod=$(awk "$1==\"go\"{print $2;exit}" go.mod); mod_minor=$(printf "%s" "$mod" | awk -F. "{print $1\".\"$2}"); echo "go.mod go version: ${mod} (minor=${mod_minor})"; explicit=$(grep -R --line-number -E "^[[:space:]]*go-version:[[:space:]]*[\x27\"]?[0-9]+\.[0-9]+(\.[0-9]+)?[\x27\"]?" .github/workflows 2>/dev/null | grep -v "go-version-file" || true); if [ -n "$explicit" ]; then echo "Explicit Go version pins found (must match go.mod at least major.minor):"; echo "$explicit"; while IFS= read -r line; do v=$(printf "%s" "$line" | sed -E "s/^.*go-version:[[:space:]]*//" | tr -d "\x27\"" | awk "{print $1}"); if [ "$v" != "$mod" ] && [ "$v" != "$mod_minor" ]; then echo "FAIL: workflow go-version pin $v does not match go.mod $mod (or $mod_minor)"; exit 1; fi; done <<< "$explicit"; fi; if grep -n "golangci-lint@latest" Makefile >/dev/null 2>&1; then echo "FAIL: Makefile uses golangci-lint@latest"; exit 1; fi; echo "OK"'`
- `bash -c 'make test-coverage >/dev/null; pct=$(go tool cover -func=coverage.out | awk "/^total:/ {gsub(/%/,\"\",\$3); print \$3}"); echo "coverage=${pct}%"; awk -v p="$pct" "BEGIN{exit (p+0<90)?1:0}"'`
- `Not implemented: add a security scan config (gosec or equivalent) and validate it`
- `golangci-lint run --timeout=5m --enable gosec ./...`
- `govulncheck ./...`
- `check_supply_chain`
- Built-in parity check (DOC-5)
- `bash -c 'set -e; o="{"; c="}"; pat="${o}${o}[A-Z0-9_][A-Z0-9_]*${c}${c}"; if grep -R --line-number -E "$pat" gov-infra; then exit 1; fi; echo "OK"'`

## Rubric-to-milestone mapping
| Rubric ID | Status | Milestone |
| --- | --- | --- |
| QUA-1 | UNKNOWN | M1.5 — Coverage/quality gates |
| QUA-2 | BLOCKED | M1.5 — Coverage/quality gates |
| QUA-3 | UNKNOWN | M1.5 — Coverage/quality gates |
| CON-1 | UNKNOWN | M1 — Make core lint/build loop reproducible |
| CON-2 | UNKNOWN | M1 — Make core lint/build loop reproducible |
| CON-3 | BLOCKED | M3+ — Domain/feature hardening |
| COM-1 | UNKNOWN | M2 — Enforce in CI |
| COM-2 | EXPECTED FAIL | M2 — Enforce in CI |
| COM-3 | BLOCKED | M1 — Make core lint/build loop reproducible |
| COM-4 | UNKNOWN | M1.5 — Coverage/quality gates |
| COM-5 | BLOCKED | M3+ — Domain/feature hardening |
| COM-6 | BLOCKED | M3+ — Domain/feature hardening |
| SEC-1 | UNKNOWN | M2 — Enforce in CI |
| SEC-2 | BLOCKED | M2 — Enforce in CI |
| SEC-3 | EXPECTED FAIL | M2 — Enforce in CI |
| SEC-4 | BLOCKED | M3+ — Domain/feature hardening |
| CMP-1 | PASS (doc present) | M0 — Freeze rubric + planning artifacts |
| CMP-2 | PASS (doc present) | M0 — Freeze rubric + planning artifacts |
| CMP-3 | PASS (doc present) | M0 — Freeze rubric + planning artifacts |
| MAI-1 | UNKNOWN | M3+ — Domain/feature hardening |
| MAI-2 | PASS (doc present) | M0 — Freeze rubric + planning artifacts |
| MAI-3 | BLOCKED | M3+ — Domain/feature hardening |
| MAI-4 | EXPECTED FAIL | M2 — Enforce in CI |
| DOC-1 | PASS (doc present) | M0 — Freeze rubric + planning artifacts |
| DOC-2 | PASS (doc present) | M0 — Freeze rubric + planning artifacts |
| DOC-3 | PASS (doc present) | M0 — Freeze rubric + planning artifacts |
| DOC-4 | UNKNOWN | M0 — Freeze rubric + planning artifacts |
| DOC-5 | UNKNOWN | M0 — Freeze rubric + planning artifacts |

## Workstream tracking docs (when blockers require a dedicated plan)
Large remediation workstreams usually need their own roadmaps so they can be executed in reviewable slices and keep the
main roadmap readable:
- Lint remediation: `gov-infra/planning/git-wrapper-lint-green-roadmap.md`
- Coverage remediation: `gov-infra/planning/git-wrapper-coverage-roadmap.md`
- Other blocker workstreams: `gov-infra/planning/git-wrapper-workstream-<name>-roadmap.md`

## Milestones (sequenced)
### M0 — Freeze rubric + planning artifacts
**Closes:** COM-3, DOC-1, DOC-2, DOC-3 (adjust for your rubric)
**Goal:** prevent goalpost drift by making the definition of “good” explicit and versioned.

**Acceptance criteria**
- Rubric exists and is versioned.
- Threat model exists and is owned.
- Evidence plan maps rubric IDs → verifiers → artifacts.

### M1 — Make core lint/build loop reproducible
**Closes:** CON-1, CON-2, COM-4
**Goal:** strict lint/format enforcement with pinned tools; no drift.

Tracking document: `gov-infra/planning/git-wrapper-lint-green-roadmap.md`

**Acceptance criteria**
- Formatter clean; lint green with schema-valid config; pinned tool versions; no blanket excludes.

### M1.5 — Coverage/quality gates (if applicable)
**Closes:** QUA-*
**Goal:** reach coverage floor (≥ 90%) without reducing scope; tests green.

Tracking document: `gov-infra/planning/git-wrapper-coverage-roadmap.md`

### M2 — Enforce in CI
**Closes:** COM-1, COM-2, COM-6, MAI-4, SEC-1..3
**Goal:** run the rubric surface in CI with pinned tooling; upload artifacts.

**Acceptance criteria**
- CI runs `bash gov-infra/verifiers/gov-verify-rubric.sh`.
- CI archives `gov-infra/evidence/` artifacts for review.

### M3+ — Domain/feature hardening
Add domain-specific milestones appropriate for git-wrapper:
- destructive-operation guardrails (THR-1)
- branch/metadata consistency property tests (THR-2)
- command construction hardening (THR-3)
- log redaction checks (THR-4)
- release integrity hardening (THR-6)

For each milestone, specify:
- **Goal**
- **Rubric IDs closed**
- **Acceptance criteria**
- **Suggested verification commands**
- **Evidence location**

