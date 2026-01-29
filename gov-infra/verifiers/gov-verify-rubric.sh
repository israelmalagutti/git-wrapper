#!/usr/bin/env bash
# GovTheory Rubric Verifier (Single Entrypoint)
# Generated from pack version: 2f9275c2707d
# Project: git-wrapper (git-wrapper)
#
# This script is the deterministic verifier entrypoint for gov.validate.
# It reads planning state from gov-infra/planning/, runs repo-specific check
# commands, writes evidence under gov-infra/evidence/, and emits a fixed JSON
# report at gov-infra/evidence/gov-rubric-report.json.
#
# Usage (from repo root; scripts may be non-executable by default):
#   bash gov-infra/verifiers/gov-verify-rubric.sh
#
# Exit codes:
#   0 - All rubric items PASS
#   1 - One or more rubric items FAIL or BLOCKED
#   2 - Script error (missing dependencies, invalid config, etc.)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
GOV_INFRA="${REPO_ROOT}/gov-infra"
PLANNING_DIR="${GOV_INFRA}/planning"
EVIDENCE_DIR="${GOV_INFRA}/evidence"
REPORT_PATH="${EVIDENCE_DIR}/gov-rubric-report.json"

# Always run checks from repo root so relative commands are stable.
cd "${REPO_ROOT}"

# Optional repo-local tools directory (to enforce pinned tool versions deterministically).
# Tools are installed here (never system-wide) and put first on PATH.
GOV_TOOLS_DIR="${GOV_INFRA}/.tools"
GOV_TOOLS_BIN="${GOV_TOOLS_DIR}/bin"
mkdir -p "${GOV_TOOLS_BIN}"
export PATH="${GOV_TOOLS_BIN}:${PATH}"

# Tool pins (optional; populated by gov.init when possible).
# If these remain unset, checks that depend on them should be marked BLOCKED (never "use whatever is installed").
PIN_GOLANGCI_LINT_VERSION="v1.64.8"
PIN_GOVULNCHECK_VERSION="Not implemented: pin govulncheck version (e.g., v1.1.4)"

# Optional feature flags (opt-in pack features).
# These must be explicitly enabled during gov.init; when unset, default is disabled.
FEATURE_OSS_RELEASE="false"

# Ensure evidence directory exists
mkdir -p "${EVIDENCE_DIR}"

# Clean previous run outputs to prevent stale evidence from being misattributed.
# Only remove files this verifier owns (do not wipe arbitrary user evidence).
rm -f \
  "${REPORT_PATH}" \
  "${EVIDENCE_DIR}/"*-output.log \
  "${EVIDENCE_DIR}/DOC-5-parity.log"

# Initialize report structure
REPORT_SCHEMA_VERSION=1
REPORT_TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
PASS_COUNT=0
FAIL_COUNT=0
BLOCKED_COUNT=0

# Results array (will be populated by run_check)
declare -a RESULTS=()

json_escape() {
  local s="$1"
  s="${s//\\/\\\\}"
  s="${s//\"/\\\"}"
  s="${s//$'\n'/\\n}"
  s="${s//$'\r'/\\r}"
  printf '%s' "$s"
}

record_result() {
  local id="$1"
  local category="$2"
  local status="$3"
  local message="$4"
  local evidence_path="$5"

  case "$status" in
    PASS) ((PASS_COUNT++)) || true ;;
    FAIL) ((FAIL_COUNT++)) || true ;;
    BLOCKED) ((BLOCKED_COUNT++)) || true ;;
    *) echo "Internal error: invalid status '${status}'" >&2; exit 2 ;;
  esac

  RESULTS+=(
    "{\"id\":\"$(json_escape "$id")\",\"category\":\"$(json_escape "$category")\",\"status\":\"$(json_escape "$status")\",\"message\":\"$(json_escape "$message")\",\"evidencePath\":\"$(json_escape "$evidence_path")\"}"
  )
}

is_unset_token() {
  # Treat as unset if empty, TODO placeholder, or a still-rendered template token.
  local v="$1"
  [[ -z "${v//[[:space:]]/}" ]] && return 0
  [[ "$v" == "Not implemented:"* ]] && return 0

  # Avoid hardcoding double-curly braces so doc-integrity greps can distinguish real template tokens from code.
  local o="{"
  [[ "$v" == "${o}${o}"* ]] && return 0

  return 1
}

normalize_feature_flags() {
  if is_unset_token "$FEATURE_OSS_RELEASE"; then
    FEATURE_OSS_RELEASE="false"
  fi
  FEATURE_OSS_RELEASE="$(printf '%s' "$FEATURE_OSS_RELEASE" | tr '[:upper:]' '[:lower:]')"
  case "$FEATURE_OSS_RELEASE" in
    true|false) ;;
    *) FEATURE_OSS_RELEASE="false" ;;
  esac
}

