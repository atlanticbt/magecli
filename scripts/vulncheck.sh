#!/usr/bin/env bash
# Run govulncheck and fail only on symbol-level findings whose OSV IDs are not
# allowlisted in .govulncheck-ignore. Advisories without an upstream fix (e.g.
# unmaintained transitive dependencies we don't actually call) are recorded
# there with a justification instead of leaving CI permanently red.
set -euo pipefail

cd "$(dirname "$0")/.."

report=$(mktemp)
trap 'rm -f "$report"' EXIT

# Exit code 0 = clean, 3 = vulnerabilities found; anything else is a scan error.
set +e
go run golang.org/x/vuln/cmd/govulncheck@latest -format json ./... >"$report"
status=$?
set -e
if [ "$status" -ne 0 ] && [ "$status" -ne 3 ]; then
    echo "govulncheck failed with exit code $status" >&2
    cat "$report" >&2
    exit "$status"
fi

# Symbol-level findings only (trace[0].function set) — the same bar govulncheck
# uses to fail in its default text mode.
found=$(jq -r 'select(.finding != null) | .finding | select(.trace[0].function != null) | .osv' "$report" | sort -u)

ignored=""
if [ -f .govulncheck-ignore ]; then
    ignored=$(grep -vE '^[[:space:]]*(#|$)' .govulncheck-ignore | sort -u)
fi

unexpected=$(comm -23 <(printf '%s' "$found") <(printf '%s' "$ignored"))
suppressed=$(comm -12 <(printf '%s' "$found") <(printf '%s' "$ignored"))

if [ -n "$suppressed" ]; then
    echo "Ignored per .govulncheck-ignore:"
    for id in $suppressed; do
        jq -r --arg id "$id" 'select(.osv != null) | .osv | select(.id == $id) | "  \(.id): \(.summary)"' "$report"
    done
fi

if [ -n "$unexpected" ]; then
    echo ""
    echo "govulncheck found vulnerabilities not listed in .govulncheck-ignore:" >&2
    for id in $unexpected; do
        jq -r --arg id "$id" 'select(.osv != null) | .osv | select(.id == $id) | "  \(.id): \(.summary)\n    https://pkg.go.dev/vuln/\(.id)"' "$report" >&2
    done
    exit 1
fi

echo "No new vulnerabilities found."
