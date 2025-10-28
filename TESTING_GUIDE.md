# DeFi Rates Testing Guide

This guide covers the comprehensive test suite created to ensure everything works correctly after the debugging session.

## Quick Start

Run all tests:
```bash
chmod +x run_tests.sh
./run_tests.sh
```

Or run specific test suites:
```bash
# API client tests
go test -v ./internal/api

# Database tests
go test -v ./internal/database

# Handler tests
go test -v ./internal/handlers

# Integration tests only
go test -v ./internal/api -run Integration
```

## Test Coverage

### 1. API Client Tests (`internal/api/pendle_test.go`)

Tests everything we debugged related to the Pendle API:

#### ✅ What's Tested:
- **Correct API endpoint structure** (`/api/core/v1/{chainId}/markets/active`)
- **Response parsing** (markets vs results, nested details)
- **JSON structure matching** actual API response
- **ChainID injection** (since API doesn't return it)
- **User-Agent header** presence
- **Error handling** (403, 400, invalid JSON)
- **Empty response handling**
- **Expiry date filtering** (active vs expired markets)
- **Chain name mapping** (all supported chains)

**Key Tests:**
- `TestPendleClient_GetMarketsForChain` - Tests market fetching with mock server
- `TestMarket_JSONParsing` - Verifies struct matches real API response
- `TestGetActiveMarkets_ExpiryFiltering` - Tests expiry date logic
- `TestGetChainName` - Tests all chain mappings including unknown chains

### 2. Database Tests (`internal/database/database_test.go`)

Tests all database operations:

#### ✅ What's Tested:
- **Database initialization** and schema creation
- **Protocol CRUD operations** (create, update, retrieve)
- **YieldRate upsert logic** (insert vs update)
- **Filtering** by asset, chain, APY range, TVL range
- **Combined filters** (multiple criteria)
- **Sorting** (APY, TVL, updated_at - both ASC and DESC)
- **Distinct queries** (unique assets and chains)

**Key Tests:**
- `TestCreateOrUpdateProtocol` - Tests upsert behavior
- `TestUpsertYieldRate` - Verifies update vs insert logic
- `TestGetYieldRates_Filtering` - Tests all filter combinations
- `TestGetYieldRates_Sorting` - Verifies sort orders
- `TestGetDistinctAssets` - Tests unique value queries

### 3. Handler Tests (`internal/handlers/handlers_test.go`)

Tests HTTP layer and template rendering:

#### ✅ What's Tested:
- **Index page rendering** with and without data
- **Query parameter parsing** (all filter types)
- **HTMX partial rendering** (vs full page)
- **Template type comparisons** (float64 vs float literals)
  - Fixed: `ge .TVL 1000000` → `ge .TVL 1000000.0`
  - Fixed: `ne .Filters.MinAPY 0` → `ne .Filters.MinAPY 0.0`
- **TVL formatting** (K and M suffixes)
- **Date formatting** (maturity dates)
- **APY color coding** (high/medium/low classes)

**Key Tests:**
- `TestHandleIndex_EmptyDatabase` - Tests empty state
- `TestHandleIndex_WithData` - Verifies data rendering
- `TestHandleIndex_Filtering` - Tests query parameters
- `TestHandleIndex_HTMX` - Tests partial responses
- `TestHandleIndex_TemplateTypeComparisons` - Tests type safety fixes

### 4. Integration Tests (`internal/api/integration_test.go`)

Tests complete end-to-end flows:

#### ✅ What's Tested:
- **Full fetch and store pipeline** (API → DB)
- **Market to YieldRate conversion**
- **Multi-chain fetching**
- **Real API interaction** (when accessible)

**Note:** Integration tests may fail if Pendle API is blocked by firewall. This is expected and handled gracefully.

## Test Organization

```
defirates/
├── internal/
│   ├── api/
│   │   ├── pendle.go
│   │   ├── pendle_test.go         ← API unit tests
│   │   ├── integration_test.go    ← End-to-end tests
│   │   └── fetcher.go
│   ├── database/
│   │   ├── database.go
│   │   └── database_test.go       ← Database tests
│   └── handlers/
│       ├── handlers.go
│       └── handlers_test.go       ← HTTP/template tests
├── run_tests.sh                   ← Test runner script
└── TESTING_GUIDE.md              ← This file
```

## Running Specific Tests

### By Package
```bash
go test ./internal/api       # API tests
go test ./internal/database  # Database tests
go test ./internal/handlers  # Handler tests
```

### By Test Name
```bash
# Run specific test
go test -v ./internal/api -run TestPendleClient_GetMarketsForChain

# Run tests matching pattern
go test -v ./internal/api -run ".*Expiry.*"
```

### With Verbose Output
```bash
go test -v ./...
```

### Skip Integration Tests
```bash
go test -short ./...
```

## Coverage Reports

Generate coverage report:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

View coverage in browser:
```bash
go tool cover -html=coverage.out
```

## Race Detection

Check for race conditions:
```bash
go test -race ./...
```

## Benchmarking

Run benchmarks (when added):
```bash
go test -bench=. ./...
```

## Common Issues Debugged

### Issue 1: API Response Structure Mismatch
**Problem:** API returns `"markets"` but code expected `"results"`
**Tests:** `TestMarket_JSONParsing`, `TestPendleClient_GetMarketsForChain`
**Fix:** Updated `MarketsResponse` struct

### Issue 2: Template Type Comparisons
**Problem:** `ge .TVL 1000000` - can't compare float64 with int
**Tests:** `TestHandleIndex_TemplateTypeComparisons`
**Fix:** Changed to `ge .TVL 1000000.0`

### Issue 3: Missing User-Agent
**Problem:** API may block requests without User-Agent
**Tests:** `TestPendleClient_GetMarketsForChain` checks for header
**Fix:** Added `User-Agent` header to all requests

### Issue 4: ChainID Not in Response
**Problem:** API doesn't include chainId in response
**Tests:** `TestPendleClient_GetMarketsForChain` verifies ChainID is set
**Fix:** Manually inject ChainID after parsing

### Issue 5: Unsupported Chains
**Problem:** Polygon (137) and others no longer supported
**Tests:** `TestGetChainName` includes all current chains
**Fix:** Updated chain list to match API

## Continuous Integration

To integrate with CI/CD:

```yaml
# Example GitHub Actions workflow
- name: Run tests
  run: |
    chmod +x run_tests.sh
    ./run_tests.sh

- name: Upload coverage
  run: |
    go test -coverprofile=coverage.out ./...
    # Upload to codecov or similar
```

## Adding New Tests

When adding features, add corresponding tests:

1. **Unit tests** in same package as code (`*_test.go`)
2. **Integration tests** for complete flows
3. **Handler tests** for new endpoints
4. Update `run_tests.sh` if needed

Example test structure:
```go
func TestNewFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   interface{}
        want    interface{}
        wantErr bool
    }{
        // Test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

## Test Database Cleanup

Tests automatically clean up temporary databases:
```go
defer func() {
    db.Close()
    os.Remove(dbPath)
}()
```

If tests are interrupted, manually clean:
```bash
rm -f test_*.db
```

## Best Practices

1. ✅ **Table-driven tests** for multiple scenarios
2. ✅ **Descriptive test names** (`TestComponent_Behavior_Condition`)
3. ✅ **Test both success and failure cases**
4. ✅ **Mock external dependencies** (HTTP servers, databases)
5. ✅ **Clean up resources** (temp files, connections)
6. ✅ **Test edge cases** (empty data, invalid input)
7. ✅ **Verify error messages** (not just error presence)

## Getting Help

If tests fail:

1. Run with verbose output: `go test -v ./...`
2. Check specific failing test: `go test -v -run TestName`
3. Verify test databases are cleaned: `rm -f test_*.db`
4. For integration tests: May fail if API is blocked (expected)

## Summary

This test suite ensures:
- ✅ API client correctly parses Pendle API responses
- ✅ Database operations work correctly with all filters
- ✅ Templates render without type errors
- ✅ HTMX partial updates work
- ✅ End-to-end flow from API → DB → UI works
- ✅ All bugs from debugging session are covered

Run `./run_tests.sh` before committing to ensure everything works!