detect_golangci_lint_version_from_ci() {
  # Best-effort detection of the pinned golangci-lint version from GitHub Actions workflow config.
  # If multiple versions are found, fail closed.
  local wf_dir="${REPO_ROOT}/.github/workflows"
  [[ -d "$wf_dir" ]] || return 1

  local versions=""
  local file
  for file in "$wf_dir"/*.yml "$wf_dir"/*.yaml; do
    [[ -f "$file" ]] || continue

    local line
    while IFS= read -r line; do
      local start="${line%%:*}"
      [[ "$start" =~ ^[0-9]+$ ]] || continue

      local snippet
      snippet="$(sed -n "${start},$((start + 40))p" "$file" | grep -E '^[[:space:]]*version:[[:space:]]*v[0-9]+\.[0-9]+\.[0-9]+' | head -n 1 || true)"
      if [[ -n "$snippet" ]]; then
        versions+=$'\n'"$(printf '%s' "$snippet" | awk '{print $2}')"
      fi
    done < <(grep -n "golangci/golangci-lint-action@" "$file" 2>/dev/null || true)
  done

  versions="$(printf '%s' "$versions" | sed '/^$/d' | sort -u)"
  if [[ -z "$versions" ]]; then
    return 1
  fi
  if [[ "$(printf '%s\n' "$versions" | wc -l | tr -d ' ')" != "1" ]]; then
    return 1
  fi
  printf '%s' "$versions"
}

ensure_golangci_lint_pinned() {
  # Ensure golangci-lint is available at a pinned version (prefer repo-local tools dir).
  local v="$PIN_GOLANGCI_LINT_VERSION"
  if is_unset_token "$v"; then
    if v="$(detect_golangci_lint_version_from_ci)"; then
      :
    else
      echo "BLOCKED: golangci-lint version pin missing (set PIN_GOLANGCI_LINT_VERSION or pin in CI workflow)" >&2
      return 2
    fi
  fi
  if [[ "$v" != v* ]]; then
    v="v${v}"
  fi

  if ! command -v go >/dev/null 2>&1; then
    echo "BLOCKED: go toolchain not available to install golangci-lint ${v}" >&2
    return 2
  fi

  local want="${v#v}"
  if command -v golangci-lint >/dev/null 2>&1; then
    if golangci-lint --version 2>/dev/null | grep -q "$want"; then
      return 0
    fi
  fi

  echo "Installing golangci-lint ${v} into ${GOV_TOOLS_BIN}..." >&2
  if ! GOBIN="${GOV_TOOLS_BIN}" go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${v}"; then
    echo "BLOCKED: failed to install pinned golangci-lint ${v} (check network/toolchain)" >&2
    return 2
  fi

  if ! golangci-lint --version 2>/dev/null | grep -q "$want"; then
    echo "FAIL: installed golangci-lint does not report expected version ${v}" >&2
    golangci-lint --version 2>/dev/null || true
    return 1
  fi

  return 0
}

ensure_govulncheck_pinned() {
  local v="$PIN_GOVULNCHECK_VERSION"
  if is_unset_token "$v"; then
    echo "BLOCKED: govulncheck version pin missing (set PIN_GOVULNCHECK_VERSION)" >&2
    return 2
  fi
  if [[ "$v" != v* ]]; then
    v="v${v}"
  fi

  if ! command -v go >/dev/null 2>&1; then
    echo "BLOCKED: go toolchain not available to install govulncheck ${v}" >&2
    return 2
  fi

  if command -v govulncheck >/dev/null 2>&1; then
    if govulncheck -version 2>/dev/null | grep -q "govulncheck@${v}"; then
      return 0
    fi
  fi

  echo "Installing govulncheck ${v} into ${GOV_TOOLS_BIN}..." >&2
  if ! GOBIN="${GOV_TOOLS_BIN}" go install "golang.org/x/vuln/cmd/govulncheck@${v}"; then
    echo "BLOCKED: failed to install pinned govulncheck ${v} (check network/toolchain)" >&2
    return 2
  fi

  if ! govulncheck -version 2>/dev/null | grep -q "govulncheck@${v}"; then
    echo "FAIL: installed govulncheck does not report expected version ${v}" >&2
    govulncheck -version 2>/dev/null || true
    return 1
  fi

  return 0
}

allowlist_has_id() {
  local allowlist_path="$1"
  local id="$2"
  [[ -f "${allowlist_path}" ]] || return 1
  grep -Fqx -- "${id}" "${allowlist_path}"
}

sha256_12() {
  local s="$1"
  local hash=""
  if command -v sha256sum >/dev/null 2>&1; then
    hash="$(printf '%s' "${s}" | sha256sum | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    hash="$(printf '%s' "${s}" | shasum -a 256 | awk '{print $1}')"
  else
    echo "BLOCKED: sha256 tool missing (need sha256sum or shasum)" >&2
    return 2
  fi
  printf '%s' "${hash:0:12}"
  return 0
}

extract_go_mod_replaces() {
  local mod="${REPO_ROOT}/go.mod"
  [[ -f "${mod}" ]] || return 0

  awk '
    BEGIN { inblock=0 }
    $1 == "replace" && $2 == "(" { inblock=1; next }
    $1 == "replace" && $2 != "(" {
      $1=""; sub(/^[[:space:]]+/, ""); print; next
    }
    inblock && $1 == ")" { inblock=0; next }
    inblock { print; next }
  ' "${mod}"
}

scan_go_supply_chain() {
  # Scans Go module metadata for supply-chain risk signals.
  local allowlist_path="$1"

  if [[ ! -f "${REPO_ROOT}/go.mod" ]]; then
    echo "Go supply-chain scan: no go.mod detected; skipping."
    return 0
  fi

  local failures=0
  local allowlisted=0

  if [[ ! -f "${REPO_ROOT}/go.sum" ]]; then
    local id="GOV-SUPPLY:GO:MOD:rule=MISSING_GO_SUM"
    if allowlist_has_id "${allowlist_path}" "${id}"; then
      allowlisted=$((allowlisted + 1))
    else
      failures=$((failures + 1))
      echo "- ${id} file=go.sum"
    fi
  fi

  local known_malicious=(
    "github.com/boltdb-go/bolt"
    "github.com/gin-goinc"
    "github.com/go-chi/chi/v6"
  )

  local mod
  for mod in "${known_malicious[@]}"; do
    if grep -Fq -- "${mod}" "${REPO_ROOT}/go.mod" 2>/dev/null || ( [[ -f "${REPO_ROOT}/go.sum" ]] && grep -Fq -- "${mod}" "${REPO_ROOT}/go.sum" 2>/dev/null ); then
      local id="GOV-SUPPLY:GO:MOD:rule=KNOWN_MALICIOUS_MODULE:module=${mod}"
      if allowlist_has_id "${allowlist_path}" "${id}"; then
        allowlisted=$((allowlisted + 1))
      else
        failures=$((failures + 1))
        echo "- ${id}"
      fi
    fi
  done

  local line
  while IFS= read -r line; do
    [[ -z "${line//[[:space:]]/}" ]] && continue
    [[ "${line}" == "//"* ]] && continue
    [[ "${line}" == *"=>"* ]] || continue

    local left="${line%%=>*}"
    local right="${line#*=>}"
    left="$(printf '%s' "${left}" | xargs)"
    right="$(printf '%s' "${right}" | xargs)"
    [[ -z "${left}" || -z "${right}" ]] && continue

    local from_mod=""
    local from_ver=""
    local to_mod=""
    local to_ver=""
    from_mod="$(printf '%s' "${left}" | awk '{print $1}')"
    from_ver="$(printf '%s' "${left}" | awk '{print $2}')"
    to_mod="$(printf '%s' "${right}" | awk '{print $1}')"
    to_ver="$(printf '%s' "${right}" | awk '{print $2}')"
    [[ -z "${from_mod}" || -z "${to_mod}" ]] && continue

    # Local replace targets are common in multi-module repos; ignore.
    if [[ "${to_mod}" == ./* || "${to_mod}" == ../* || "${to_mod}" == /* ]]; then
      continue
    fi

    local from="${from_mod}@${from_ver:-_}"
    local to="${to_mod}@${to_ver:-_}"
    local id="GOV-SUPPLY:GO:REPLACE:rule=REMOTE_REPLACE:from=${from}:to=${to}"
    if allowlist_has_id "${allowlist_path}" "${id}"; then
      allowlisted=$((allowlisted + 1))
    else
      failures=$((failures + 1))
      echo "- ${id} detail=$(printf '%s' "${line}" | tr -d '\r')"
    fi
  done < <(extract_go_mod_replaces)

  echo "Supply-chain scan (Go): findings=${failures} allowlisted=${allowlisted}"

  if [[ "${failures}" -ne 0 ]]; then
    return 1
  fi
  return 0
}

scan_python_supply_chain() {
  # Scans Python dependency/config files for supply-chain risk signals.
  local allowlist_path="$1"

  local -a files=()
  while IFS= read -r f; do
    files+=("$f")
  done < <(
    find "${REPO_ROOT}" -maxdepth 6 -type f \( \
      -name 'requirements*.txt' -o \
      -name 'constraints*.txt' -o \
      -name 'Pipfile' -o \
      -name 'Pipfile.lock' -o \
      -name 'poetry.lock' -o \
      -name 'pdm.lock' -o \
      -name 'uv.lock' -o \
      -name 'pyproject.toml' \
    \) \
    -not -path '*/node_modules/*' \
    -not -path '*/.git/*' \
    -not -path '*/.venv/*' \
    -not -path '*/venv/*' \
    -not -path '*/__pycache__/*' \
    2>/dev/null | LC_ALL=C sort
  )

  if [[ "${#files[@]}" -eq 0 ]]; then
    echo "Python supply-chain scan: no Python dependency files detected; skipping."
    return 0
  fi

  local known_malicious=(
    "python3-dateutil"
    "jeilyfish"
    "python-binance"
    "request"
    "urllib"
    "djanga"
    "coloursama"
    "larpexodus"
    "graphalgo"
    "acloud-client"
    "tcloud-python-test"
  )

  local failures=0
  local allowlisted=0
  local file_count=0

  local f
  for f in "${files[@]}"; do
    file_count=$((file_count + 1))
    local rel="${f#${REPO_ROOT}/}"

    local line
    while IFS= read -r line || [[ -n "${line}" ]]; do
      local raw="${line}"
      raw="${raw//$'\r'/}"
      local trimmed
      trimmed="$(printf '%s' "${raw}" | sed -E 's/[[:space:]]+/ /g; s/^ +//; s/ +$//')"
      [[ -z "${trimmed}" ]] && continue

      local lower
      lower="$(printf '%s' "${trimmed}" | tr '[:upper:]' '[:lower:]')"

      local rule=""

      # Known malicious packages (typosquats / compromised).
      local pkg
      for pkg in "${known_malicious[@]}"; do
        if [[ "${lower}" == *"${pkg}"* ]]; then
          rule="KNOWN_MALICIOUS_PACKAGE"
          break
        fi
      done

      # Dependency sources that bypass standard indexes (VCS / direct URL).
      if [[ -z "${rule}" ]]; then
        if [[ "${lower}" == *"git+https://"* || "${lower}" == *"git+http://"* || "${lower}" == *"git+ssh://"* || "${lower}" == *"hg+http"* || "${lower}" == *"svn+http"* || "${lower}" == *"bzr+http"* ]]; then
          rule="VCS_OR_URL_DEP"
        elif [[ "${lower}" == *" @ https://"* || "${lower}" == *" @ http://"* || "${lower}" == *" @ file://"* || "${lower}" == *" @ ssh://"* ]]; then
          rule="VCS_OR_URL_DEP"
        elif [[ "${lower}" == *"git = \""* && ( "${lower}" == *"http://"* || "${lower}" == *"https://"* || "${lower}" == *"ssh://"* ) ]]; then
          rule="VCS_OR_URL_DEP"
        elif [[ "${lower}" == *"\"git\":"* && ( "${lower}" == *"http://"* || "${lower}" == *"https://"* || "${lower}" == *"ssh://"* ) ]]; then
          rule="VCS_OR_URL_DEP"
        fi
      fi

      # Custom indexes and trusted hosts (higher supply-chain risk).
      if [[ -z "${rule}" ]]; then
        if [[ "${lower}" == *"--index-url"* || "${lower}" == *"--extra-index-url"* || "${lower}" == *"--find-links"* ]] || [[ "${lower}" =~ (^|[[:space:]])-f([[:space:]]|$) ]]; then
          rule="CUSTOM_INDEX"
        elif [[ "${lower}" == *"--trusted-host"* ]]; then
          rule="TRUSTED_HOST"
        elif [[ "${lower}" == "-e "* || "${lower}" == "--editable "* ]]; then
          rule="EDITABLE_INSTALL"
        fi
      fi

      [[ -z "${rule}" ]] && continue

      local h=""
      h="$(sha256_12 "${rel}|${rule}|${trimmed}")" || return $?
      local id="GOV-SUPPLY:PYTHON:LINE:file=${rel}:rule=${rule}:sha256=${h}"

      if allowlist_has_id "${allowlist_path}" "${id}"; then
        allowlisted=$((allowlisted + 1))
      else
        failures=$((failures + 1))
        echo "- ${id} detail=${trimmed:0:200}"
      fi
    done < "${f}"
  done

  echo "Supply-chain scan (Python): files=${file_count} findings=${failures} allowlisted=${allowlisted}"

  if [[ "${failures}" -ne 0 ]]; then
    return 1
  fi
  return 0
}

check_supply_chain_actions_pinned() {
  # Enforces integrity pinning for GitHub Actions (reject floating tags like @v4).
  # If the repo doesn't use GitHub Actions, this is a no-op.
  local wf_dir="${REPO_ROOT}/.github/workflows"
  if [[ ! -d "${wf_dir}" ]]; then
    echo "GitHub Actions pin check: no workflows detected; skipping."
    return 0
  fi

  # Fail if any workflow uses floating tags like @v2/@v4 (integrity pinning requirement).
  local matches=""
  matches="$(grep -R --include='*.yml' --include='*.yaml' -nE '^[[:space:]]*uses:[[:space:]].*@v[0-9]+' "${wf_dir}" 2>/dev/null || true)"
  if [[ -n "${matches}" ]]; then
    echo "FAIL: unpinned GitHub Action detected (uses @vN; pin by commit SHA)"
    echo "${matches}"
    return 1
  fi

  echo "GitHub Actions pin check: PASS (no uses @vN detected)"
  return 0
}

check_supply_chain() {
  # SEC-3: Supply-chain verification gate.
  # - Enforces GitHub Actions SHA pinning (no uses: ...@vN).
  # - Scans Go and Python dependency metadata for common supply-chain risk signals.

  local allowlist="${PLANNING_DIR}/git-wrapper-supply-chain-allowlist.txt"
  if [[ -f "${allowlist}" ]]; then
    echo "Supply-chain allowlist: ${allowlist}"
  else
    echo "Supply-chain allowlist: missing (treated as empty): ${allowlist}"
  fi

  local fail=0
  local blocked=0

  set +e
  check_supply_chain_actions_pinned
  local ec_actions=$?
  set -e
  if [[ $ec_actions -ne 0 ]]; then
    fail=1
  fi

  set +e
  scan_go_supply_chain "${allowlist}"
  local ec_go=$?
  set -e
  if [[ $ec_go -eq 2 ]]; then
    blocked=1
  elif [[ $ec_go -ne 0 ]]; then
    fail=1
  fi

  set +e
  scan_python_supply_chain "${allowlist}"
  local ec_py=$?
  set -e
  if [[ $ec_py -eq 2 ]]; then
    blocked=1
  elif [[ $ec_py -ne 0 ]]; then
    fail=1
  fi

  if [[ "${fail}" -ne 0 ]]; then
    return 1
  fi
  if [[ "${blocked}" -ne 0 ]]; then
    return 2
  fi
  return 0
}

prepare_check_env() {
  # Optional preflight to enforce pinned tools for known Go gates.
  local id="$1"
  local cmd="$2"

  # Only attempt Go tool bootstraps if this appears to be a Go module.
  if [[ ! -f "${REPO_ROOT}/go.mod" ]]; then
    return 0
  fi

  case "$id" in
    CON-2|COM-3|SEC-1)
      if [[ "$cmd" == *"golangci-lint"* ]]; then
        ensure_golangci_lint_pinned
      fi
      ;;
    SEC-2)
      if [[ "$cmd" == *"govulncheck"* ]]; then
        ensure_govulncheck_pinned
      fi
      ;;
    *) return 0 ;;
  esac
}

# Helper: run a single check and record result
# Usage: run_check <rubric_id> <category> <command>
run_check() {
  local id="$1"
  local category="$2"
  local cmd="$3"

  local output_file="${EVIDENCE_DIR}/${id}-output.log"

  local o="{" 
  if [[ -z "${cmd//[[:space:]]/}" ]] || [[ "${cmd}" == "Not implemented:"* ]] || [[ "${cmd}" == "${o}${o}CMD_"* ]]; then
    printf '%s\n' "Verifier command not configured: ${cmd}" > "${output_file}"
    record_result "$id" "$category" "BLOCKED" "Verifier command not configured" "$output_file"
    return 0
  fi

  set +e
  (
    set -euo pipefail
    prepare_check_env "$id" "$cmd"
    eval "${cmd}"
  ) >"${output_file}" 2>&1
  local ec=$?
  set -e

  if [[ $ec -eq 0 ]]; then
    record_result "$id" "$category" "PASS" "Command succeeded" "$output_file"
  elif [[ $ec -eq 2 || $ec -eq 126 || $ec -eq 127 ]]; then
    record_result "$id" "$category" "BLOCKED" "Command reported BLOCKED (exit code ${ec})" "$output_file"
  else
    record_result "$id" "$category" "FAIL" "Command failed with exit code ${ec}" "$output_file"
  fi
}

# Helper: check if a file exists (for doc verification)
check_file_exists() {
  local id="$1"
  local category="$2"
  local file_path="$3"

  if [[ -f "${file_path}" ]]; then
    record_result "$id" "$category" "PASS" "File exists" "$file_path"
  else
    record_result "$id" "$category" "FAIL" "Required file missing" "$file_path"
  fi
}

# Helper: check threat/controls parity
check_parity() {
  local threat_model="${PLANNING_DIR}/git-wrapper-threat-model.md"
  local controls_matrix="${PLANNING_DIR}/git-wrapper-controls-matrix.md"
  local evidence_path="${EVIDENCE_DIR}/DOC-5-parity.log"

  if [[ ! -f "${threat_model}" ]] || [[ ! -f "${controls_matrix}" ]]; then
    printf '%s\n' "Threat model or controls matrix missing" > "${evidence_path}"
    record_result "DOC-5" "Docs" "BLOCKED" "Threat model or controls matrix missing" "${evidence_path}"
  else
    local threat_ids
    threat_ids=$(grep -oE 'THR-[0-9]+' "${threat_model}" | sort -u || true)

    local missing=""
    for thr_id in ${threat_ids}; do
      if ! grep -q "${thr_id}" "${controls_matrix}"; then
        missing="${missing} ${thr_id}"
      fi
    done

    echo "Threat IDs found: ${threat_ids:-none}" > "${evidence_path}"
    echo "Missing from controls:${missing:-none}" >> "${evidence_path}"

    if [[ -z "${missing}" ]]; then
      record_result "DOC-5" "Docs" "PASS" "All threat IDs mapped in controls matrix" "${evidence_path}"
    else
      record_result "DOC-5" "Docs" "FAIL" "Unmapped threats:${missing}" "${evidence_path}"
    fi
  fi
}

check_mai_ci_rubric_enforced() {
  # MAI-4: verifies that CI runs the deterministic rubric verifier.
  local found_ci="false"
  local found_hook="false"

  if [[ -d "${REPO_ROOT}/.github/workflows" ]]; then
    found_ci="true"
    local wf
    for wf in "${REPO_ROOT}/.github/workflows/"*.yml "${REPO_ROOT}/.github/workflows/"*.yaml; do
      [[ -f "${wf}" ]] || continue
      if grep -q 'gov-verify-rubric\.sh' "${wf}"; then
        found_hook="true"
        echo "Found gov-verify-rubric.sh invocation in: ${wf#${REPO_ROOT}/}"
        break
      fi
    done
  fi

  if [[ "${found_ci}" != "true" ]]; then
    echo "MAI-4: FAIL (no CI configuration detected)"
    echo "Add CI that runs: bash gov-infra/verifiers/gov-verify-rubric.sh"
    return 1
  fi
  if [[ "${found_hook}" != "true" ]]; then
    echo "MAI-4: FAIL (CI configuration detected, but no job runs gov-verify-rubric.sh)"
    echo "Ensure CI runs: bash gov-infra/verifiers/gov-verify-rubric.sh"
    return 1
  fi

  echo "MAI-4: PASS"
  return 0
}

echo "=== GovTheory Rubric Verifier ==="
echo "Project: git-wrapper"
echo "Timestamp: ${REPORT_TIMESTAMP}"
echo ""

normalize_feature_flags

# Commands are intentionally centralized here so the rubric docs and verifier stay aligned.
CMD_UNIT=$(cat <<'__GOV_CMD_UNIT__'
make test
__GOV_CMD_UNIT__
)
CMD_INTEGRATION=$(cat <<'__GOV_CMD_INTEGRATION__'
Not implemented: add integration/contract test suite (or bump rubric and remove if truly N/A)
__GOV_CMD_INTEGRATION__
)
CMD_COVERAGE=$(cat <<'__GOV_CMD_COVERAGE__'
make test-coverage
__GOV_CMD_COVERAGE__
)

CMD_FMT=$(cat <<'__GOV_CMD_FMT__'
bash -c 'files=$(gofmt -l .); if [ -n "$files" ]; then echo "$files"; exit 1; fi'
__GOV_CMD_FMT__
)
CMD_LINT=$(cat <<'__GOV_CMD_LINT__'
golangci-lint run --timeout=5m ./...
__GOV_CMD_LINT__
)
CMD_CONTRACT=$(cat <<'__GOV_CMD_CONTRACT__'
Not implemented: define exported-API contract parity checks (or bump rubric and remove if N/A)
__GOV_CMD_CONTRACT__
)

CMD_MODULES=$(cat <<'__GOV_CMD_MODULES__'
go test -run TestNonExistent ./...
__GOV_CMD_MODULES__
)
CMD_TOOLCHAIN=$(cat <<'__GOV_CMD_TOOLCHAIN__'
bash -c 'set -e; mod=$(awk "$1==\"go\"{print $2;exit}" go.mod); mod_minor=$(printf "%s" "$mod" | awk -F. "{print $1\".\"$2}"); echo "go.mod go version: ${mod} (minor=${mod_minor})"; explicit=$(grep -R --line-number -E "^[[:space:]]*go-version:[[:space:]]*[\x27\"]?[0-9]+\.[0-9]+(\.[0-9]+)?[\x27\"]?" .github/workflows 2>/dev/null | grep -v "go-version-file" || true); if [ -n "$explicit" ]; then echo "Explicit Go version pins found (must match go.mod at least major.minor):"; echo "$explicit"; while IFS= read -r line; do v=$(printf "%s" "$line" | sed -E "s/^.*go-version:[[:space:]]*//" | tr -d "\x27\"" | awk "{print $1}"); if [ "$v" != "$mod" ] && [ "$v" != "$mod_minor" ]; then echo "FAIL: workflow go-version pin $v does not match go.mod $mod (or $mod_minor)"; exit 1; fi; done <<< "$explicit"; fi; if grep -n "golangci-lint@latest" Makefile >/dev/null 2>&1; then echo "FAIL: Makefile uses golangci-lint@latest"; exit 1; fi; echo "OK"'
__GOV_CMD_TOOLCHAIN__
)
CMD_LINT_CONFIG=$(cat <<'__GOV_CMD_LINT_CONFIG__'
bash -c 'if ls .golangci.* >/dev/null 2>&1; then echo "golangci config present"; else echo "Not implemented: add .golangci.yml (explicit lint policy)"; exit 2; fi'
__GOV_CMD_LINT_CONFIG__
)
CMD_COV_THRESHOLD=$(cat <<'__GOV_CMD_COV_THRESHOLD__'
bash -c 'make test-coverage >/dev/null; pct=$(go tool cover -func=coverage.out | awk "/^total:/ {gsub(/%/,\"\",\$3); print \$3}"); echo "coverage=${pct}%"; awk -v p="$pct" "BEGIN{exit (p+0<90)?1:0}"'
__GOV_CMD_COV_THRESHOLD__
)
CMD_SEC_CONFIG=$(cat <<'__GOV_CMD_SEC_CONFIG__'
Not implemented: add a security scan config (gosec or equivalent) and validate it
__GOV_CMD_SEC_CONFIG__
)
CMD_LOGGING=$(cat <<'__GOV_CMD_LOGGING__'
Not implemented: add deterministic checks for redaction / safe logging (no secrets in logs)
__GOV_CMD_LOGGING__
)

