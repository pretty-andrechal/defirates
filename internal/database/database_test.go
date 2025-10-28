package database

import (
	"os"
	"testing"
	"time"

	"github.com/pretty-andrechal/defirates/internal/models"
)

// setupTestDB creates a temporary test database
func setupTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	// Create temp database file
	dbPath := "test_defirates_" + t.Name() + ".db"

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db, cleanup
}

// TestNew_DatabaseCreation tests database initialization
func TestNew_DatabaseCreation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	if db == nil {
		t.Fatal("Expected database to be created")
	}

	// Verify tables exist by trying to query them
	_, err := db.GetYieldRates(models.FilterParams{})
	if err != nil {
		t.Errorf("Database schema not created properly: %v", err)
	}
}

// TestCreateOrUpdateProtocol tests protocol creation and updates
func TestCreateOrUpdateProtocol(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	protocol := &models.Protocol{
		Name:        "TestProtocol",
		URL:         "https://test.protocol",
		Description: "Test description",
	}

	// Test creation
	err := db.CreateOrUpdateProtocol(protocol)
	if err != nil {
		t.Fatalf("CreateOrUpdateProtocol() failed: %v", err)
	}

	if protocol.ID == 0 {
		t.Error("Protocol ID should be set after creation")
	}

	originalID := protocol.ID

	// Test update (same name should update, not create new)
	protocol.Description = "Updated description"
	err = db.CreateOrUpdateProtocol(protocol)
	if err != nil {
		t.Fatalf("CreateOrUpdateProtocol() update failed: %v", err)
	}

	if protocol.ID != originalID {
		t.Errorf("Protocol ID changed on update: got %d, want %d", protocol.ID, originalID)
	}

	// Test retrieval
	retrieved, err := db.GetProtocolByName("TestProtocol")
	if err != nil {
		t.Fatalf("GetProtocolByName() failed: %v", err)
	}

	if retrieved.Description != "Updated description" {
		t.Errorf("Protocol description = %s, want 'Updated description'", retrieved.Description)
	}
}

// TestUpsertYieldRate tests yield rate insertion and updates
func TestUpsertYieldRate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create protocol first
	protocol := &models.Protocol{
		Name: "TestProtocol",
	}
	db.CreateOrUpdateProtocol(protocol)

	maturityDate := time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)
	rate := &models.YieldRate{
		ProtocolID:   protocol.ID,
		Asset:        "ETH",
		Chain:        "Ethereum",
		APY:          12.5,
		TVL:          1000000.50,
		MaturityDate: &maturityDate,
		PoolName:     "ETH-Pool-1",
		ExternalURL:  "https://test.com/pool",
	}

	// Test insert
	err := db.UpsertYieldRate(rate)
	if err != nil {
		t.Fatalf("UpsertYieldRate() insert failed: %v", err)
	}

	if rate.ID == 0 {
		t.Error("YieldRate ID should be set after insert")
	}

	originalID := rate.ID

	// Test update (same protocol + pool + chain should update)
	rate.APY = 15.0
	rate.TVL = 2000000.00
	err = db.UpsertYieldRate(rate)
	if err != nil {
		t.Fatalf("UpsertYieldRate() update failed: %v", err)
	}

	// Verify it was updated, not inserted as new
	rates, err := db.GetYieldRates(models.FilterParams{})
	if err != nil {
		t.Fatalf("GetYieldRates() failed: %v", err)
	}

	if len(rates) != 1 {
		t.Errorf("Expected 1 rate after upsert, got %d", len(rates))
	}

	if rates[0].APY != 15.0 {
		t.Errorf("Rate APY = %.2f, want 15.0", rates[0].APY)
	}

	if rates[0].ID != originalID {
		t.Errorf("Rate ID changed on update: got %d, want %d", rates[0].ID, originalID)
	}
}

