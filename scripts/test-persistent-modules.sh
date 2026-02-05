#!/usr/bin/env bash

# test-persistent-modules.sh
# Test script to verify persistent module directory and reuse functionality

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_MODULES_DIR="/tmp/test-persistent-modules"
GOMOD_ORIG="$WORKSPACE_ROOT/go.mod"

cleanup() {
    rm -rf "$TEST_MODULES_DIR" 2>/dev/null || true
    rm -f "$GOMOD_ORIG.backup" 2>/dev/null || true
}

print_test() {
    echo -e "\n${BLUE}[TEST]${NC} $*"
}

print_pass() {
    echo -e "${GREEN}[PASS]${NC} $*"
}

print_fail() {
    echo -e "${RED}[FAIL]${NC} $*"
    exit 1
}

# Test 1: Verify --dir is required without --restore
test_dir_required() {
    print_test "Verify --dir is required without --restore"
    
    cd "$WORKSPACE_ROOT"
    if bash "$SCRIPT_DIR/use-local-modules.sh" 2>/dev/null >/dev/null; then
        print_fail "Script should fail without --dir"
    else
        print_pass "Script correctly fails without --dir parameter"
    fi
}

# Test 2: Verify --restore works without --dir
test_restore_without_dir() {
    print_test "Verify --restore works without --dir"
    
    cd "$WORKSPACE_ROOT"
    output=$(bash "$SCRIPT_DIR/use-local-modules.sh" --restore 2>&1 || true)
    if echo "$output" | grep -qi "dir parameter"; then
        print_fail "Restore mode should not require --dir parameter"
    else
        print_pass "Restore mode works without --dir"
    fi
}

# Test 3: Verify --dir parameter is accepted
test_dir_accepted() {
    print_test "Verify --dir parameter is accepted and directory created"
    
    cd "$WORKSPACE_ROOT"
    output=$(bash "$SCRIPT_DIR/use-local-modules.sh" --dir "$TEST_MODULES_DIR" 2>&1 || true)
    if echo "$output" | grep -qi "no modules specified"; then
        if [[ -d "$TEST_MODULES_DIR" ]]; then
            print_pass "Directory created: $TEST_MODULES_DIR"
        else
            print_fail "Directory not created"
        fi
    else
        print_fail "Script should accept --dir parameter"
    fi
}

# Test 4: Verify reuse logic exists in clone_module
test_reuse_logic() {
    print_test "Verify clone_module contains reuse checking logic"
    
    if grep -q '\-d.*\.git' "$SCRIPT_DIR/use-local-modules.sh"; then
        print_pass "Reuse checking logic present"
    else
        print_fail "Reuse checking logic not found"
    fi
}

# Test 5: Verify MODULES_DIR configuration exists
test_modules_dir_config() {
    print_test "Verify MODULES_DIR configuration variable exists"
    
    if grep -q '^MODULES_DIR=""' "$SCRIPT_DIR/use-local-modules.sh"; then
        print_pass "MODULES_DIR configuration variable found"
    else
        print_fail "MODULES_DIR configuration variable not found"
    fi
}

# Test 6: Verify directory structure is documented
test_structure_doc() {
    print_test "Verify directory structure is documented in help"
    
    cd "$WORKSPACE_ROOT"
    output=$(bash "$SCRIPT_DIR/use-local-modules.sh" --help 2>&1 || true)
    if echo "$output" | grep -q "DIRECTORY/MODULE"; then
        print_pass "Directory structure documented in help"
    else
        print_fail "Directory structure not documented"
    fi
}

# Test 7: Verify --dir appears in help
test_help_shows_dir() {
    print_test "Verify --help shows --dir requirement"
    
    cd "$WORKSPACE_ROOT"
    output=$(bash "$SCRIPT_DIR/use-local-modules.sh" --help 2>&1 || true)
    if echo "$output" | grep -q "Directory to store cloned modules"; then
        print_pass "Help correctly documents --dir requirement"
    else
        print_fail "Help missing --dir documentation"
    fi
}

# Run all tests
echo "===================================================="
echo "Testing Persistent Module Directory Feature"
echo "===================================================="

cleanup
test_dir_required
test_restore_without_dir
test_dir_accepted
test_reuse_logic
test_modules_dir_config
test_structure_doc
test_help_shows_dir

echo ""
echo "===================================================="
echo -e "${GREEN}All tests passed!${NC}"
echo "===================================================="

cleanup