CMD_SAST=$(cat <<'__GOV_CMD_SAST__'
golangci-lint run --timeout=5m --enable gosec ./...
__GOV_CMD_SAST__
)
CMD_VULN=$(cat <<'__GOV_CMD_VULN__'
govulncheck ./...
__GOV_CMD_VULN__
)
CMD_SUPPLY=$(cat <<'__GOV_CMD_SUPPLY__'
check_supply_chain
__GOV_CMD_SUPPLY__
)
CMD_P0=$(cat <<'__GOV_CMD_P0__'
Not implemented: add P0 regression tests for high-risk operations
__GOV_CMD_P0__
)

CMD_FILE_BUDGET=$(cat <<'__GOV_CMD_FILE_BUDGET__'
bash -c 'set -e; max=600; bad=$(find . -type f -name "*.go" -not -path "./gov-infra/*" -not -path "./dist/*" -not -path "./bin/*" -not -path "./node_modules/*" -print0 | xargs -0 -I{} sh -c "n=$(wc -l < \"{}\"); if [ \"$n\" -gt ${max} ]; then echo \"${n} {}\"; fi" | sort -nr || true); if [ -n "$bad" ]; then echo "FAIL: oversized Go files (>${max} lines):"; echo "$bad"; exit 1; fi; echo "OK"'
__GOV_CMD_FILE_BUDGET__
)
CMD_MAINTAINABILITY=$(cat <<'__GOV_CMD_MAINTAINABILITY__'
bash -c 'set -e; f=gov-infra/planning/git-wrapper-10of10-roadmap.md; test -f "$f"; o="{"; c="}"; pat="${o}${o}[A-Z0-9_][A-Z0-9_]*${c}${c}"; if grep -nE "$pat" "$f"; then exit 1; fi; echo "OK"'
__GOV_CMD_MAINTAINABILITY__
)
CMD_SINGLETON=$(cat <<'__GOV_CMD_SINGLETON__'
Not implemented: add a deterministic duplicate-semantics check for critical abstractions
__GOV_CMD_SINGLETON__
)

