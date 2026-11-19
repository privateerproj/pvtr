#!/usr/bin/env bash
#
# Check that test coverage meets or exceeds the required threshold.
#
# Usage:
#   ./test/coverage.sh [profile] [threshold]
#
# Environment variables:
#   COVERAGE_FILE: path to coverprofile file (default: coverage.out)
#   TESTCOVERAGE_THRESHOLD: minimum coverage percentage (default: 80)

set -euo pipefail

COVERAGE_FILE="${1:-${COVERAGE_FILE:-coverage.out}}"
THRESHOLD="${2:-${TESTCOVERAGE_THRESHOLD:-80}}"

if [[ ! -f "$COVERAGE_FILE" ]]; then
    echo "Coverage file not found: $COVERAGE_FILE" >&2
    exit 1
fi

echo "Quality Gate: checking test coverage is above threshold ..."
echo "Threshold             : ${THRESHOLD} %"

TOTAL_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep total | grep -Eo '[0-9]+\.[0-9]+')
echo "Current test coverage : ${TOTAL_COVERAGE} %"

# Preserve the exact comparison behavior from CI: coverage must be greater than threshold
IS_ABOVE=$(awk -v total="$TOTAL_COVERAGE" -v thresh="$THRESHOLD" 'BEGIN { print (total > thresh) }')

if [[ "$IS_ABOVE" -eq 1 ]]; then
    echo "OK"
else
    echo "Current test coverage is below threshold. Please add more unit tests or adjust threshold to a lower value."
    echo "Failed"
    exit 1
fi
