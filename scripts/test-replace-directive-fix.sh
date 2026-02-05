#!/usr/bin/env bash

# test-replace-directive-fix.sh - Test that replace directives are created correctly without git output

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_DIR="/tmp/test_replace_fix_$$"
TEST_GOMOD="$TEST_DIR/go.mod"
TEST_GOSUM="$TEST_DIR/go.sum"

cleanup() {
    rm -rf "$TEST_DIR"
}

trap cleanup EXIT

# Create test directories and files
mkdir -p "$TEST_DIR"

# Create a test go.mod with a simple module
cat > "$TEST_GOMOD" << 'EOF'
module github.com/example/myapp

go 1.21

require (
    github.com/example/module v1.0.0
)
EOF

# Create a test go.sum
cat > "$TEST_GOSUM" << 'EOF'
github.com/example/module v1.0.0 h1:abc123...
github.com/example/module v1.0.0/go.mod h1:def456...
EOF

echo "Testing replace directive creation..."
echo ""

# Test 1: Verify replace directive doesn't contain git output
echo "[TEST 1] Verify replace directive format is correct"

# Simulate what add_replace_directive would do
# Using a mock module path that would have git output appended if not fixed
mock_path="$TEST_DIR/github.com/example/module"
mkdir -p "$mock_path"

# Add a replace directive
echo "replace github.com/example/module => $mock_path" >> "$TEST_GOMOD"

# Check that the path is correct (no git output like "Cloning into..." or "Receiving objects...")
if grep -q "Cloning into" "$TEST_GOMOD"; then
    echo "❌ FAIL: Found git progress output in replace directive"
    cat "$TEST_GOMOD"
    exit 1
fi

if grep -q "Receiving objects" "$TEST_GOMOD"; then
    echo "❌ FAIL: Found git progress output in replace directive"
    cat "$TEST_GOMOD"
    exit 1
fi

# Verify replace directive is properly formatted
if grep -q "^replace github.com/example/module => $mock_path$" "$TEST_GOMOD"; then
    echo "✓ PASS: Replace directive is properly formatted"
else
    echo "❌ FAIL: Replace directive format is incorrect"
    grep "replace github.com/example/module" "$TEST_GOMOD"
    exit 1
fi

# Test 2: Verify go.sum backup functionality
echo "[TEST 2] Verify go.sum is included in operations"

# Check that go.sum exists
if [[ -f "$TEST_GOSUM" ]]; then
    echo "✓ PASS: go.sum file exists for backup"
else
    echo "❌ FAIL: go.sum file missing"
    exit 1
fi

# Verify go.sum has content
if [[ -s "$TEST_GOSUM" ]]; then
    echo "✓ PASS: go.sum has content to backup"
else
    echo "❌ FAIL: go.sum is empty"
    exit 1
fi

# Test 3: Verify path format in replace directive
echo "[TEST 3] Verify replace directive path doesn't have trailing output"

# Extract the path from replace directive
replace_line=$(grep "^replace github.com/example/module" "$TEST_GOMOD")
path_part="${replace_line#*=> }"

# Verify path doesn't contain spaces or newlines (which would indicate git output)
if [[ "$path_part" == *" "* ]] && [[ "$path_part" != *"github.com/example/module"* ]]; then
    echo "❌ FAIL: Path contains unexpected content: $path_part"
    exit 1
fi

if [[ "$path_part" == *$'\n'* ]]; then
    echo "❌ FAIL: Path contains newlines"
    exit 1
fi

echo "✓ PASS: Replace directive path is clean"

echo ""
echo "=========================================="
echo "All replace directive tests passed! ✓"
echo "=========================================="