CMD_DOC_INTEGRITY=$(cat <<'__GOV_CMD_DOC_INTEGRITY__'
bash -c 'set -e; o="{"; c="}"; pat="${o}${o}[A-Z0-9_][A-Z0-9_]*${c}${c}"; if grep -R --line-number -E "$pat" gov-infra; then exit 1; fi; echo "OK"'
__GOV_CMD_DOC_INTEGRITY__
)

CMD_CI_ENFORCED="check_mai_ci_rubric_enforced"

# === Quality (QUA) ===
run_check "QUA-1" "Quality" "$CMD_UNIT"
run_check "QUA-2" "Quality" "$CMD_INTEGRATION"
run_check "QUA-3" "Quality" "$CMD_COVERAGE"

# === Consistency (CON) ===
run_check "CON-1" "Consistency" "$CMD_FMT"
run_check "CON-2" "Consistency" "$CMD_LINT"
run_check "CON-3" "Consistency" "$CMD_CONTRACT"

# === Completeness (COM) ===
run_check "COM-1" "Completeness" "$CMD_MODULES"
run_check "COM-2" "Completeness" "$CMD_TOOLCHAIN"
run_check "COM-3" "Completeness" "$CMD_LINT_CONFIG"
run_check "COM-4" "Completeness" "$CMD_COV_THRESHOLD"
run_check "COM-5" "Completeness" "$CMD_SEC_CONFIG"
run_check "COM-6" "Completeness" "$CMD_LOGGING"

