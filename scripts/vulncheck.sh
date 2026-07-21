#!/usr/bin/env bash
# Run govulncheck, failing only on advisories not allowlisted in .govulncheck-ignore.
# govulncheck has no native suppression mechanism (golang/go#59507), so this
# wrapper filters its JSON output. Only symbol-level findings (code that is
# actually reachable) are considered, matching govulncheck's default behavior.
set -euo pipefail

cd "$(dirname "$0")/.."

IGNORE_FILE=".govulncheck-ignore"
ignored=$(grep -vE '^[[:space:]]*(#|$)' "$IGNORE_FILE" 2>/dev/null || true)

output=$(govulncheck -format json ./...)

found=$(jq -r 'select(.finding != null) | select(.finding.trace[0].function != null) | .finding.osv' <<<"$output" | sort -u)

fail=0
for id in $found; do
  if grep -qxF "$id" <<<"$ignored"; then
    echo "ignored:       $id (allowlisted in $IGNORE_FILE)"
  else
    echo "vulnerability: $id — https://pkg.go.dev/vuln/$id"
    fail=1
  fi
done

if [ "$fail" -ne 0 ]; then
  echo "govulncheck found vulnerabilities not allowlisted in $IGNORE_FILE" >&2
  exit 1
fi
echo "govulncheck: no findings outside $IGNORE_FILE allowlist"
