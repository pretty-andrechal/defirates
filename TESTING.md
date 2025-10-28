# API Testing Guide

## Quick Diagnosis

Since `curl` works for you but the app doesn't, run this diagnostic:

```bash
cd /home/user/defirates
go run diagnose_api.go
```

This will test:
- ✅ curl (known to work for you)
- ❓ Go HTTP client (default)
- ❓ Go HTTP client (with User-Agent)
- 📋 Environment variables
- ⏱️ Different timeout values

## Step-by-Step Validation

For detailed API implementation testing:

```bash
go run api_validator.go
```

This validates each layer:
1. Basic HTTP connectivity
2. Request headers & response
3. JSON parsing
4. Market data structure
5. PendleClient implementation
6. Full multi-chain integration

## Expected Results

### If curl works but Go fails:

**Problem**: Missing User-Agent or other headers

**Solution**: I've added a User-Agent header to the client. Rebuild and test:

```bash
rm -f defirates defirates.db
go build -o defirates ./cmd/server
./defirates
```

### If both work:

You should see:
```
2025/10/28 XX:XX:XX Fetching Pendle markets...
2025/10/28 XX:XX:XX Found 50+ active Pendle markets
2025/10/28 XX:XX:XX Successfully stored 50+ yield rates
```

### If both fail:

Network access is still restricted. Use sample data mode:
```bash
./defirates -load-sample
```

## Manual API Test

Quick one-liner to test the API:

```bash
curl -s "https://api-v2.pendle.finance/api/core/v1/1/markets/active" | \
  jq '.results | length'
```

Should return a number like `25` or `50` (count of markets).

## Debugging Checklist

- [ ] Run `diagnose_api.go` and check all 5 tests
- [ ] Verify curl returns HTTP 200
- [ ] Check if Go client needs User-Agent
- [ ] Check environment variables (HTTP_PROXY, etc.)
- [ ] Rebuild app with latest changes
- [ ] Test with `./defirates` (no flags)
- [ ] If still failing, use `./defirates -load-sample`

## What Changed

I've updated `internal/api/pendle.go` to include:
```go
req.Header.Set("User-Agent", "DeFiRates/1.0 (+https://github.com/pretty-andrechal/defirates)")
```

Many APIs and WAFs (Web Application Firewalls) block requests without a User-Agent header.
