package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/pretty-andrechal/defirates/internal/database"
	"github.com/pretty-andrechal/defirates/internal/models"
)

// setupTestHandler creates a handler with test database
func setupTestHandler(t *testing.T) (*Handler, *database.DB, func()) {
	t.Helper()

	dbPath := "test_handlers_" + t.Name() + ".db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	handler, err := New(db)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return handler, db, cleanup
}

// TestHandleIndex_EmptyDatabase tests index page with no data
func TestHandleIndex_EmptyDatabase(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.HandleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HandleIndex() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("HandleIndex() returned empty body")
	}

	// Should contain the "no results" message
	if !contains(body, "No yield opportunities found") && !contains(body, "Showing 0 yield opportunities") {
		t.Error("Expected 'no results' message in response")
	}
}

// TestHandleIndex_WithData tests index page with yield rates
func TestHandleIndex_WithData(t *testing.T) {
	handler, db, cleanup := setupTestHandler(t)
	defer cleanup()

	// Add test data
	protocol := &models.Protocol{Name: "TestProtocol"}
	db.CreateOrUpdateProtocol(protocol)

	rate := &models.YieldRate{
		ProtocolID: protocol.ID,
		Asset:      "ETH",
		Chain:      "Ethereum",
		APY:        12.5,
		TVL:        1000000,
		PoolName:   "TestPool-1",
	}
	db.UpsertYieldRate(rate)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.HandleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HandleIndex() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()

	// Verify data is rendered
	tests := []string{
		"ETH",           // Asset
		"Ethereum",      // Chain
		"12.50%",        // APY
		"TestProtocol",  // Protocol name
	}

	for _, want := range tests {
		if !contains(body, want) {
			t.Errorf("HandleIndex() response missing expected content: %s", want)
		}
	}
}

// TestHandleIndex_Filtering tests query parameter filtering
func TestHandleIndex_Filtering(t *testing.T) {
	handler, db, cleanup := setupTestHandler(t)
	defer cleanup()

	// Add test data
	protocol := &models.Protocol{Name: "TestProtocol"}
	db.CreateOrUpdateProtocol(protocol)

	rates := []models.YieldRate{
		{ProtocolID: protocol.ID, Asset: "ETH", Chain: "Ethereum", APY: 10.0, TVL: 1000000, PoolName: "Pool1-1"},
		{ProtocolID: protocol.ID, Asset: "USDC", Chain: "Arbitrum", APY: 5.0, TVL: 500000, PoolName: "Pool2-42161"},
	}

	for i := range rates {
		db.UpsertYieldRate(&rates[i])
	}

	tests := []struct {
		name         string
		queryParams  string
		shouldContain string
		shouldNotContain string
	}{
		{
			name:         "filter by asset",
			queryParams:  "?asset=ETH",
			shouldContain: "ETH",
			shouldNotContain: "USDC",
		},
		{
			name:         "filter by chain",
			queryParams:  "?chain=Arbitrum",
			shouldContain: "Arbitrum",
			shouldNotContain: "Ethereum",
		},
		{
			name:         "filter by min APY",
			queryParams:  "?min_apy=8",
			shouldContain: "10.00%",
			shouldNotContain: "5.00%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.HandleIndex(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("HandleIndex() status = %d, want %d", w.Code, http.StatusOK)
			}

			body := w.Body.String()

			if tt.shouldContain != "" && !contains(body, tt.shouldContain) {
				t.Errorf("Response should contain %s", tt.shouldContain)
			}

			if tt.shouldNotContain != "" && contains(body, tt.shouldNotContain) {
				t.Errorf("Response should not contain %s", tt.shouldNotContain)
			}
		})
	}
}

