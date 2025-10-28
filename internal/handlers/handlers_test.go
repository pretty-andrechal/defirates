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

// setupTestDB creates a test database without handler
func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()

	dbPath := "test_handlers_" + t.Name() + ".db"
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db, cleanup
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
		name             string
		queryParams      string
		wantCount        int // Expected number of results
		shouldContainAPY string
		shouldNotContainAPY string
	}{
		{
			name:             "filter by asset",
			queryParams:      "?asset=ETH",
			wantCount:        1,
			shouldContainAPY: "10.00%", // ETH pool
			shouldNotContainAPY: "5.00%",  // USDC pool should be filtered out
		},
		{
			name:             "filter by chain",
			queryParams:      "?chain=Arbitrum",
			wantCount:        1,
			shouldContainAPY: "5.00%", // Arbitrum pool
			shouldNotContainAPY: "10.00%", // Ethereum pool should be filtered out
		},
		{
			name:             "filter by min APY",
			queryParams:      "?min_apy=8",
			wantCount:        1,
			shouldContainAPY: "10.00%",
			shouldNotContainAPY: "5.00%",
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

			// Check result count in the response
			expectedCountText := "Showing " + string(rune(tt.wantCount+'0')) + " yield"
			if !contains(body, expectedCountText) {
				t.Errorf("Response should show %d result(s), body doesn't contain '%s'", tt.wantCount, expectedCountText)
			}

			// Check that the expected APY is present (from included pool)
			if tt.shouldContainAPY != "" && !contains(body, tt.shouldContainAPY) {
				t.Errorf("Response should contain APY %s", tt.shouldContainAPY)
			}

			// Check that the filtered-out APY is NOT present
			if tt.shouldNotContainAPY != "" && contains(body, tt.shouldNotContainAPY) {
				t.Errorf("Response should not contain APY %s (should be filtered out)", tt.shouldNotContainAPY)
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

// TestHandleAPIRates tests the HandleAPIRates endpoint
func TestHandleAPIRates(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	h, err := New(db)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Create test data
	protocol := &models.Protocol{Name: "TestProtocol"}
	db.CreateOrUpdateProtocol(protocol)

	rate1 := &models.YieldRate{
		ProtocolID: protocol.ID,
		Asset:      "ETH",
		Chain:      "Ethereum",
		APY:        10.5,
		TVL:        1000000,
		PoolName:   "PoolA-1",
	}
	rate2 := &models.YieldRate{
		ProtocolID: protocol.ID,
		Asset:      "USDC",
		Chain:      "Arbitrum",
		APY:        5.5,
		TVL:        2000000,
		PoolName:   "PoolB-1",
	}

	db.UpsertYieldRate(rate1)
	db.UpsertYieldRate(rate2)

	tests := []struct {
		name           string
		ids            string
		wantStatus     int
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:       "fetch single rate",
			ids:        "1",
			wantStatus: http.StatusOK,
			wantContains: []string{
				"rate-1",
				"data-rate-id=\"1\"",
				"ETH",
				"10.50%",
			},
		},
		{
			name:       "fetch multiple rates",
			ids:        "1,2",
			wantStatus: http.StatusOK,
			wantContains: []string{
				"rate-1",
				"rate-2",
				"ETH",
				"USDC",
			},
		},
		{
			name:       "missing ids parameter",
			ids:        "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid ids",
			ids:        "abc,def",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "non-existent ids",
			ids:        "999,998",
			wantStatus: http.StatusOK,
			wantContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.ids != "" {
				req = httptest.NewRequest(http.MethodGet, "/api/rates?ids="+tt.ids, nil)
			} else {
				req = httptest.NewRequest(http.MethodGet, "/api/rates", nil)
			}
			w := httptest.NewRecorder()

			h.HandleAPIRates(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("HandleAPIRates() status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				body := w.Body.String()
				for _, wantStr := range tt.wantContains {
					if !contains(body, wantStr) {
						t.Errorf("Response should contain: %s", wantStr)
					}
				}
				for _, notWantStr := range tt.wantNotContain {
					if contains(body, notWantStr) {
						t.Errorf("Response should not contain: %s", notWantStr)
					}
				}

				// Verify content type
				contentType := w.Header().Get("Content-Type")
				if contentType != "text/html" {
					t.Errorf("Content-Type = %s, want text/html", contentType)
				}
			}
		})
	}
}

// TestHandleAPIRates_RowFormat tests that rows are correctly formatted
func TestHandleAPIRates_RowFormat(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	h, err := New(db)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Create test data
	protocol := &models.Protocol{Name: "Pendle"}
	db.CreateOrUpdateProtocol(protocol)

	rate := &models.YieldRate{
		ProtocolID: protocol.ID,
		Asset:      "wstETH",
		Chain:      "Ethereum",
		APY:        12.45,
		TVL:        5000000,
		PoolName:   "wstETH-Pool",
	}
	db.UpsertYieldRate(rate)

	req := httptest.NewRequest(http.MethodGet, "/api/rates?ids=1", nil)
	w := httptest.NewRecorder()

	h.HandleAPIRates(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HandleAPIRates() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()

	// Verify row structure
	requiredElements := []string{
		"<tr",
		"id=\"rate-",
		"data-rate-id=",
		"<td>",
		"</td>",
		"</tr>",
		"Pendle",
		"wstETH",
		"Ethereum",
		"12.45%",
	}

	for _, elem := range requiredElements {
		if !contains(body, elem) {
			t.Errorf("Row should contain element: %s", elem)
		}
	}
}

// TestGetEventManager tests the GetEventManager accessor
func TestGetEventManager(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	h, err := New(db)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	em := h.GetEventManager()
	if em == nil {
		t.Error("GetEventManager() should return non-nil EventManager")
	}

	// Verify it's the same instance
	em2 := h.GetEventManager()
	if em != em2 {
		t.Error("GetEventManager() should return the same instance")
	}
}

// TestHandleEvents tests the SSE endpoint
func TestHandleEvents(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	h, err := New(db)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test SSE headers
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	w := httptest.NewRecorder()

	// Run HandleEvents in a goroutine since it blocks
	done := make(chan bool)
	go func() {
		h.HandleEvents(w, req)
		done <- true
	}()

	// Give it time to set headers and send initial message
	time.Sleep(200 * time.Millisecond)

	// Cancel the request context to close the connection
	// In real testing with httptest, we can't properly test streaming,
	// so we just verify headers were set
	result := w.Result()

	// Verify SSE headers
	if ct := result.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %s, want text/event-stream", ct)
	}
	if cc := result.Header.Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("Cache-Control = %s, want no-cache", cc)
	}
	if conn := result.Header.Get("Connection"); conn != "keep-alive" {
		t.Errorf("Connection = %s, want keep-alive", conn)
	}

	// Note: We can't fully test SSE streaming with httptest.ResponseRecorder
	// as it doesn't support flushing. This test just verifies headers.
}