# === Security (SEC) ===
run_check "SEC-1" "Security" "$CMD_SAST"
run_check "SEC-2" "Security" "$CMD_VULN"
run_check "SEC-3" "Security" "$CMD_SUPPLY"
run_check "SEC-4" "Security" "$CMD_P0"

# === Compliance Readiness (CMP) ===
check_file_exists "CMP-1" "Compliance" "${PLANNING_DIR}/git-wrapper-controls-matrix.md"
check_file_exists "CMP-2" "Compliance" "${PLANNING_DIR}/git-wrapper-evidence-plan.md"
check_file_exists "CMP-3" "Compliance" "${PLANNING_DIR}/git-wrapper-threat-model.md"

# === Maintainability (MAI) ===
run_check "MAI-1" "Maintainability" "$CMD_FILE_BUDGET"
run_check "MAI-2" "Maintainability" "$CMD_MAINTAINABILITY"
run_check "MAI-3" "Maintainability" "$CMD_SINGLETON"
run_check "MAI-4" "Maintainability" "$CMD_CI_ENFORCED"

# === Docs (DOC) ===
check_file_exists "DOC-1" "Docs" "${PLANNING_DIR}/git-wrapper-threat-model.md"
check_file_exists "DOC-2" "Docs" "${PLANNING_DIR}/git-wrapper-evidence-plan.md"
check_file_exists "DOC-3" "Docs" "${PLANNING_DIR}/git-wrapper-10of10-rubric.md"
run_check "DOC-4" "Docs" "$CMD_DOC_INTEGRITY"
check_parity

