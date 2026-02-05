#!/usr/bin/env bash

# test-sed-fix.sh - Test that sed correctly removes modules and doesn't insert "pis/d"

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_GOMOD="/tmp/test_gomod_$$"
TEST_GOMOD_BACKUP="/tmp/test_gomod_backup_$$"

cleanup() {
    rm -f "$TEST_GOMOD" "$TEST_GOMOD_BACKUP" "$TEST_GOMOD.bak"
}

trap cleanup EXIT

# Create a test go.mod file
create_test_gomod() {
    cat > "$TEST_GOMOD" << 'EOF'
module github.com/example/myapp

go 1.21

require (
    github.com/ironcore-dev/ironcore v0.35.0
    github.com/some/other v1.0.0
)
EOF
}

echo "Testing sed fix for module removal..."
echo ""

# Test 1: Verify "pis/d" doesn't appear in file
echo "[TEST 1] Verify no 'pis/d' corruption in go.mod"
create_test_gomod
original_size=$(wc -c < "$TEST_GOMOD")

# Use the same sed command as the script (with fixed syntax - escape forward slashes)
escaped_module="github.com/ironcore-dev/ironcore"
escaped_module="${escaped_module//\//\\/}"
sed -i.bak "/^[[:space:]]*$escaped_module[[:space:]]/d" "$TEST_GOMOD"

if grep -q "pis/d" "$TEST_GOMOD"; then
    echo "❌ FAIL: Found 'pis/d' corruption in go.mod"
    cat "$TEST_GOMOD"
    exit 1
else
    echo "✓ PASS: No 'pis/d' corruption found"
fi

# Test 2: Verify module line was actually removed
echo "[TEST 2] Verify module line was removed"
if grep -q "ironcore-dev/ironcore" "$TEST_GOMOD"; then
    echo "❌ FAIL: Module line still present in go.mod"
    cat "$TEST_GOMOD"
    exit 1
else
    echo "✓ PASS: Module line was successfully removed"
fi

# Test 3: Verify other modules are still present
echo "[TEST 3] Verify other modules are preserved"
if grep -q "github.com/some/other" "$TEST_GOMOD"; then
    echo "✓ PASS: Other modules preserved"
else
    echo "❌ FAIL: Other modules were removed"
    cat "$TEST_GOMOD"
    exit 1
fi

# Test 4: Verify backup was created
echo "[TEST 4] Verify backup file created"
if [[ -f "$TEST_GOMOD.bak" ]]; then
    echo "✓ PASS: Backup file created"
else
    echo "❌ FAIL: Backup file not created"
    exit 1
fi

# Test 5: Test with leading whitespace
echo "[TEST 5] Test module removal with leading whitespace"
create_test_gomod
sed -i 's/^    github\.com/  github\.com/' "$TEST_GOMOD"  # Change indentation
escaped_module="github.com/ironcore-dev/ironcore"
escaped_module="${escaped_module//\//\\/}"
sed -i.bak "/^[[:space:]]*$escaped_module[[:space:]]/d" "$TEST_GOMOD"

if ! grep -q "ironcore-dev/ironcore" "$TEST_GOMOD"; then
    echo "✓ PASS: Module removed even with different whitespace"
else
    echo "❌ FAIL: Module not removed with different whitespace"
    exit 1
fi

echo ""
echo "=========================================="
echo "All sed fix tests passed! ✓"
echo "=========================================="
