package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	BeefyBaseURL = "https://api.beefy.finance"
)

// BeefyClient handles communication with Beefy Finance API
type BeefyClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewBeefyClient creates a new Beefy API client
func NewBeefyClient() *BeefyClient {
	return &BeefyClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: BeefyBaseURL,
	}
}

// BeefyVault represents a Beefy vault from the API
type BeefyVault struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Token               string   `json:"token"`
	TokenAddress        string   `json:"tokenAddress"`
	TokenDecimals       int      `json:"tokenDecimals"`
	TokenProviderId     string   `json:"tokenProviderId"`
	EarnedToken         string   `json:"earnedToken"`
	EarnedTokenAddress  string   `json:"earnedTokenAddress"`
	EarnContractAddress string   `json:"earnContractAddress"`
	Oracle              string   `json:"oracle"`
	OracleId            string   `json:"oracleId"`
	Status              string   `json:"status"`
	PlatformId          string   `json:"platformId"`
	Assets              []string `json:"assets"`
	StrategyTypeId      string   `json:"strategyTypeId"`
	Risks               []string `json:"risks"`
	Chain               string   `json:"chain"` // May be in some responses or derived from endpoint
}

// BeefyAPYData represents APY data from the breakdown endpoint
type BeefyAPYBreakdown struct {
	TotalApy            float64 `json:"totalApy"`
	VaultApr            float64 `json:"vaultApr"`
	CompoundingsPerYear int     `json:"compoundingsPerYear"`
	BeefyPerformanceFee float64 `json:"beefyPerformanceFee"`
	VaultApy            float64 `json:"vaultApy"`
	LpFee               float64 `json:"lpFee"`
	TradingApr          float64 `json:"tradingApr"`
}

// BeefyVaultWithMetrics combines vault info with APY and TVL
type BeefyVaultWithMetrics struct {
	Vault BeefyVault
	APY   float64
	TVL   float64
	Chain string
}

// SupportedChains lists the chains Beefy supports
var BeefySupportedChains = []string{
	"arbitrum", "aurora", "avax", "base", "bsc", "canto", "celo",
	"cronos", "emerald", "ethereum", "fantom", "fuse", "harmony",
	"heco", "kava", "metis", "moonbeam", "moonriver", "optimism",
	"polygon", "zkevm", "zksync",
}

// ChainNameMapping maps Beefy chain IDs to human-readable names
var BeefyChainNameMapping = map[string]string{
	"arbitrum":  "Arbitrum",
	"aurora":    "Aurora",
	"avax":      "Avalanche",
	"base":      "Base",
	"bsc":       "BSC",
	"canto":     "Canto",
	"celo":      "Celo",
	"cronos":    "Cronos",
	"emerald":   "Emerald",
	"ethereum":  "Ethereum",
	"fantom":    "Fantom",
	"fuse":      "Fuse",
	"harmony":   "Harmony",
	"heco":      "Heco",
	"kava":      "Kava",
	"metis":     "Metis",
	"moonbeam":  "Moonbeam",
	"moonriver": "Moonriver",
	"optimism":  "Optimism",
	"polygon":   "Polygon",
	"zkevm":     "Polygon zkEVM",
	"zksync":    "zkSync",
}

// GetVaults fetches vault metadata for a specific chain
func (c *BeefyClient) GetVaults(chain string) ([]BeefyVault, error) {
	url := fmt.Sprintf("%s/vaults/%s", c.baseURL, chain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "DeFiRates/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vaults: %w", err)
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

	var vaults []BeefyVault
	if err := json.Unmarshal(body, &vaults); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Set chain on each vault
	for i := range vaults {
		vaults[i].Chain = chain
	}

	return vaults, nil
}

// GetAPYData fetches APY breakdown data for all vaults
func (c *BeefyClient) GetAPYData() (map[string]BeefyAPYBreakdown, error) {
	url := fmt.Sprintf("%s/apy/breakdown", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "DeFiRates/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch APY data: %w", err)
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

	var apyData map[string]BeefyAPYBreakdown
	if err := json.Unmarshal(body, &apyData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return apyData, nil
}

// GetTVLData fetches TVL data for all vaults
func (c *BeefyClient) GetTVLData() (map[string]float64, error) {
	url := fmt.Sprintf("%s/tvl", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "DeFiRates/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch TVL data: %w", err)
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

	var tvlData map[string]float64
	if err := json.Unmarshal(body, &tvlData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return tvlData, nil
}

// GetAllVaultsWithMetrics fetches vaults from all supported chains with APY and TVL data
func (c *BeefyClient) GetAllVaultsWithMetrics() ([]BeefyVaultWithMetrics, error) {
	// Fetch APY and TVL data once for all vaults
	apyData, err := c.GetAPYData()
	if err != nil {
		fmt.Printf("Warning: failed to fetch Beefy APY data: %v\n", err)
		apyData = make(map[string]BeefyAPYBreakdown)
	}

	tvlData, err := c.GetTVLData()
	if err != nil {
		fmt.Printf("Warning: failed to fetch Beefy TVL data: %v\n", err)
		tvlData = make(map[string]float64)
	}

	var allVaults []BeefyVaultWithMetrics

	// Fetch vaults from each supported chain
	for _, chain := range BeefySupportedChains {
		vaults, err := c.GetVaults(chain)
		if err != nil {
			fmt.Printf("Warning: failed to fetch Beefy vaults for chain %s: %v\n", chain, err)
			continue
		}

		for _, vault := range vaults {
			// Only include active vaults
			if vault.Status != "active" {
				continue
			}

			apy := 0.0
			if apyBreakdown, ok := apyData[vault.ID]; ok {
				apy = apyBreakdown.TotalApy
			}

			tvl := 0.0
			if tvlValue, ok := tvlData[vault.ID]; ok {
				tvl = tvlValue
			}

			allVaults = append(allVaults, BeefyVaultWithMetrics{
				Vault: vault,
				APY:   apy * 100, // Convert from decimal to percentage
				TVL:   tvl,
				Chain: GetBeefyChainName(chain),
			})
		}
	}

	fmt.Printf("DEBUG: Fetched %d active Beefy vaults across all chains\n", len(allVaults))

	return allVaults, nil
}

// GetBeefyChainName returns the human-readable chain name
func GetBeefyChainName(chainID string) string {
	if name, ok := BeefyChainNameMapping[chainID]; ok {
		return name
	}
	// Capitalize first letter if no mapping exists
	if len(chainID) > 0 {
		return strings.ToUpper(chainID[:1]) + chainID[1:]
	}
	return chainID
}
