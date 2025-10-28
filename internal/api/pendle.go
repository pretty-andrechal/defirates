package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pretty-andrechal/defirates/internal/database"
)

const (
	PendleBaseURL = "https://api-v2.pendle.finance/api/core"
)

// PendleClient handles communication with Pendle API
type PendleClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewPendleClient creates a new Pendle API client
func NewPendleClient() *PendleClient {
	return &PendleClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: PendleBaseURL,
	}
}

// NewPendleClientWithDebug creates a new Pendle API client with debug logging
func NewPendleClientWithDebug(db *database.DB) *PendleClient {
	baseClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Wrap with debug client
	debugClient := NewDebugHTTPClient(baseClient, db, "pendle", true)
	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: &debugHTTPTransport{debugClient: debugClient},
	}

	return &PendleClient{
		httpClient: httpClient,
		baseURL:    PendleBaseURL,
	}
}

// Market represents a Pendle market (matching actual API response)
type Market struct {
	Name            string         `json:"name"`
	Address         string         `json:"address"`
	Expiry          string         `json:"expiry"`
	PT              string         `json:"pt"`
	YT              string         `json:"yt"`
	SY              string         `json:"sy"`
	UnderlyingAsset string         `json:"underlyingAsset"`
	Details         MarketDetails  `json:"details"`
	Timestamp       string         `json:"timestamp"`
	CategoryIDs     []string       `json:"categoryIds"`
	ChainID         int            `json:"-"` // Not in API response, set manually
}

// MarketDetails contains the nested details from API response
type MarketDetails struct {
	Liquidity    float64 `json:"liquidity"`
	PendleAPY    float64 `json:"pendleApy"`
	ImpliedAPY   float64 `json:"impliedApy"`
	AggregatedAPY float64 `json:"aggregatedApy"`
	FeeRate      float64 `json:"feeRate"`
}

// MarketsResponse is the response from the markets endpoint
type MarketsResponse struct {
	Markets []Market `json:"markets"`
}

// ChainIDToName converts Pendle chain IDs to readable names
var ChainIDToName = map[int]string{
	1:     "Ethereum",
	10:    "Optimism",
	56:    "BSC",
	146:   "Sonic",
	999:   "Zora",
	5000:  "Mantle",
	8453:  "Base",
	9745:  "Taiko",
	42161: "Arbitrum",
	80094: "Berachain",
}

// GetMarkets fetches all active markets from Pendle across all supported chains
func (c *PendleClient) GetMarkets() ([]Market, error) {
	var allMarkets []Market

	// Fetch markets from each supported chain (as of API response)
	// Supported chains: 1, 10, 56, 146, 999, 5000, 8453, 9745, 42161, 80094
	chainIDs := []int{1, 10, 56, 146, 999, 5000, 8453, 9745, 42161, 80094}

	for _, chainID := range chainIDs {
		markets, err := c.GetMarketsForChain(chainID)
		if err != nil {
			// Log error but continue with other chains
			fmt.Printf("Warning: failed to fetch markets for chain %d: %v\n", chainID, err)
			continue
		}
		allMarkets = append(allMarkets, markets...)
	}

	if len(allMarkets) == 0 {
		return nil, fmt.Errorf("no markets fetched from any chain")
	}

	return allMarkets, nil
}

// GetMarketsForChain fetches active markets for a specific chain
func (c *PendleClient) GetMarketsForChain(chainID int) ([]Market, error) {
	url := fmt.Sprintf("%s/v1/%d/markets/active", c.baseURL, chainID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers - User-Agent is important for some APIs/WAFs
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	// Note: Don't set Accept-Encoding - Go's http client handles gzip automatically
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Origin", "https://app.pendle.finance")
	req.Header.Set("Referer", "https://app.pendle.finance/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch markets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var marketsResp MarketsResponse
	if err := json.Unmarshal(body, &marketsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Set ChainID on each market since it's not in the API response
	for i := range marketsResp.Markets {
		marketsResp.Markets[i].ChainID = chainID
	}

	return marketsResp.Markets, nil
}

// GetActiveMarkets fetches only active (non-expired) markets
func (c *PendleClient) GetActiveMarkets() ([]Market, error) {
	allMarkets, err := c.GetMarkets()
	if err != nil {
		return nil, err
	}

	fmt.Printf("DEBUG: GetActiveMarkets received %d markets total\n", len(allMarkets))

	now := time.Now()
	var activeMarkets []Market
	skippedCount := 0
	expiredCount := 0

	for i, market := range allMarkets {
		// Parse expiry date
		expiry, err := time.Parse("2006-01-02T15:04:05.000Z", market.Expiry)
		if err != nil {
			// Try alternative format
			expiry, err = time.Parse(time.RFC3339, market.Expiry)
			if err != nil {
				// Log the first few unparseable dates to debug
				if skippedCount < 3 {
					fmt.Printf("DEBUG: [%d] Skipping %s - unparseable expiry: '%s'\n", i, market.Name, market.Expiry)
				}
				skippedCount++
				// Skip markets with unparseable expiry
				continue
			}
		}

		// Only include markets that haven't expired yet
		if expiry.After(now) {
			activeMarkets = append(activeMarkets, market)
		} else {
			if expiredCount < 3 {
				fmt.Printf("DEBUG: [%d] Skipping %s - expired on %s\n", i, market.Name, expiry.Format("2006-01-02"))
			}
			expiredCount++
		}
	}

	fmt.Printf("DEBUG: Result - %d active, %d expired, %d unparseable\n", len(activeMarkets), expiredCount, skippedCount)

	return activeMarkets, nil
}

// GetChainName returns the human-readable chain name for a chain ID
func GetChainName(chainID int) string {
	if name, ok := ChainIDToName[chainID]; ok {
		return name
	}
	return fmt.Sprintf("Chain-%d", chainID)
}
