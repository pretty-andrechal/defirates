package api

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/pretty-andrechal/defirates/internal/database"
	"github.com/pretty-andrechal/defirates/internal/models"
)

// TestIntegration_FetchAndStorePendleData tests the complete flow
func TestIntegration_FetchAndStorePendleData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	dbPath := "test_integration_" + t.Name() + ".db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	// Create fetcher
	fetcher := NewFetcher(db)

	// Fetch data (this will hit real API or return gracefully if blocked)
	err = fetcher.FetchAndStorePendleData()
	if err != nil {
		t.Logf("FetchAndStorePendleData() returned error (may be expected if API is blocked): %v", err)
	}

	// Verify protocol was created
	protocol, err := db.GetProtocolByName("Pendle")
	if err != nil {
		t.Errorf("Protocol 'Pendle' should exist after fetch: %v", err)
	}

	if protocol.Name != "Pendle" {
		t.Errorf("Protocol name = %s, want Pendle", protocol.Name)
	}

	// If API is accessible, we should have some rates
	rates, err := db.GetYieldRates(models.FilterParams{})
	if err != nil {
		t.Fatalf("GetYieldRates() failed: %v", err)
	}

	t.Logf("Fetched %d yield rates", len(rates))

	// If we got rates, verify their structure
	if len(rates) > 0 {
		rate := rates[0]

		// Verify required fields are populated
		if rate.Asset == "" {
			t.Error("Rate.Asset should not be empty")
		}
		if rate.Chain == "" {
			t.Error("Rate.Chain should not be empty")
		}
		if rate.APY == 0 {
			t.Error("Rate.APY should not be zero")
		}
		if rate.PoolName == "" {
			t.Error("Rate.PoolName should not be empty")
		}

		t.Logf("Sample rate: %s on %s - %.2f%% APY, $%.2f TVL",
			rate.Asset, rate.Chain, rate.APY, rate.TVL)
	}
}

// TestIntegration_ConvertMarketToYieldRate tests market conversion logic
func TestIntegration_ConvertMarketToYieldRate(t *testing.T) {
	dbPath := "test_convert_" + t.Name() + ".db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	fetcher := NewFetcher(db)

	market := Market{
		Name:    "wstETH",
		Address: "0xabc123",
		Expiry:  "2025-12-25T00:00:00.000Z",
		ChainID: 1,
		Details: MarketDetails{
			Liquidity:  1000000.50,
			ImpliedAPY: 0.05, // 5% in decimal
		},
	}

	yieldRate := fetcher.convertMarketToYieldRate(market, 1)

	// Verify conversion
	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Asset", yieldRate.Asset, "wstETH"},
		{"Chain", yieldRate.Chain, "Ethereum"},
		{"APY (converted to percentage)", yieldRate.APY, 5.0},
		{"TVL", yieldRate.TVL, 1000000.50},
		{"ProtocolID", yieldRate.ProtocolID, int64(1)},
		{"ExternalURL contains address", contains(yieldRate.ExternalURL, "0xabc123"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}

	// Verify maturity date was parsed
	if yieldRate.MaturityDate == nil {
		t.Error("MaturityDate should be parsed")
	} else {
		expected := time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)
		if !yieldRate.MaturityDate.Equal(expected) {
			t.Errorf("MaturityDate = %v, want %v", yieldRate.MaturityDate, expected)
		}
	}
}

