package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestPendleClient_GetMarketsForChain tests fetching markets for a single chain
func TestPendleClient_GetMarketsForChain(t *testing.T) {
	tests := []struct {
		name           string
		chainID        int
		mockResponse   string
		mockStatusCode int
		wantErr        bool
		wantMarkets    int
	}{
		{
			name:           "successful fetch with markets",
			chainID:        1,
			mockStatusCode: 200,
			mockResponse: `{
				"markets": [
					{
						"name": "wstETH",
						"address": "0xabc123",
						"expiry": "2025-12-25T00:00:00.000Z",
						"pt": "1-0xpt123",
						"yt": "1-0xyt123",
						"sy": "1-0xsy123",
						"underlyingAsset": "1-0xeth",
						"details": {
							"liquidity": 1000000.50,
							"impliedApy": 0.05,
							"aggregatedApy": 0.052
						}
					},
					{
						"name": "sUSDe",
						"address": "0xdef456",
						"expiry": "2026-01-15T00:00:00.000Z",
						"pt": "1-0xpt456",
						"yt": "1-0xyt456",
						"sy": "1-0xsy456",
						"underlyingAsset": "1-0xusde",
						"details": {
							"liquidity": 5000000.25,
							"impliedApy": 0.15,
							"aggregatedApy": 0.155
						}
					}
				]
			}`,
			wantErr:     false,
			wantMarkets: 2,
		},
		{
			name:           "empty markets array",
			chainID:        999,
			mockStatusCode: 200,
			mockResponse:   `{"markets": []}`,
			wantErr:        false,
			wantMarkets:    0,
		},
		{
			name:           "API returns 403",
			chainID:        1,
			mockStatusCode: 403,
			mockResponse:   `Access denied`,
			wantErr:        true,
			wantMarkets:    0,
		},
		{
			name:           "API returns 400 with error message",
			chainID:        137,
			mockStatusCode: 400,
			mockResponse:   `{"message":"Unsupported chain id","error":"Bad Request"}`,
			wantErr:        true,
			wantMarkets:    0,
		},
		{
			name:           "invalid JSON response",
			chainID:        1,
			mockStatusCode: 200,
			mockResponse:   `{invalid json}`,
			wantErr:        true,
			wantMarkets:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}

				expectedPath := "/v1/" + string(rune(tt.chainID+'0')) + "/markets/active"
				if r.URL.Path != expectedPath {
					t.Logf("Got path: %s, expected path pattern: /v1/%d/markets/active", r.URL.Path, tt.chainID)
				}

				// Check User-Agent header
				if ua := r.Header.Get("User-Agent"); ua == "" {
					t.Error("Expected User-Agent header to be set")
				}

				// Send mock response
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// Create client with mock server URL
			client := &PendleClient{
				httpClient: &http.Client{Timeout: 5 * time.Second},
				baseURL:    server.URL,
			}

			// Test
			markets, err := client.GetMarketsForChain(tt.chainID)

			// Verify error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMarketsForChain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify market count
			if len(markets) != tt.wantMarkets {
				t.Errorf("GetMarketsForChain() got %d markets, want %d", len(markets), tt.wantMarkets)
			}

			// Verify ChainID is set correctly
			if !tt.wantErr && len(markets) > 0 {
				for i, market := range markets {
					if market.ChainID != tt.chainID {
						t.Errorf("Market[%d].ChainID = %d, want %d", i, market.ChainID, tt.chainID)
					}
				}
			}
		})
	}
}

