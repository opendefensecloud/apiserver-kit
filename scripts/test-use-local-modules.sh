#!/usr/bin/env bash

# test-use-local-modules.sh
# 
# Test script for use-local-modules.sh
# Demonstrates functionality without actually cloning real modules

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT="$SCRIPT_DIR/use-local-modules.sh"
TEST_DIR="${TMPDIR:-/tmp}/use-local-modules-test-$$"
TEST_GOMOD="$TEST_DIR/go.mod"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_test() {
    echo -e "${BLUE}→${NC} $*"
}

log_pass() {
    echo -e "${GREEN}✓${NC} $*"
}

log_info() {
    echo -e "${YELLOW}ℹ${NC} $*"
}

# Cleanup
cleanup() {
    if [[ -d "$TEST_DIR" ]]; then
        rm -rf "$TEST_DIR"
    fi
}

trap cleanup EXIT

# Setup
mkdir -p "$TEST_DIR"

# Create a minimal test go.mod
cat > "$TEST_GOMOD" << 'EOF'
module go.opendefense.cloud/kit

go 1.25

require (
    k8s.io/api v0.35.0
    k8s.io/apimachinery v0.35.0
)
EOF

log_test "Test 1: Help command"
if "$SCRIPT" --help | grep -q "use-local-modules.sh"; then
    log_pass "Help command works"
else
    echo "FAIL: Help command"
    exit 1
fi

log_test "Test 2: No arguments error"
OUTPUT=$("$SCRIPT" 2>&1 || true)
if echo "$OUTPUT" | grep -q "No modules specified"; then
    log_pass "Correctly rejects missing modules"
else
    echo "FAIL: Should show error when no modules provided"
    echo "Got: $OUTPUT"
    exit 1
fi

log_test "Test 3: Invalid module format"
OUTPUT=$("$SCRIPT" "invalid" 2>&1 || true)
if echo "$OUTPUT" | grep -q "Invalid module reference"; then
    log_pass "Validates module references"
else
    echo "FAIL: Should reject invalid module format"
    echo "Got: $OUTPUT"
    exit 1
fi

log_test "Test 4: Syntax validation"
if ! bash -n "$SCRIPT"; then
    echo "FAIL: Script has syntax errors"
    exit 1
fi
log_pass "Script syntax is valid"

log_test "Test 5: Script is executable"
if [[ ! -x "$SCRIPT" ]]; then
    echo "FAIL: Script is not executable"
    exit 1
fi
log_pass "Script is executable"

log_test "Test 6: Required functions exist"
for func in log_info log_error log_success cleanup validate_gomod backup_gomod; do
    if ! grep -q "^$func()" "$SCRIPT"; then
        echo "FAIL: Missing function: $func"
        exit 1
    fi
done
log_pass "All required functions exist"

echo ""
log_info "All tests passed!"
log_info ""
log_info "To test with real modules (requires git and network):"
log_info "  cd /path/to/apiserver-kit"
log_info "  ./scripts/use-local-modules.sh --help"
log_info "  ./scripts/use-local-modules.sh github.com/ironcore-dev/ironcore"
log_info "  ./scripts/use-local-modules.sh --restore"
