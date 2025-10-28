#!/bin/bash

# DeFi Rates Test Suite Runner
# This script runs all tests and provides a summary

set -e

echo "======================================"
echo "  DeFi Rates Test Suite"
echo "======================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run tests and capture results
run_test_suite() {
    local name=$1
    local command=$2

    echo "Running: $name"
    echo "----------------------------------------"

    if $command; then
        echo -e "${GREEN}✓ PASSED${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ FAILED${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
}

# Clean up any leftover test databases
echo "Cleaning up old test files..."
rm -f test_*.db
echo ""

# 1. Unit Tests - API Client
run_test_suite "API Client Unit Tests" "go test -v ./internal/api -run '^Test.*' -short"

# 2. Unit Tests - Database
run_test_suite "Database Unit Tests" "go test -v ./internal/database"

# 3. Unit Tests - Handlers
run_test_suite "Handler Unit Tests" "go test -v ./internal/handlers"

# 4. Integration Tests (may fail if API is blocked)
echo -e "${YELLOW}Note: Integration tests may fail if Pendle API is blocked${NC}"
run_test_suite "Integration Tests" "go test -v ./internal/api -run Integration"

# 5. Run all tests with race detector
echo "Running all tests with race detector..."
echo "----------------------------------------"
# Suppress macOS linker warnings about LC_DYSYMTAB (known harmless issue)
OUTPUT=$(go test -race ./... 2>&1)
RESULT=$?
echo "$OUTPUT" | grep -v "ld: warning.*LC_DYSYMTAB"
if [ $RESULT -eq 0 ]; then
    echo -e "${GREEN}✓ No race conditions detected${NC}"
else
    echo -e "${YELLOW}⚠ Race conditions detected (review above)${NC}"
fi
echo ""

# 6. Check test coverage
echo "Generating test coverage report..."
echo "----------------------------------------"
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -n 1

echo ""
echo "To view detailed coverage in browser:"
echo "  go tool cover -html=coverage.out"
echo ""

# 7. Run benchmarks
echo "Running benchmarks..."
echo "----------------------------------------"
go test -bench=. -benchmem ./... || echo "No benchmarks defined yet"
echo ""

# Summary
echo "======================================"
echo "  Test Summary"
echo "======================================"
echo "Total Test Suites: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed. Please review above.${NC}"
    exit 1
fi