// TestGetYieldRates_Filtering tests various filter combinations
func TestGetYieldRates_Filtering(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Setup test data
	protocol := &models.Protocol{Name: "TestProtocol"}
	db.CreateOrUpdateProtocol(protocol)

	testRates := []models.YieldRate{
		{ProtocolID: protocol.ID, Asset: "ETH", Chain: "Ethereum", APY: 10.0, TVL: 1000000, PoolName: "Pool1-1"},
		{ProtocolID: protocol.ID, Asset: "ETH", Chain: "Arbitrum", APY: 12.0, TVL: 500000, PoolName: "Pool2-42161"},
		{ProtocolID: protocol.ID, Asset: "USDC", Chain: "Ethereum", APY: 5.0, TVL: 2000000, PoolName: "Pool3-1"},
		{ProtocolID: protocol.ID, Asset: "USDC", Chain: "Optimism", APY: 8.0, TVL: 750000, PoolName: "Pool4-10"},
	}

	for i := range testRates {
		db.UpsertYieldRate(&testRates[i])
	}

	tests := []struct {
		name    string
		filters models.FilterParams
		want    int
	}{
		{
			name:    "no filters",
			filters: models.FilterParams{},
			want:    4,
		},
		{
			name:    "filter by asset",
			filters: models.FilterParams{Asset: "ETH"},
			want:    2,
		},
		{
			name:    "filter by chain",
			filters: models.FilterParams{Chain: "Ethereum"},
			want:    2,
		},
		{
			name:    "filter by min APY",
			filters: models.FilterParams{MinAPY: 10.0},
			want:    2,
		},
		{
			name:    "filter by max APY",
			filters: models.FilterParams{MaxAPY: 10.0},
			want:    3,
		},
		{
			name:    "filter by min TVL",
			filters: models.FilterParams{MinTVL: 1000000},
			want:    2,
		},
		{
			name:    "combined filters",
			filters: models.FilterParams{Asset: "ETH", MinAPY: 11.0},
			want:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set default sort
			tt.filters.SortBy = "apy"
			tt.filters.SortOrder = "desc"

			rates, err := db.GetYieldRates(tt.filters)
			if err != nil {
				t.Fatalf("GetYieldRates() error = %v", err)
			}

			if len(rates) != tt.want {
				t.Errorf("GetYieldRates() returned %d rates, want %d", len(rates), tt.want)
			}
		})
	}
}

// TestGetYieldRates_Sorting tests sorting functionality
func TestGetYieldRates_Sorting(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	protocol := &models.Protocol{Name: "TestProtocol"}
	db.CreateOrUpdateProtocol(protocol)

	// Insert rates with different APYs and TVLs
	rates := []models.YieldRate{
		{ProtocolID: protocol.ID, Asset: "A", Chain: "Ethereum", APY: 10.0, TVL: 1000000, PoolName: "Pool1-1"},
		{ProtocolID: protocol.ID, Asset: "B", Chain: "Ethereum", APY: 15.0, TVL: 500000, PoolName: "Pool2-1"},
		{ProtocolID: protocol.ID, Asset: "C", Chain: "Ethereum", APY: 5.0, TVL: 2000000, PoolName: "Pool3-1"},
	}

	for i := range rates {
		db.UpsertYieldRate(&rates[i])
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	tests := []struct {
		name      string
		sortBy    string
		sortOrder string
		wantFirst string // Asset of first result
	}{
		{
			name:      "sort by APY desc",
			sortBy:    "apy",
			sortOrder: "desc",
			wantFirst: "B", // Highest APY
		},
		{
			name:      "sort by APY asc",
			sortBy:    "apy",
			sortOrder: "asc",
			wantFirst: "C", // Lowest APY
		},
		{
			name:      "sort by TVL desc",
			sortBy:    "tvl",
			sortOrder: "desc",
			wantFirst: "C", // Highest TVL
		},
		{
			name:      "sort by TVL asc",
			sortBy:    "tvl",
			sortOrder: "asc",
			wantFirst: "B", // Lowest TVL
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.GetYieldRates(models.FilterParams{
				SortBy:    tt.sortBy,
				SortOrder: tt.sortOrder,
			})
			if err != nil {
				t.Fatalf("GetYieldRates() error = %v", err)
			}

			if len(results) == 0 {
				t.Fatal("Expected results, got none")
			}

			if results[0].Asset != tt.wantFirst {
				t.Errorf("First result asset = %s, want %s", results[0].Asset, tt.wantFirst)
			}
		})
	}
}

