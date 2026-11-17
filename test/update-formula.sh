#!/usr/bin/env bash
#
# update-formula.sh: Safely update Homebrew formula URLs and SHA-256 checksums
#
# Usage:
#   update-formula.sh <formula_path> <tag> <darwin_sha> <linux_amd64_sha> <linux_arm64_sha>
#
# Exits with non-zero status if the formula layout does not match expectations
# or if replacement verification fails.

set -euo pipefail

FORMULA_FILE="${1:-}"
TAG="${2:-}"
DARWIN_SHA="${3:-}"
LINUX_AMD64_SHA="${4:-}"
LINUX_ARM64_SHA="${5:-}"

if [[ -z "$FORMULA_FILE" || -z "$TAG" || -z "$DARWIN_SHA" || -z "$LINUX_AMD64_SHA" || -z "$LINUX_ARM64_SHA" ]]; then
    echo "Usage: $0 <formula_path> <tag> <darwin_sha> <linux_amd64_sha> <linux_arm64_sha>" >&2
    exit 1
fi

if [[ ! -f "$FORMULA_FILE" ]]; then
    echo "Error: Formula file not found: $FORMULA_FILE" >&2
    exit 1
fi

# Validate SHA256 format (64 hex characters)
for sha in "$DARWIN_SHA" "$LINUX_AMD64_SHA" "$LINUX_ARM64_SHA"; do
    if [[ ! "$sha" =~ ^[0-9a-fA-F]{64}$ ]]; then
        echo "Error: Invalid SHA256 checksum format: $sha" >&2
        exit 1
    fi
done

# Verify initial structure: exactly 3 url lines and 3 sha256 lines
URL_COUNT=$(grep -c '^[[:space:]]*url ' "$FORMULA_FILE" || true)
SHA_COUNT=$(grep -c '^[[:space:]]*sha256 ' "$FORMULA_FILE" || true)

if [[ "$URL_COUNT" -ne 3 ]]; then
    echo "Error: Expected exactly 3 url entries in $FORMULA_FILE, found $URL_COUNT" >&2
    exit 1
fi

if [[ "$SHA_COUNT" -ne 3 ]]; then
    echo "Error: Expected exactly 3 sha256 entries in $FORMULA_FILE, found $SHA_COUNT" >&2
    exit 1
fi

# Verify the 3 expected platform archives exist in the formula before modifying
if ! grep -q 'pvtr_Darwin_all\.tar\.gz' "$FORMULA_FILE" && ! grep -q 'privateer_Darwin_all\.tar\.gz' "$FORMULA_FILE"; then
    echo "Error: Darwin archive entry not found in $FORMULA_FILE" >&2
    exit 1
fi

if ! grep -q 'pvtr_Linux_x86_64\.tar\.gz' "$FORMULA_FILE" && ! grep -q 'privateer_Linux_x86_64\.tar\.gz' "$FORMULA_FILE"; then
    echo "Error: Linux x86_64 archive entry not found in $FORMULA_FILE" >&2
    exit 1
fi

if ! grep -q 'pvtr_Linux_arm64\.tar\.gz' "$FORMULA_FILE" && ! grep -q 'privateer_Linux_arm64\.tar\.gz' "$FORMULA_FILE"; then
    echo "Error: Linux arm64 archive entry not found in $FORMULA_FILE" >&2
    exit 1
fi

# Perform replacement
TMP_FILE="$(mktemp)"
trap 'rm -f "$TMP_FILE"' EXIT

awk -v tag="$TAG" -v d="$DARWIN_SHA" -v x="$LINUX_AMD64_SHA" -v a="$LINUX_ARM64_SHA" '
BEGIN { section = 0 }
/if OS\.mac\?/ { section = 1 }
/elsif OS\.linux\? && Hardware::CPU\.intel\?/ { section = 2 }
/elsif OS\.linux\? && Hardware::CPU\.arm\?/ { section = 3 }
/^[[:space:]]*url / {
    gsub(/releases\/download\/v[^\/]*\//, "releases/download/" tag "/")
    gsub(/\/privateer_/, "/pvtr_")
    print
    next
}
/^[[:space:]]*sha256 / {
    if (section == 1) {
        sub(/sha256 "[^"]*"/, "sha256 \"" d "\"")
    } else if (section == 2) {
        sub(/sha256 "[^"]*"/, "sha256 \"" x "\"")
    } else if (section == 3) {
        sub(/sha256 "[^"]*"/, "sha256 \"" a "\"")
    }
    print
    next
}
{ print }
' "$FORMULA_FILE" > "$TMP_FILE"

# Post-replacement validation
POST_URL_COUNT=$(grep -c '^[[:space:]]*url ' "$TMP_FILE" || true)
POST_SHA_COUNT=$(grep -c '^[[:space:]]*sha256 ' "$TMP_FILE" || true)

if [[ "$POST_URL_COUNT" -ne 3 || "$POST_SHA_COUNT" -ne 3 ]]; then
    echo "Error: Post-update entry count mismatch (urls: $POST_URL_COUNT, shas: $POST_SHA_COUNT)" >&2
    exit 1
fi

if ! grep -q "releases/download/${TAG}/pvtr_Darwin_all.tar.gz" "$TMP_FILE"; then
    echo "Error: Darwin URL was not updated correctly" >&2
    exit 1
fi

if ! grep -q "sha256 \"${DARWIN_SHA}\"" "$TMP_FILE"; then
    echo "Error: Darwin SHA was not updated correctly" >&2
    exit 1
fi

if ! grep -q "releases/download/${TAG}/pvtr_Linux_x86_64.tar.gz" "$TMP_FILE"; then
    echo "Error: Linux x86_64 URL was not updated correctly" >&2
    exit 1
fi

if ! grep -q "sha256 \"${LINUX_AMD64_SHA}\"" "$TMP_FILE"; then
    echo "Error: Linux x86_64 SHA was not updated correctly" >&2
    exit 1
fi

if ! grep -q "releases/download/${TAG}/pvtr_Linux_arm64.tar.gz" "$TMP_FILE"; then
    echo "Error: Linux arm64 URL was not updated correctly" >&2
    exit 1
fi

if ! grep -q "sha256 \"${LINUX_ARM64_SHA}\"" "$TMP_FILE"; then
    echo "Error: Linux arm64 SHA was not updated correctly" >&2
    exit 1
fi

# Atomically overwrite target formula
cp "$TMP_FILE" "$FORMULA_FILE"
echo "Successfully updated $FORMULA_FILE for release $TAG"
