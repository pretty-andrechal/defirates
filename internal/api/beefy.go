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
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	// Note: Don't set Accept-Encoding - Go's http client handles gzip automatically
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Origin", "https://app.beefy.finance")
	req.Header.Set("Referer", "https://app.beefy.finance/")

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
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	// Note: Don't set Accept-Encoding - Go's http client handles gzip automatically
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Origin", "https://app.beefy.finance")
	req.Header.Set("Referer", "https://app.beefy.finance/")

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
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	// Note: Don't set Accept-Encoding - Go's http client handles gzip automatically
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Origin", "https://app.beefy.finance")
	req.Header.Set("Referer", "https://app.beefy.finance/")

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
	fmt.Println("DEBUG: Fetching Beefy APY data from API...")
	apyData, err := c.GetAPYData()
	if err != nil {
		fmt.Printf("WARNING: failed to fetch Beefy APY data: %v\n", err)
		apyData = make(map[string]BeefyAPYBreakdown)
	} else {
		fmt.Printf("DEBUG: Successfully fetched APY data for %d vaults\n", len(apyData))
		// Log first 3 entries as sample
		count := 0
		for id, breakdown := range apyData {
			if count < 3 {
				fmt.Printf("DEBUG: Sample APY - %s: %.4f (%.2f%%)\n", id, breakdown.TotalApy, breakdown.TotalApy*100)
				count++
			} else {
				break
			}
		}
	}

	fmt.Println("DEBUG: Fetching Beefy TVL data from API...")
	tvlData, err := c.GetTVLData()
	if err != nil {
		fmt.Printf("WARNING: failed to fetch Beefy TVL data: %v\n", err)
		tvlData = make(map[string]float64)
	} else {
		fmt.Printf("DEBUG: Successfully fetched TVL data for %d vaults\n", len(tvlData))
		// Log first 3 entries as sample
		count := 0
		for id, tvl := range tvlData {
			if count < 3 {
				fmt.Printf("DEBUG: Sample TVL - %s: $%.2f\n", id, tvl)
				count++
			} else {
				break
			}
		}
	}

	var allVaults []BeefyVaultWithMetrics
	totalVaultsFound := 0
	vaultsWithAPY := 0
	vaultsWithTVL := 0

	// Fetch vaults from each supported chain
	for _, chain := range BeefySupportedChains {
		vaults, err := c.GetVaults(chain)
		if err != nil {
			fmt.Printf("WARNING: failed to fetch Beefy vaults for chain %s: %v\n", chain, err)
			continue
		}

		fmt.Printf("DEBUG: Chain %s returned %d vaults\n", chain, len(vaults))
		totalVaultsFound += len(vaults)

		activeCount := 0
		for _, vault := range vaults {
			// Only include active vaults
			if vault.Status != "active" {
				continue
			}
			activeCount++

			apy := 0.0
			apyFound := false
			if apyBreakdown, ok := apyData[vault.ID]; ok {
				apy = apyBreakdown.TotalApy
				apyFound = true
				vaultsWithAPY++
			}

			tvl := 0.0
			tvlFound := false
			if tvlValue, ok := tvlData[vault.ID]; ok {
				tvl = tvlValue
				tvlFound = true
				vaultsWithTVL++
			}

			// Log first few vaults with detailed info
			if len(allVaults) < 3 {
				fmt.Printf("DEBUG: Vault %s - APY: %.4f (found: %v), TVL: %.2f (found: %v)\n",
					vault.ID, apy, apyFound, tvl, tvlFound)
			}

			allVaults = append(allVaults, BeefyVaultWithMetrics{
				Vault: vault,
				APY:   apy * 100, // Convert from decimal to percentage
				TVL:   tvl,
				Chain: GetBeefyChainName(chain),
			})
		}

		if len(vaults) > 0 {
			fmt.Printf("DEBUG: Chain %s: %d active vaults out of %d total\n", chain, activeCount, len(vaults))
		}
	}

	fmt.Printf("DEBUG: Summary - Total vaults found: %d, Active vaults: %d\n", totalVaultsFound, len(allVaults))
	fmt.Printf("DEBUG: Vaults with APY data: %d (%.1f%%)\n", vaultsWithAPY, float64(vaultsWithAPY)/float64(len(allVaults))*100)
	fmt.Printf("DEBUG: Vaults with TVL data: %d (%.1f%%)\n", vaultsWithTVL, float64(vaultsWithTVL)/float64(len(allVaults))*100)

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
