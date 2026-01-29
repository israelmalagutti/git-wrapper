# git-wrapper Evidence Plan (Rubric v0.1.0)

Defines where evidence for rubric items is produced and how to regenerate it. Evidence should be reproducible from a commit SHA (no hand-assembled screenshots unless unavoidable).

## Evidence sources
### CI artifacts (preferred)
- Coverage: `make test-coverage` → `coverage.out` (and `coverage.html` if desired)
- Lint: `golangci-lint run --timeout=5m ./...` output (pinned version)
- Security: `golangci-lint run --timeout=5m --enable gosec ./...`, `govulncheck ./...`
- Supply-chain: `check_supply_chain` (via `bash gov-infra/verifiers/gov-verify-rubric.sh`)

### Deterministic in-repo artifacts
- Controls matrix: `gov-infra/planning/git-wrapper-controls-matrix.md`
- Rubric: `gov-infra/planning/git-wrapper-10of10-rubric.md`
- Roadmap: `gov-infra/planning/git-wrapper-10of10-roadmap.md`
- Evidence plan: `gov-infra/planning/git-wrapper-evidence-plan.md`
- Supply-chain allowlist: `gov-infra/planning/git-wrapper-supply-chain-allowlist.txt`
- Threat model: `gov-infra/planning/git-wrapper-threat-model.md`
- AI drift recovery: `gov-infra/planning/git-wrapper-ai-drift-recovery.md`
- Signature bundle (local certification): `gov-infra/signatures/gov-signature-bundle.json`

## Rubric-to-evidence map
Every rubric ID maps to exactly one verifier and one primary evidence location.

