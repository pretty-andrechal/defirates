# How to Fix 403 API Access Errors

This guide provides **practical, tested solutions** to resolve the 403 "Access Denied" errors when accessing Pendle and Beefy Finance APIs.

## Quick Solutions (Easiest First)

### 1. Deploy to Cloud/VPS (Most Reliable)

**Why it works:** Cloud provider IPs typically have good reputation and aren't blocked by Cloudflare.

**Providers to try:**
```bash
# DigitalOcean, AWS, GCP, Azure, Linode, Vultr, Hetzner
# Even free tier instances work!

# On your cloud instance:
git clone https://github.com/pretty-andrechal/defirates.git
cd defirates
go build -o defirates ./cmd/server
./defirates
```

**Cost:** $5-10/month for basic VPS, or free tier (AWS/GCP)

### 2. Use Your Home Network

**Why it works:** Residential IPs are rarely blocked by WAF.

If you're currently on a corporate network, VPN, or data center:
- Try running from home internet
- Use mobile hotspot
- Connect via residential ISP

### 3. Use Environment Variables for Proxy (Quick Setup)

Add proxy support without code changes:

```bash
# HTTP Proxy
export HTTP_PROXY=http://your-proxy:8080
export HTTPS_PROXY=http://your-proxy:8080

# SOCKS5 Proxy
export HTTP_PROXY=socks5://your-proxy:1080
export HTTPS_PROXY=socks5://your-proxy:1080

# Then run normally
./defirates
```

### 4. Try a Different Network

Simple tests to try:
1. **Mobile hotspot** - Use your phone's data
2. **Different WiFi** - Coffee shop, library, friend's house
3. **VPN service** - NordVPN, ExpressVPN, Mullvad, ProtonVPN

## Advanced Solutions

### Solution A: Enable Retry Logic with Delays

We've included a resilient HTTP client that implements:
- ✅ Exponential backoff retry (2s, 4s, 8s...)
- ✅ Request delays to avoid rate limiting
- ✅ Proxy support
- ✅ Automatic 403 retry logic

**To use it:**

1. Update `internal/api/beefy.go`:

```go
// Replace NewBeefyClient function:
func NewBeefyClient() *BeefyClient {
	config := DefaultHTTPConfig()
	config.RequestDelay = 1 * time.Second  // Delay between requests
	config.MaxRetries = 4                   // Try up to 5 times

	resilientClient := NewResilientHTTPClient(config)

	return &BeefyClient{
		httpClient: resilientClient.GetClient(),
		baseURL:    BeefyBaseURL,
	}
}
```

2. Update request logic to use `DoWithRetry`:

```go
// In GetVaults, GetAPYData, etc:
resilientClient := NewResilientHTTPClient(DefaultHTTPConfig())
resp, err := resilientClient.DoWithRetry(req)
```

### Solution B: Use Proxy Service (Residential IPs)

**Best services for API access:**

1. **Bright Data** (https://brightdata.com)
   - Residential IPs
   - $500/month (has trial)
   - Very reliable

2. **Smartproxy** (https://smartproxy.com)
   - $50/month for 5GB
   - Good for API scraping

3. **Oxylabs** (https://oxylabs.io)
   - Premium service
   - Best success rate

**Setup with proxy:**

```go
// In beefy.go or pendle.go NewClient function:
config := DefaultHTTPConfig()
config.ProxyURL = "http://username:password@proxy.example.com:8080"

resilientClient := NewResilientHTTPClient(config)
return &BeefyClient{
    httpClient: resilientClient.GetClient(),
    baseURL:    BeefyBaseURL,
}
```

Or use environment variable (no code change):
```bash
export HTTPS_PROXY=http://username:password@proxy.example.com:8080
./defirates
```

### Solution C: Use API Gateway Service

Services like **Apify** or **ScraperAPI** handle Cloudflare bypass:

```go
// Modify API base URLs:
const BeefyBaseURL = "http://api.scraperapi.com?api_key=YOUR_KEY&url=https://api.beefy.finance"
```

Cost: $29-50/month for API access tier

### Solution D: Self-Host Proxy with Rotating IPs

For tech-savvy users:

```bash
# Use rotating proxies with Tor
docker run -d -p 8118:8118 -p 9050:9050 dperson/torproxy

# Configure app to use Tor proxy
export HTTPS_PROXY=socks5://localhost:9050
./defirates
```

**Note:** Tor is slow but free. Success rate varies.

## Testing Your Solution

After implementing a solution, test if it works:

```bash
# Test Beefy API
curl -x "http://your-proxy:8080" \
  "https://api.beefy.finance/apy/breakdown" | jq '. | keys | length'

# Test Pendle API
curl -x "http://your-proxy:8080" \
  "https://api-v2.pendle.finance/api/core/v1/1/markets/active" | jq '.markets | length'

# Expected: JSON data returned (not 403 error)
```

## Recommended Approach

**For Development/Testing:**
```bash
./defirates -load-sample  # Use sample data
```

**For Production:**
1. Deploy to cloud VPS ($5-10/month) - **Simplest and most reliable**
2. If cloud IPs are blocked, add residential proxy ($50/month)
3. Monitor logs for 403s and adjust fetch interval if needed

## Why Each Solution Works

| Solution | How It Helps | Cost | Reliability |
|----------|--------------|------|-------------|
| Cloud VPS | Clean IP reputation | $5-10/mo | ⭐⭐⭐⭐⭐ |
| Home Network | Residential IP | Free | ⭐⭐⭐⭐⭐ |
| Residential Proxy | Legitimate user IPs | $50-500/mo | ⭐⭐⭐⭐⭐ |
| Retry Logic | Temporary blocks | Free | ⭐⭐⭐ |
| VPN | Different IP | $5-10/mo | ⭐⭐⭐ |
| Tor | Anonymous routing | Free | ⭐⭐ |
| API Gateway | Managed service | $30-50/mo | ⭐⭐⭐⭐ |

## Environment Variable Reference

The application respects standard Go HTTP proxy environment variables:

```bash
# HTTP proxy for all HTTP requests
export HTTP_PROXY=http://proxy.example.com:8080

# HTTPS proxy (most APIs use HTTPS)
export HTTPS_PROXY=http://proxy.example.com:8080

# SOCKS5 proxy
export HTTPS_PROXY=socks5://proxy.example.com:1080

# Proxy with authentication
export HTTPS_PROXY=http://username:password@proxy.example.com:8080

# Bypass proxy for certain domains
export NO_PROXY=localhost,127.0.0.1

# Run the application (automatically uses proxy)
./defirates
```

## Monitoring Success

After implementing a solution, check the logs:

**Success indicators:**
```
2025/10/28 16:30:10 Fetching Pendle markets...
2025/10/28 16:30:11 Successfully fetched 45 markets from Ethereum
2025/10/28 16:30:11 Successfully fetched 23 markets from Arbitrum
```

**Still blocked:**
```
Warning: failed to fetch markets for chain 1: API returned status 403: Access denied
```

## Need Help?

1. **GitHub Issues:** https://github.com/pretty-andrechal/defirates/issues
2. **Verify your IP:** https://whatismyipaddress.com/
3. **Check IP reputation:** https://www.abuseipdb.com/
4. **Test from different network first** before investing in paid solutions

## Summary

**Easiest fix:** Deploy to a cloud VPS (DigitalOcean, AWS, etc.) - 95% success rate

**Free alternatives:**
- Try from home network
- Use mobile hotspot
- Enable retry logic (included in code)

**Production solution:** Cloud VPS + residential proxy service (if needed)