// TestGetDistinctAssets tests unique asset retrieval
func TestGetDistinctAssets(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	protocol := &models.Protocol{Name: "TestProtocol"}
	db.CreateOrUpdateProtocol(protocol)

	// Insert rates with duplicate assets
	assets := []string{"ETH", "USDC", "ETH", "DAI", "USDC"}
	for i, asset := range assets {
		rate := &models.YieldRate{
			ProtocolID: protocol.ID,
			Asset:      asset,
			Chain:      "Ethereum",
			APY:        5.0,
			TVL:        1000000,
			PoolName:   "Pool" + string(rune(i+'0')) + "-1",
		}
		db.UpsertYieldRate(rate)
	}

	distinctAssets, err := db.GetDistinctAssets()
	if err != nil {
		t.Fatalf("GetDistinctAssets() error = %v", err)
	}

	if len(distinctAssets) != 3 {
		t.Errorf("Expected 3 distinct assets, got %d: %v", len(distinctAssets), distinctAssets)
	}

	// Verify they're sorted
	expectedOrder := []string{"DAI", "ETH", "USDC"}
	for i, expected := range expectedOrder {
		if distinctAssets[i] != expected {
			t.Errorf("Asset[%d] = %s, want %s", i, distinctAssets[i], expected)
		}
	}
}

// TestGetDistinctChains tests unique chain retrieval
func TestGetDistinctChains(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	protocol := &models.Protocol{Name: "TestProtocol"}
	db.CreateOrUpdateProtocol(protocol)

	chains := []string{"Ethereum", "Arbitrum", "Ethereum", "Base"}
	for i, chain := range chains {
		rate := &models.YieldRate{
			ProtocolID: protocol.ID,
			Asset:      "ETH",
			Chain:      chain,
			APY:        5.0,
			TVL:        1000000,
			PoolName:   "Pool" + string(rune(i+'0')) + "-1",
		}
		db.UpsertYieldRate(rate)
	}

	distinctChains, err := db.GetDistinctChains()
	if err != nil {
		t.Fatalf("GetDistinctChains() error = %v", err)
	}

	if len(distinctChains) != 3 {
		t.Errorf("Expected 3 distinct chains, got %d: %v", len(distinctChains), distinctChains)
	}
}

// TestGetYieldRatesByIDs tests fetching rates by specific IDs
func TestGetYieldRatesByIDs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a protocol
	protocol := &models.Protocol{Name: "TestProtocol"}
	if err := db.CreateOrUpdateProtocol(protocol); err != nil {
		t.Fatalf("Failed to create protocol: %v", err)
	}

	// Create multiple yield rates
	var rateIDs []int64
	for i := 0; i < 5; i++ {
		rate := &models.YieldRate{
			ProtocolID: protocol.ID,
			Asset:      "ETH",
			Chain:      "Ethereum",
			APY:        float64(5 + i),
			TVL:        float64(1000000 * (i + 1)),
			PoolName:   "Pool" + string(rune(i+'A')) + "-1",
		}
		if err := db.UpsertYieldRate(rate); err != nil {
			t.Fatalf("Failed to create rate: %v", err)
		}
		rateIDs = append(rateIDs, rate.ID)
	}

	tests := []struct {
		name     string
		ids      []int64
		wantLen  int
		wantErr  bool
	}{
		{
			name:    "fetch single rate",
			ids:     []int64{rateIDs[0]},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "fetch multiple rates",
			ids:     []int64{rateIDs[0], rateIDs[2], rateIDs[4]},
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "fetch all rates",
			ids:     rateIDs,
			wantLen: 5,
			wantErr: false,
		},
		{
			name:    "fetch non-existent rate",
			ids:     []int64{99999},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "fetch mix of existing and non-existent",
			ids:     []int64{rateIDs[0], 99999, rateIDs[1]},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "empty ID list",
			ids:     []int64{},
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rates, err := db.GetYieldRatesByIDs(tt.ids)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetYieldRatesByIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(rates) != tt.wantLen {
				t.Errorf("GetYieldRatesByIDs() returned %d rates, want %d", len(rates), tt.wantLen)
			}

			// Verify returned rates have correct IDs
			if len(tt.ids) > 0 && len(rates) > 0 {
				idMap := make(map[int64]bool)
				for _, id := range tt.ids {
					idMap[id] = true
				}

				for _, rate := range rates {
					if !idMap[rate.ID] {
						t.Errorf("Returned rate ID %d was not in requested IDs", rate.ID)
					}
					// Verify protocol name is populated (from JOIN)
					if rate.ProtocolName == "" {
						t.Error("ProtocolName should be populated from JOIN")
					}
				}
			}
		})
	}
}

