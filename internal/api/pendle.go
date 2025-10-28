package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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

// Market represents a Pendle market
type Market struct {
	Address      string       `json:"address"`
	ChainID      int          `json:"chainId"`
	Symbol       string       `json:"symbol"`
	Name         string       `json:"name"`
	Expiry       string       `json:"expiry"`
	PT           TokenInfo    `json:"pt"`
	SY           TokenInfo    `json:"sy"`
	YT           TokenInfo    `json:"yt"`
	Underlyings  []TokenInfo  `json:"underlyings"`
	ImpliedAPY   float64      `json:"impliedApy"`
	Liquidity    Liquidity    `json:"liquidity"`
	TotalPT      string       `json:"totalPt"`
	TotalSY      string       `json:"totalSy"`
	Aggregated   *Aggregated  `json:"aggregated,omitempty"`
}

type TokenInfo struct {
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
	Name     string `json:"name"`
}

type Liquidity struct {
	USD float64 `json:"usd"`
}

type Aggregated struct {
	Total string `json:"total"`
}

// MarketsResponse is the response from the markets endpoint
type MarketsResponse struct {
	Results []Market `json:"results"`
	Total   int      `json:"total"`
	Limit   int      `json:"limit"`
	Skip    int      `json:"skip"`
}

// ChainIDToName converts Pendle chain IDs to readable names
var ChainIDToName = map[int]string{
	1:     "Ethereum",
	42161: "Arbitrum",
	10:    "Optimism",
	8453:  "Base",
	56:    "BSC",
	137:   "Polygon",
	59144: "Linea",
	534352: "Scroll",
}

// GetMarkets fetches all active markets from Pendle across all supported chains
func (c *PendleClient) GetMarkets() ([]Market, error) {
	var allMarkets []Market

	// Fetch markets from each supported chain
	chainIDs := []int{1, 42161, 10, 8453, 56, 137}

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

	req.Header.Set("Accept", "application/json")

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

	return marketsResp.Results, nil
}

// GetActiveMarkets fetches only active (non-expired) markets
func (c *PendleClient) GetActiveMarkets() ([]Market, error) {
	allMarkets, err := c.GetMarkets()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var activeMarkets []Market

	for _, market := range allMarkets {
		// Parse expiry date
		expiry, err := time.Parse("2006-01-02T15:04:05.000Z", market.Expiry)
		if err != nil {
			// Try alternative format
			expiry, err = time.Parse(time.RFC3339, market.Expiry)
			if err != nil {
				// Skip markets with unparseable expiry
				continue
			}
		}

		// Only include markets that haven't expired yet
		if expiry.After(now) {
			activeMarkets = append(activeMarkets, market)
		}
	}

	return activeMarkets, nil
}

// GetChainName returns the human-readable chain name for a chain ID
func GetChainName(chainID int) string {
	if name, ok := ChainIDToName[chainID]; ok {
		return name
	}
	return fmt.Sprintf("Chain-%d", chainID)
}
