#!/usr/bin/env bash
#
# update_formula_test.sh: Tests for update-formula.sh validation and replacement
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UPDATE_SCRIPT="$SCRIPT_DIR/update-formula.sh"

TEST_TMP="$(mktemp -d)"
trap 'rm -rf "$TEST_TMP"' EXIT

# Sample valid formula fixture (based on pvtr.rb)
cat << 'EOF' > "$TEST_TMP/valid_pvtr.rb"
# typed: false
# frozen_string_literal: true

class Pvtr < Formula
  desc "Pluggable compliance testing framework"
  homepage "https://github.com/privateerproj/pvtr"
  license "Apache-2.0"

  if OS.mac?
    url "https://github.com/privateerproj/pvtr/releases/download/v0.24.2/pvtr_Darwin_all.tar.gz"
    sha256 "1111111111111111111111111111111111111111111111111111111111111111"
  elsif OS.linux? && Hardware::CPU.intel?
    url "https://github.com/privateerproj/pvtr/releases/download/v0.24.2/pvtr_Linux_x86_64.tar.gz"
    sha256 "2222222222222222222222222222222222222222222222222222222222222222"
  elsif OS.linux? && Hardware::CPU.arm?
    url "https://github.com/privateerproj/pvtr/releases/download/v0.24.2/pvtr_Linux_arm64.tar.gz"
    sha256 "3333333333333333333333333333333333333333333333333333333333333333"
  end

  link_overwrite "bin/privateer"

  def install
    bin.install "pvtr"
    bin.install_symlink "pvtr" => "privateer"
  end

  test do
    system "#{bin}/pvtr", "version"
  end
end
EOF

DARWIN_SHA="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
AMD64_SHA="bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
ARM64_SHA="cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
NEW_TAG="v0.24.4"

echo "Test 1: Valid formula updates successfully"
cp "$TEST_TMP/valid_pvtr.rb" "$TEST_TMP/test1.rb"
bash "$UPDATE_SCRIPT" "$TEST_TMP/test1.rb" "$NEW_TAG" "$DARWIN_SHA" "$AMD64_SHA" "$ARM64_SHA"
grep -q "releases/download/${NEW_TAG}/pvtr_Darwin_all.tar.gz" "$TEST_TMP/test1.rb"
grep -q "sha256 \"${DARWIN_SHA}\"" "$TEST_TMP/test1.rb"
grep -q "releases/download/${NEW_TAG}/pvtr_Linux_x86_64.tar.gz" "$TEST_TMP/test1.rb"
grep -q "sha256 \"${AMD64_SHA}\"" "$TEST_TMP/test1.rb"
grep -q "releases/download/${NEW_TAG}/pvtr_Linux_arm64.tar.gz" "$TEST_TMP/test1.rb"
grep -q "sha256 \"${ARM64_SHA}\"" "$TEST_TMP/test1.rb"
echo "  -> PASSED"

echo "Test 2: Legacy privateer_ prefix is updated to pvtr_"
sed 's/pvtr_/privateer_/g' "$TEST_TMP/valid_pvtr.rb" > "$TEST_TMP/test2.rb"
bash "$UPDATE_SCRIPT" "$TEST_TMP/test2.rb" "$NEW_TAG" "$DARWIN_SHA" "$AMD64_SHA" "$ARM64_SHA"
grep -q "releases/download/${NEW_TAG}/pvtr_Darwin_all.tar.gz" "$TEST_TMP/test2.rb"
grep -q "releases/download/${NEW_TAG}/pvtr_Linux_x86_64.tar.gz" "$TEST_TMP/test2.rb"
grep -q "releases/download/${NEW_TAG}/pvtr_Linux_arm64.tar.gz" "$TEST_TMP/test2.rb"
echo "  -> PASSED"

echo "Test 3: Missing checksum entry fails safely"
sed '/Hardware::CPU.arm?/,+2d' "$TEST_TMP/valid_pvtr.rb" > "$TEST_TMP/test3.rb"
if bash "$UPDATE_SCRIPT" "$TEST_TMP/test3.rb" "$NEW_TAG" "$DARWIN_SHA" "$AMD64_SHA" "$ARM64_SHA" 2>/dev/null; then
    echo "  -> FAILED (expected non-zero exit for missing checksum)" >&2
    exit 1
fi
echo "  -> PASSED"

echo "Test 4: Unexpected extra checksum fails safely"
sed '/sha256/ { p; }' "$TEST_TMP/valid_pvtr.rb" > "$TEST_TMP/test4.rb"
if bash "$UPDATE_SCRIPT" "$TEST_TMP/test4.rb" "$NEW_TAG" "$DARWIN_SHA" "$AMD64_SHA" "$ARM64_SHA" 2>/dev/null; then
    echo "  -> FAILED (expected non-zero exit for extra checksums)" >&2
    exit 1
fi
echo "  -> PASSED"

echo "Test 5: Invalid SHA256 format fails safely"
cp "$TEST_TMP/valid_pvtr.rb" "$TEST_TMP/test5.rb"
if bash "$UPDATE_SCRIPT" "$TEST_TMP/test5.rb" "$NEW_TAG" "invalid_short_hash" "$AMD64_SHA" "$ARM64_SHA" 2>/dev/null; then
    echo "  -> FAILED (expected non-zero exit for malformed SHA)" >&2
    exit 1
fi
echo "  -> PASSED"

echo "All formula updater tests passed successfully!"