| Rubric ID | Primary evidence | Evidence path | How to refresh |
| --- | --- | --- | --- |
| QUA-1 | Unit test output | `gov-infra/evidence/QUA-1-output.log` | `make test` |
| QUA-2 | Integration test output | `gov-infra/evidence/QUA-2-output.log` | `Not implemented: add integration/contract test suite (or bump rubric and remove if truly N/A)` |
| QUA-3 | Coverage profile + summary | `gov-infra/evidence/QUA-3-output.log` | `make test-coverage` |
| CON-1 | Formatter diff list | `gov-infra/evidence/CON-1-output.log` | `bash -c 'files=$(gofmt -l .); if [ -n "$files" ]; then echo "$files"; exit 1; fi'` |
| CON-2 | Lint output | `gov-infra/evidence/CON-2-output.log` | `golangci-lint run --timeout=5m ./...` |
| CON-3 | Contract verification output | `gov-infra/evidence/CON-3-output.log` | `Not implemented: define exported-API contract parity checks (or bump rubric and remove if N/A)` |
| COM-1 | Module compile check | `gov-infra/evidence/COM-1-output.log` | `go test -run TestNonExistent ./...` |
| COM-2 | Toolchain pin verification | `gov-infra/evidence/COM-2-output.log` | `bash -c 'set -e; mod=$(awk "$1==\"go\"{print $2;exit}" go.mod); mod_minor=$(printf "%s" "$mod" | awk -F. "{print $1\".\"$2}"); echo "go.mod go version: ${mod} (minor=${mod_minor})"; explicit=$(grep -R --line-number -E "^[[:space:]]*go-version:[[:space:]]*[\x27\"]?[0-9]+\.[0-9]+(\.[0-9]+)?[\x27\"]?" .github/workflows 2>/dev/null | grep -v "go-version-file" || true); if [ -n "$explicit" ]; then echo "Explicit Go version pins found (must match go.mod at least major.minor):"; echo "$explicit"; while IFS= read -r line; do v=$(printf "%s" "$line" | sed -E "s/^.*go-version:[[:space:]]*//" | tr -d "\x27\"" | awk "{print $1}"); if [ "$v" != "$mod" ] && [ "$v" != "$mod_minor" ]; then echo "FAIL: workflow go-version pin $v does not match go.mod $mod (or $mod_minor)"; exit 1; fi; done <<< "$explicit"; fi; if grep -n "golangci-lint@latest" Makefile >/dev/null 2>&1; then echo "FAIL: Makefile uses golangci-lint@latest"; exit 1; fi; echo "OK"'` |
| COM-3 | Lint config validation | `gov-infra/evidence/COM-3-output.log` | `bash -c 'if ls .golangci.* >/dev/null 2>&1; then echo "golangci config present"; else echo "Not implemented: add .golangci.yml (explicit lint policy)"; exit 2; fi'` |
| COM-4 | Coverage threshold check | `gov-infra/evidence/COM-4-output.log` | `bash -c 'make test-coverage >/dev/null; pct=$(go tool cover -func=coverage.out | awk "/^total:/ {gsub(/%/,\"\",\$3); print \$3}"); echo "coverage=${pct}%"; awk -v p="$pct" "BEGIN{exit (p+0<90)?1:0}"'` |
| COM-5 | Security config validation | `gov-infra/evidence/COM-5-output.log` | `Not implemented: add a security scan config (gosec or equivalent) and validate it` |
| COM-6 | Logging standards check | `gov-infra/evidence/COM-6-output.log` | `Not implemented: add deterministic checks for redaction / safe logging (no secrets in logs)` |
| SEC-1 | SAST scan output | `gov-infra/evidence/SEC-1-output.log` | `golangci-lint run --timeout=5m --enable gosec ./...` |
| SEC-2 | Vulnerability scan output | `gov-infra/evidence/SEC-2-output.log` | `govulncheck ./...` |
| SEC-3 | Supply-chain verification | `gov-infra/evidence/SEC-3-output.log` | `check_supply_chain` |
| SEC-4 | Domain P0 regression tests | `gov-infra/evidence/SEC-4-output.log` | `Not implemented: add P0 regression tests for high-risk operations` |
| CMP-1 | Controls matrix exists | `gov-infra/planning/git-wrapper-controls-matrix.md` | File existence check |
| CMP-2 | Evidence plan exists | `gov-infra/planning/git-wrapper-evidence-plan.md` | File existence check |
| CMP-3 | Threat model exists | `gov-infra/planning/git-wrapper-threat-model.md` | File existence check |
| MAI-1 | File budget check | `gov-infra/evidence/MAI-1-output.log` | `bash -c 'set -e; max=600; bad=$(find . -type f -name "*.go" -not -path "./gov-infra/*" -not -path "./dist/*" -not -path "./bin/*" -not -path "./node_modules/*" -print0 | xargs -0 -I{} sh -c "n=$(wc -l < \"{}\"); if [ \"$n\" -gt ${max} ]; then echo \"${n} {}\"; fi" | sort -nr || true); if [ -n "$bad" ]; then echo "FAIL: oversized Go files (>${max} lines):"; echo "$bad"; exit 1; fi; echo "OK"'` |
| MAI-2 | Maintainability roadmap check | `gov-infra/evidence/MAI-2-output.log` | `bash -c 'set -e; f=gov-infra/planning/git-wrapper-10of10-roadmap.md; test -f "$f"; o="{"; c="}"; pat="${o}${o}[A-Z0-9_][A-Z0-9_]*${c}${c}"; if grep -nE "$pat" "$f"; then exit 1; fi; echo "OK"'` |
| MAI-3 | Singleton check | `gov-infra/evidence/MAI-3-output.log` | `Not implemented: add a deterministic duplicate-semantics check for critical abstractions` |
| MAI-4 | CI runs rubric verifier | `gov-infra/evidence/MAI-4-output.log` | CI config scan (built into verifier) |
| DOC-1 | Threat model present | `gov-infra/planning/git-wrapper-threat-model.md` | File existence check |
| DOC-2 | Evidence plan present | `gov-infra/planning/git-wrapper-evidence-plan.md` | File existence check |
| DOC-3 | Rubric + roadmap present | `gov-infra/planning/git-wrapper-10of10-rubric.md` | File existence check |
| DOC-4 | Doc integrity (tokens) | `gov-infra/evidence/DOC-4-output.log` | `bash -c 'set -e; o="{"; c="}"; pat="${o}${o}[A-Z0-9_][A-Z0-9_]*${c}${c}"; if grep -R --line-number -E "$pat" gov-infra; then exit 1; fi; echo "OK"'` |
| DOC-5 | Threat ↔ controls parity | `gov-infra/evidence/DOC-5-parity.log` | Parity check (built into verifier) |

## Rubric Report (Fixed Location)
The deterministic verifier (`gov-infra/verifiers/gov-verify-rubric.sh`) produces a machine-readable report at:
- `gov-infra/evidence/gov-rubric-report.json`

## Notes
- All evidence paths are relative to repo root and must live under `gov-infra/`.
- Treat evidence refresh as part of `gov validate`; CI should archive artifacts.
