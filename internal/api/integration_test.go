package api

import (
	"os"
	"testing"
	"time"

	"github.com/pretty-andrechal/defirates/internal/database"
	"github.com/pretty-andrechal/defirates/internal/models"
)

// TestFetcher_FetchAndStorePendleData tests the complete flow
func TestFetcher_FetchAndStorePendleData(t *testing.T) {
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

// TestConvertMarketToYieldRate tests market conversion logic
func TestConvertMarketToYieldRate(t *testing.T) {
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

// TestFetcher_MultipleChains tests fetching from multiple chains
func TestFetcher_MultipleChains(t *testing.T) {
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