// TestHandleIndex_HTMX tests HTMX partial response
func TestHandleIndex_HTMX(t *testing.T) {
	handler, db, cleanup := setupTestHandler(t)
	defer cleanup()

	protocol := &models.Protocol{Name: "TestProtocol"}
	db.CreateOrUpdateProtocol(protocol)

	rate := &models.YieldRate{
		ProtocolID: protocol.ID,
		Asset:      "ETH",
		Chain:      "Ethereum",
		APY:        10.0,
		TVL:        1000000,
		PoolName:   "TestPool-1",
	}
	db.UpsertYieldRate(rate)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	handler.HandleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HandleIndex() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()

	// HTMX request should return table partial, not full page
	if contains(body, "<html>") || contains(body, "<!DOCTYPE") {
		t.Error("HTMX request should not return full HTML page")
	}

	// Should still contain table data
	if !contains(body, "ETH") {
		t.Error("HTMX response should contain table data")
	}
}

// TestHandleIndex_TemplateTypeComparisons tests that template renders without type errors
func TestHandleIndex_TemplateTypeComparisons(t *testing.T) {
	handler, db, cleanup := setupTestHandler(t)
	defer cleanup()

	protocol := &models.Protocol{Name: "TestProtocol"}
	db.CreateOrUpdateProtocol(protocol)

	// Test with various TVL values to ensure template comparisons work
	tvlValues := []float64{500.0, 5000.0, 500000.0, 5000000.0}

	for i, tvl := range tvlValues {
		rate := &models.YieldRate{
			ProtocolID: protocol.ID,
			Asset:      "ETH",
			Chain:      "Ethereum",
			APY:        10.0,
			TVL:        tvl,
			PoolName:   "Pool" + string(rune(i+'0')) + "-1",
		}
		db.UpsertYieldRate(rate)
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.HandleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HandleIndex() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()

	// Check for proper TVL formatting (K, M suffixes)
	if !contains(body, "K") && !contains(body, "M") {
		t.Error("TVL formatting not working - should contain K or M suffixes")
	}
}

// TestHandleIndex_MaturityDateFormatting tests date formatting in templates
func TestHandleIndex_MaturityDateFormatting(t *testing.T) {
	handler, db, cleanup := setupTestHandler(t)
	defer cleanup()

	protocol := &models.Protocol{Name: "TestProtocol"}
	db.CreateOrUpdateProtocol(protocol)

	maturityDate := time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)
	rate := &models.YieldRate{
		ProtocolID:   protocol.ID,
		Asset:        "ETH",
		Chain:        "Ethereum",
		APY:          10.0,
		TVL:          1000000,
		MaturityDate: &maturityDate,
		PoolName:     "TestPool-1",
	}
	db.UpsertYieldRate(rate)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.HandleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HandleIndex() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()

	// Check for formatted date (Jan 02, 2006 format)
	if !contains(body, "Dec 25, 2025") {
		t.Error("Maturity date not formatted correctly")
	}
}

// TestHandleIndex_APYColorCoding tests APY color class assignment
func TestHandleIndex_APYColorCoding(t *testing.T) {
	handler, db, cleanup := setupTestHandler(t)
	defer cleanup()

	protocol := &models.Protocol{Name: "TestProtocol"}
	db.CreateOrUpdateProtocol(protocol)

	// Test different APY levels
	rates := []struct {
		apy   float64
		class string
	}{
		{15.0, "apy-high"},   // >= 10
		{7.0, "apy-medium"},  // >= 5 but < 10
		{2.0, "apy-low"},     // < 5
	}

	for i, r := range rates {
		rate := &models.YieldRate{
			ProtocolID: protocol.ID,
			Asset:      "ETH",
			Chain:      "Ethereum",
			APY:        r.apy,
			TVL:        1000000,
			PoolName:   "Pool" + string(rune(i+'0')) + "-1",
		}
		db.UpsertYieldRate(rate)
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.HandleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HandleIndex() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()

	// Verify all color classes are present
	for _, r := range rates {
		if !contains(body, r.class) {
			t.Errorf("Response should contain APY color class: %s", r.class)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsRec(s, substr))
}

func containsRec(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