// TestGetYieldRatesByIDs_Order tests that results maintain database order
func TestGetYieldRatesByIDs_Order(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create protocol
	protocol := &models.Protocol{Name: "TestProtocol"}
	if err := db.CreateOrUpdateProtocol(protocol); err != nil {
		t.Fatalf("Failed to create protocol: %v", err)
	}

	// Create rates with different APYs
	rate1 := &models.YieldRate{
		ProtocolID: protocol.ID,
		Asset:      "ETH",
		Chain:      "Ethereum",
		APY:        10.0,
		TVL:        1000000,
		PoolName:   "PoolA-1",
	}
	rate2 := &models.YieldRate{
		ProtocolID: protocol.ID,
		Asset:      "ETH",
		Chain:      "Ethereum",
		APY:        15.0,
		TVL:        2000000,
		PoolName:   "PoolB-1",
	}
	rate3 := &models.YieldRate{
		ProtocolID: protocol.ID,
		Asset:      "ETH",
		Chain:      "Ethereum",
		APY:        5.0,
		TVL:        3000000,
		PoolName:   "PoolC-1",
	}

	db.UpsertYieldRate(rate1)
	db.UpsertYieldRate(rate2)
	db.UpsertYieldRate(rate3)

	// Fetch in specific order
	rates, err := db.GetYieldRatesByIDs([]int64{rate3.ID, rate1.ID, rate2.ID})
	if err != nil {
		t.Fatalf("GetYieldRatesByIDs() error = %v", err)
	}

	if len(rates) != 3 {
		t.Fatalf("Expected 3 rates, got %d", len(rates))
	}

	// Verify we got all the rates (order may vary based on DB)
	foundIDs := make(map[int64]bool)
	for _, rate := range rates {
		foundIDs[rate.ID] = true
	}

	if !foundIDs[rate1.ID] || !foundIDs[rate2.ID] || !foundIDs[rate3.ID] {
		t.Error("Not all requested rates were returned")
	}
}

// TestGetYieldRatesByIDs_WithMaturityDate tests fetching rates with maturity dates
func TestGetYieldRatesByIDs_WithMaturityDate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	protocol := &models.Protocol{Name: "TestProtocol"}
	if err := db.CreateOrUpdateProtocol(protocol); err != nil {
		t.Fatalf("Failed to create protocol: %v", err)
	}

	maturityDate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	rate := &models.YieldRate{
		ProtocolID:   protocol.ID,
		Asset:        "ETH",
		Chain:        "Ethereum",
		APY:          10.0,
		TVL:          1000000,
		MaturityDate: &maturityDate,
		PoolName:     "MaturityPool-1",
	}

	if err := db.UpsertYieldRate(rate); err != nil {
		t.Fatalf("Failed to create rate: %v", err)
	}

	rates, err := db.GetYieldRatesByIDs([]int64{rate.ID})
	if err != nil {
		t.Fatalf("GetYieldRatesByIDs() error = %v", err)
	}

	if len(rates) != 1 {
		t.Fatalf("Expected 1 rate, got %d", len(rates))
	}

	if rates[0].MaturityDate == nil {
		t.Fatal("MaturityDate should not be nil")
	}

	if !rates[0].MaturityDate.Equal(maturityDate) {
		t.Errorf("MaturityDate = %v, want %v", rates[0].MaturityDate, maturityDate)
	}
}