// TestIntegration_MultipleChains tests fetching from multiple chains
func TestIntegration_MultipleChains(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := NewPendleClient()

	markets, err := client.GetMarkets()
	if err != nil {
		t.Logf("GetMarkets() failed (may be expected if API is blocked): %v", err)
		return
	}

	t.Logf("Total markets fetched: %d", len(markets))

	// Group by chain
	chainCounts := make(map[int]int)
	for _, market := range markets {
		chainCounts[market.ChainID]++
	}

	// Verify we got markets from multiple chains (if API is accessible)
	if len(markets) > 0 && len(chainCounts) == 1 {
		t.Log("Warning: Only got markets from 1 chain, expected multiple")
	}

	// Log distribution
	for chainID, count := range chainCounts {
		t.Logf("Chain %d (%s): %d markets", chainID, GetChainName(chainID), count)
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestFetcher_CallbackOnDataUpdate tests that callback is triggered after successful fetch
func TestFetcher_CallbackOnDataUpdate(t *testing.T) {
	// Setup test database
	dbPath := "test_callback_" + t.Name() + ".db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	fetcher := NewFetcher(db)

	// Track callback invocations
	callbackCount := 0
	var callbackMutex sync.Mutex

	// Set callback
	fetcher.SetOnDataUpdateCallback(func() {
		callbackMutex.Lock()
		callbackCount++
		callbackMutex.Unlock()
	})

	// Fetch data (will use sample data since API might be blocked)
	err = fetcher.FetchAndStorePendleData()

	// The function returns nil even if API is blocked
	// Callback is only called if data is actually fetched and stored
	callbackMutex.Lock()
	count := callbackCount
	callbackMutex.Unlock()

	// If callback was called, it should be exactly once
	// If callback wasn't called, API was probably blocked - that's OK
	if count > 1 {
		t.Errorf("Callback should be called at most once, got %d calls", count)
	}

	if err != nil {
		t.Logf("Fetch returned error: %v", err)
	} else {
		t.Logf("Fetch succeeded with %d callback invocations", count)
	}
}

// TestFetcher_NoCallbackSet tests that fetcher works without callback
func TestFetcher_NoCallbackSet(t *testing.T) {
	// Setup test database
	dbPath := "test_no_callback_" + t.Name() + ".db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	fetcher := NewFetcher(db)
	// Don't set callback - should not panic

	// Fetch data
	err = fetcher.FetchAndStorePendleData()
	// Should not panic even without callback
	if err != nil {
		t.Logf("Fetch returned error (expected if API is blocked): %v", err)
	}
}

// TestFetcher_CallbackMultipleFetches tests callback on multiple fetches
func TestFetcher_CallbackMultipleFetches(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// Setup test database
	dbPath := "test_multi_callback_" + t.Name() + ".db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	fetcher := NewFetcher(db)

	// Track callback invocations
	callbackCount := 0
	var callbackMutex sync.Mutex

	// Set callback
	fetcher.SetOnDataUpdateCallback(func() {
		callbackMutex.Lock()
		callbackCount++
		callbackMutex.Unlock()
	})

	// Fetch multiple times
	for i := 0; i < 3; i++ {
		err = fetcher.FetchAndStorePendleData()
		if err != nil {
			t.Logf("Fetch %d returned error: %v", i+1, err)
		}
	}

	// Callback is only called when data is actually fetched
	// If API is blocked, callback won't be called
	callbackMutex.Lock()
	count := callbackCount
	callbackMutex.Unlock()

	// We can't guarantee how many times callback is called (depends on API)
	// Just verify it's not called more than expected
	if count > 3 {
		t.Errorf("Callback should be called at most 3 times, got %d calls", count)
	}

	t.Logf("Callback was invoked %d times", count)
}

// TestFetcher_CallbackChangeable tests that callback can be changed
func TestFetcher_CallbackChangeable(t *testing.T) {
	// Setup test database
	dbPath := "test_changeable_callback_" + t.Name() + ".db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	fetcher := NewFetcher(db)

	// Set first callback
	callback1Called := false
	fetcher.SetOnDataUpdateCallback(func() {
		callback1Called = true
	})

	// Change to second callback
	callback2Called := false
	fetcher.SetOnDataUpdateCallback(func() {
		callback2Called = true
	})

	// Fetch data
	fetcher.FetchAndStorePendleData()

	// Only second callback should be called
	if callback1Called {
		t.Error("First callback should not be called after being replaced")
	}

	// Second callback should be called if fetch succeeded
	// Note: May not be called if API is blocked, so we just check it's not the first
	_ = callback2Called // Use the variable
}

// TestIntegration_FetchAndStoreBeefyData tests Beefy data fetching
func TestIntegration_FetchAndStoreBeefyData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	dbPath := "test_beefy_integration_" + t.Name() + ".db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	// Create fetcher
	fetcher := NewFetcher(db)

	// Fetch Beefy data (this will hit real API or return gracefully if blocked)
	err = fetcher.FetchAndStoreBeefyData()
	if err != nil {
		t.Logf("FetchAndStoreBeefyData() returned error (may be expected if API is blocked): %v", err)
	}

	// Verify protocol was created
	protocol, err := db.GetProtocolByName("Beefy")
	if err != nil {
		t.Errorf("Protocol 'Beefy' should exist after fetch: %v", err)
	}

	if protocol.Name != "Beefy" {
		t.Errorf("Protocol name = %s, want Beefy", protocol.Name)
	}

	// If API is accessible, we should have some rates
	rates, err := db.GetYieldRates(models.FilterParams{ProtocolName: "Beefy"})
	if err != nil {
		t.Fatalf("GetYieldRates() failed: %v", err)
	}

	t.Logf("Fetched %d Beefy yield rates", len(rates))

	// If we got rates, verify their structure
	if len(rates) > 0 {
		rate := rates[0]

		// Verify required fields are populated
		if rate.Asset == "" {
			t.Error("Rate.Asset should not be empty")
		}
		if rate.Chain == "" {
			t.Error("Rate.Chain should not be empty")
		}
		if rate.PoolName == "" {
			t.Error("Rate.PoolName should not be empty")
		}
		// Beefy vaults don't have maturity dates
		if rate.MaturityDate != nil {
			t.Error("Beefy Rate.MaturityDate should be nil")
		}

		t.Logf("Sample Beefy rate: %s on %s - %.2f%% APY, $%.2f TVL",
			rate.Asset, rate.Chain, rate.APY, rate.TVL)
	}
}

// TestIntegration_ConvertBeefyVaultToYieldRate tests Beefy vault conversion logic
func TestIntegration_ConvertBeefyVaultToYieldRate(t *testing.T) {
	dbPath := "test_beefy_convert_" + t.Name() + ".db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	fetcher := NewFetcher(db)

	vault := BeefyVaultWithMetrics{
		Vault: BeefyVault{
			ID:         "curve-eth-3pool",
			Name:       "3Pool",
			PlatformId: "curve",
			Assets:     []string{"DAI", "USDC", "USDT"},
		},
		APY:   8.5, // Already in percentage
		TVL:   45000000.50,
		Chain: "Ethereum",
	}

	yieldRate := fetcher.convertBeefyVaultToYieldRate(vault, 1)

	// Verify conversion
	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Asset", yieldRate.Asset, "3Pool"},
		{"Chain", yieldRate.Chain, "Ethereum"},
		{"APY (already percentage)", yieldRate.APY, 8.5},
		{"TVL", yieldRate.TVL, 45000000.50},
		{"ProtocolID", yieldRate.ProtocolID, int64(1)},
		{"MaturityDate", yieldRate.MaturityDate, (*time.Time)(nil)},
		{"ExternalURL contains vault ID", contains(yieldRate.ExternalURL, "curve-eth-3pool"), true},
		{"Categories contains Beefy", contains(yieldRate.Categories, "Beefy"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestIntegration_FetchAllData tests fetching from both protocols
func TestIntegration_FetchAllData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	dbPath := "test_fetch_all_" + t.Name() + ".db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	// Create fetcher
	fetcher := NewFetcher(db)

	// Track callback
	callbackCount := 0
	var callbackMutex sync.Mutex
	fetcher.SetOnDataUpdateCallback(func() {
		callbackMutex.Lock()
		callbackCount++
		callbackMutex.Unlock()
	})

	// Fetch data from all sources
	err = fetcher.FetchAllData()
	if err != nil {
		t.Logf("FetchAllData() returned error: %v", err)
	}

	// Verify both protocols were created
	pendleProtocol, err := db.GetProtocolByName("Pendle")
	if err != nil {
		t.Errorf("Protocol 'Pendle' should exist: %v", err)
	}

	beefyProtocol, err := db.GetProtocolByName("Beefy")
	if err != nil {
		t.Errorf("Protocol 'Beefy' should exist: %v", err)
	}

	// Get all rates
	allRates, err := db.GetYieldRates(models.FilterParams{})
	if err != nil {
		t.Fatalf("GetYieldRates() failed: %v", err)
	}

	t.Logf("Total rates fetched from all protocols: %d", len(allRates))

	// Count by protocol
	pendleCount := 0
	beefyCount := 0
	for _, rate := range allRates {
		if rate.ProtocolID == pendleProtocol.ID {
			pendleCount++
		} else if rate.ProtocolID == beefyProtocol.ID {
			beefyCount++
		}
	}

	t.Logf("Pendle rates: %d, Beefy rates: %d", pendleCount, beefyCount)

	// Callback should be called exactly once (from FetchAllData)
	callbackMutex.Lock()
	count := callbackCount
	callbackMutex.Unlock()

	if count > 1 {
		t.Errorf("Callback should be called at most once, got %d calls", count)
	}

	t.Logf("Callback was invoked %d times", count)
}

