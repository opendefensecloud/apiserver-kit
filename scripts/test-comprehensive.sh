#!/usr/bin/env bash

# test-comprehensive.sh - Comprehensive end-to-end validation

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_ROOT="$(dirname "$SCRIPT_DIR")"

echo "====================================================="
echo "Comprehensive Validation Test Suite"
echo "====================================================="
echo ""

total_tests=0
passed_tests=0

run_test() {
    local test_script="$1"
    local test_name="$2"
    
    echo "Running: $test_name"
    total_tests=$((total_tests + 1))
    
    if bash "$test_script" 2>&1 | tail -5; then
        passed_tests=$((passed_tests + 1))
        echo ""
    else
        echo "❌ FAILED: $test_name"
        echo ""
        return 1 || true
    fi
}

# Run all test suites
run_test "$SCRIPT_DIR/test-persistent-modules.sh" "Persistent Module Tests"
run_test "$SCRIPT_DIR/test-sed-fix.sh" "Sed Fix Tests"
run_test "$SCRIPT_DIR/test-replace-directive-fix.sh" "Replace Directive Tests"

echo "====================================================="
echo "Test Summary"
echo "====================================================="
echo "Total Test Suites: $total_tests"
echo "Passed: $passed_tests"
echo "Failed: $((total_tests - passed_tests))"
echo ""

if [[ $passed_tests -eq $total_tests ]]; then
    echo "✅ ALL TESTS PASSED"
    exit 0
else
    echo "❌ SOME TESTS FAILED"
    exit 1
fi