# === Generate Report ===
echo ""
echo "=== Generating Report ==="

RESULTS_JSON=$(printf "%s," "${RESULTS[@]}")
RESULTS_JSON="[${RESULTS_JSON%,}]"

OVERALL_STATUS="PASS"
if [[ ${FAIL_COUNT} -gt 0 ]]; then
  OVERALL_STATUS="FAIL"
elif [[ ${BLOCKED_COUNT} -gt 0 ]]; then
  OVERALL_STATUS="BLOCKED"
fi

cat > "${REPORT_PATH}" <<EOF
{
  "\$schema": "https://gov.pai.dev/schemas/gov-rubric-report.schema.json",
  "schemaVersion": ${REPORT_SCHEMA_VERSION},
  "timestamp": "${REPORT_TIMESTAMP}",
  "pack": {
    "version": "2f9275c2707d",
    "digest": "bb28e962509ace6bf1d59e58b6e216290af7cee60cdb49f2aa03dd0c2b5cda76"
  },
  "project": {
    "name": "git-wrapper",
    "slug": "git-wrapper"
  },
  "summary": {
    "status": "${OVERALL_STATUS}",
    "pass": ${PASS_COUNT},
    "fail": ${FAIL_COUNT},
    "blocked": ${BLOCKED_COUNT}
  },
  "results": ${RESULTS_JSON}
}
EOF

echo "Report written to: ${REPORT_PATH}"
echo ""
echo "=== Summary ==="
echo "Status: ${OVERALL_STATUS}"
echo "Pass: ${PASS_COUNT}"
echo "Fail: ${FAIL_COUNT}"
echo "Blocked: ${BLOCKED_COUNT}"

if [[ "${OVERALL_STATUS}" == "PASS" ]]; then
  exit 0
else
  exit 1
fi
