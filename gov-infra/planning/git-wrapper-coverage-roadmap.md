# git-wrapper: Coverage Roadmap (to 90%) (Rubric v0.1.0)

Goal: raise and maintain meaningful coverage to **≥ 90%** as measured by `make test-coverage`, without
reducing the measurement surface.

This exists as a standalone roadmap because coverage improvements are usually multi-PR efforts that need clear
intermediate milestones, guardrails, and repeatable measurement.

## Prerequisites
- Lint is green (or has a dedicated lint roadmap) so coverage work does not accumulate unreviewed lint debt.
- The coverage verifier is deterministic and uses a stable default threshold (no “lower it to pass” override).

## Current state
Snapshot (2026-01-28):
- Coverage gate: `make test-coverage`
- Current result: UNKNOWN vs threshold 90% (run verifier to capture)
- Measurement surface: `go test -coverprofile=coverage.out ./...` (repo-wide, no added excludes)

## Progress snapshots
- Baseline (2026-01-28): TODO: record total coverage from `go tool cover -func=coverage.out`
- After COV-1 (DATE): TODO
- After COV-2 (DATE): TODO

## Guardrails (no denominator games)
- Do not exclude additional production code from the coverage denominator to “hit the number”.
- Do not move logic into excluded areas (examples/tests/generated) to claim progress.
- If package/module floors are needed, add explicit target-based verification rather than weakening the global gate.

## How we measure
1) Generate/refresh the coverage artifact with the canonical command: `make test-coverage`
2) Verify the global floor: coverage >= 90% (see COM-4)
3) Re-run the full quality loop as a regression gate: `make test` and `golangci-lint run --timeout=5m ./...`

## Proposed milestones (incremental, reviewable)
- COV-1: remove “0% islands” (every in-scope package has tests)
- COV-2: broad floor (25%+ across in-scope packages)
- COV-3: meaningful safety net (50%+)
- COV-4: high confidence (70%+)
- COV-5: pre-finish (80%+)
- COV-6: finish line (≥ 90% and gate is green)

## Workstreams (target the highest-leverage paths first)
- Hotspots: CLI command logic in `cmd/` and stack operations in `internal/stack/`
- Common gap patterns: error paths, edge-case branch names, repo state transitions, serialization/deserialization of `.gw/metadata.json`

## Helpful commands
```bash
make test-coverage
make test
golangci-lint run --timeout=5m ./...
```