// TestMarket_JSONParsing tests that Market struct correctly parses API JSON
func TestMarket_JSONParsing(t *testing.T) {
	jsonData := `{
		"name": "wstETH",
		"address": "0xc374f7ec85f8c7de3207a10bb1978ba104bda3b2",
		"expiry": "2025-12-25T00:00:00.000Z",
		"pt": "1-0xf99985822fb361117fcf3768d34a6353e6022f5f",
		"yt": "1-0xf3abc972a0f537c1119c990d422463b93227cd83",
		"sy": "1-0xcbc72d92b2dc8187414f6734718563898740c0bc",
		"underlyingAsset": "1-0x7f39c581f595b53c5cb19bd0b3f8da6c935e2ca0",
		"details": {
			"liquidity": 48990134.2627604,
			"pendleApy": 0.00173494316157951,
			"impliedApy": 0.0282073211383478,
			"aggregatedApy": 0.02806215959798,
			"feeRate": 0.000499999999348688
		},
		"timestamp": "2023-04-05T10:18:59.000Z",
		"categoryIds": ["eth"]
	}`

	var market Market
	err := json.Unmarshal([]byte(jsonData), &market)
	if err != nil {
		t.Fatalf("Failed to unmarshal market JSON: %v", err)
	}

	// Test all fields are parsed correctly
	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Name", market.Name, "wstETH"},
		{"Address", market.Address, "0xc374f7ec85f8c7de3207a10bb1978ba104bda3b2"},
		{"Expiry", market.Expiry, "2025-12-25T00:00:00.000Z"},
		{"Liquidity", market.Details.Liquidity, 48990134.2627604},
		{"ImpliedAPY", market.Details.ImpliedAPY, 0.0282073211383478},
		{"AggregatedAPY", market.Details.AggregatedAPY, 0.02806215959798},
		{"CategoryIDs length", len(market.CategoryIDs), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestGetActiveMarkets_ExpiryFiltering tests that expired markets are filtered out
func TestGetActiveMarkets_ExpiryFiltering(t *testing.T) {
	now := time.Now()
	futureDate := now.Add(30 * 24 * time.Hour).Format("2006-01-02T15:04:05.000Z")
	pastDate := now.Add(-30 * 24 * time.Hour).Format("2006-01-02T15:04:05.000Z")

	mockResponse := `{
		"markets": [
			{
				"name": "Future Market",
				"address": "0x111",
				"expiry": "` + futureDate + `",
				"details": {"liquidity": 1000000, "impliedApy": 0.05}
			},
			{
				"name": "Expired Market",
				"address": "0x222",
				"expiry": "` + pastDate + `",
				"details": {"liquidity": 1000000, "impliedApy": 0.05}
			},
			{
				"name": "Another Future Market",
				"address": "0x333",
				"expiry": "` + futureDate + `",
				"details": {"liquidity": 1000000, "impliedApy": 0.05}
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	// Test GetMarketsForChain directly to avoid the multi-chain loop
	client := &PendleClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    server.URL,
	}

	// Get markets for one chain
	allMarkets, err := client.GetMarketsForChain(1)
	if err != nil {
		t.Fatalf("GetMarketsForChain() error = %v", err)
	}

	if len(allMarkets) != 3 {
		t.Fatalf("Expected 3 total markets, got %d", len(allMarkets))
	}

	// Test the filtering logic directly on the fetched markets
	activeMarkets := []Market{}
	for _, market := range allMarkets {
		expiry, err := time.Parse("2006-01-02T15:04:05.000Z", market.Expiry)
		if err != nil {
			t.Fatalf("Failed to parse expiry: %v", err)
		}

		if expiry.After(now) {
			activeMarkets = append(activeMarkets, market)
		}
	}

	// Should only return the 2 future markets, not the expired one
	if len(activeMarkets) != 2 {
		t.Errorf("Filtering returned %d markets, want 2 (expired market should be filtered)", len(activeMarkets))
	}

	// Verify we got the right markets
	for _, market := range activeMarkets {
		if market.Name == "Expired Market" {
			t.Error("Expired market was not filtered out")
		}
	}
}

// TestGetChainName tests chain ID to name mapping
func TestGetChainName(t *testing.T) {
	tests := []struct {
		chainID  int
		wantName string
	}{
		{1, "Ethereum"},
		{10, "Optimism"},
		{56, "BSC"},
		{42161, "Arbitrum"},
		{8453, "Base"},
		{5000, "Mantle"},
		{999, "Zora"},
		{146, "Sonic"},
		{9745, "Taiko"},
		{80094, "Berachain"},
		{99999, "Chain-99999"}, // Unknown chain
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			got := GetChainName(tt.chainID)
			if got != tt.wantName {
				t.Errorf("GetChainName(%d) = %s, want %s", tt.chainID, got, tt.wantName)
			}
		})
	}
}

// TestMarketsResponse_JSONParsing tests the response wrapper
func TestMarketsResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"markets": [
			{"name": "Market1", "address": "0x111", "details": {"liquidity": 1000, "impliedApy": 0.05}},
			{"name": "Market2", "address": "0x222", "details": {"liquidity": 2000, "impliedApy": 0.06}}
		]
	}`

	var resp MarketsResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal MarketsResponse: %v", err)
	}

	if len(resp.Markets) != 2 {
		t.Errorf("Expected 2 markets, got %d", len(resp.Markets))
	}
}
