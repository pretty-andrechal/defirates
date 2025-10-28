package api

import (
	"fmt"
	"log"
	"time"

	"github.com/pretty-andrechal/defirates/internal/database"
	"github.com/pretty-andrechal/defirates/internal/models"
)

// Fetcher handles fetching and storing yield data
type Fetcher struct {
	db              *database.DB
	pendle          *PendleClient
	beefy           *BeefyClient
	onDataUpdate    func() // Callback function triggered when data is updated
	debugLogging    bool
}

// NewFetcher creates a new data fetcher
func NewFetcher(db *database.DB) *Fetcher {
	return &Fetcher{
		db:           db,
		pendle:       NewPendleClient(),
		beefy:        NewBeefyClient(),
		onDataUpdate: nil,
		debugLogging: false,
	}
}

// EnableDebugLogging enables HTTP debug logging for API calls
func (f *Fetcher) EnableDebugLogging() {
	f.debugLogging = true
	f.pendle = NewPendleClientWithDebug(f.db)
	f.beefy = NewBeefyClientWithDebug(f.db)
}

// SetOnDataUpdateCallback sets a callback function to be called when data is updated
func (f *Fetcher) SetOnDataUpdateCallback(callback func()) {
	f.onDataUpdate = callback
}

// FetchAndStorePendleData fetches data from Pendle and stores it in the database
func (f *Fetcher) FetchAndStorePendleData() error {
	log.Println("Fetching Pendle markets...")

	// Ensure Pendle protocol exists in database
	protocol := &models.Protocol{
		Name:        "Pendle",
		URL:         "https://www.pendle.finance",
		Description: "Pendle is a protocol that enables the tokenization and trading of future yield",
	}

	if err := f.db.CreateOrUpdateProtocol(protocol); err != nil {
		return fmt.Errorf("failed to create/update protocol: %w", err)
	}

	// Fetch active markets
	markets, err := f.pendle.GetActiveMarkets()
	if err != nil {
		log.Printf("Warning: failed to fetch Pendle markets: %v", err)
		log.Println("The Pendle API may be rate-limited or unavailable.")
		log.Println("You can still use the application - it will show any existing data.")
		log.Println("To see sample data, run with the -load-sample flag.")

		// Check if we have any existing data in the database
		// If yes, trigger update callback so browser can refresh displayed values
		existingRates, checkErr := f.db.GetYieldRates(models.FilterParams{})
		if checkErr == nil && len(existingRates) > 0 {
			log.Printf("Database contains %d existing rates, broadcasting refresh event", len(existingRates))
			if f.onDataUpdate != nil {
				log.Println("Broadcasting data update event...")
				f.onDataUpdate()
			}
		}

		return nil
	}

	log.Printf("Found %d active Pendle markets", len(markets))

	// Store each market as a yield rate
	successCount := 0
	for _, market := range markets {
		yieldRate := f.convertMarketToYieldRate(market, protocol.ID)

		if err := f.db.UpsertYieldRate(&yieldRate); err != nil {
			log.Printf("Failed to store yield rate for %s: %v", market.Name, err)
			continue
		}
		successCount++
	}

	log.Printf("Successfully stored %d yield rates", successCount)

	// Trigger data update callback if set
	if f.onDataUpdate != nil {
		log.Println("Broadcasting data update event...")
		f.onDataUpdate()
	}

	return nil
}

// convertMarketToYieldRate converts a Pendle market to our internal YieldRate model
func (f *Fetcher) convertMarketToYieldRate(market Market, protocolID int64) models.YieldRate {
	// Parse expiry date
	var maturityDate *time.Time
	if expiry, err := time.Parse("2006-01-02T15:04:05.000Z", market.Expiry); err == nil {
		maturityDate = &expiry
	} else if expiry, err := time.Parse(time.RFC3339, market.Expiry); err == nil {
		maturityDate = &expiry
	}

	// Use market name as asset (e.g., "wstETH", "sUSDe")
	asset := market.Name

	// Get chain name
	chain := GetChainName(market.ChainID)

	// Convert implied APY from decimal to percentage
	apy := market.Details.ImpliedAPY * 100

	// TVL is the liquidity in USD
	tvl := market.Details.Liquidity

	// Generate pool name and external URL
	poolName := fmt.Sprintf("%s-%d", market.Name, market.ChainID)
	externalURL := fmt.Sprintf("https://app.pendle.finance/trade/pools/%s/", market.Address)

	// Join category IDs into comma-separated string
	var categories string
	if len(market.CategoryIDs) > 0 {
		categories = fmt.Sprintf("%s", market.CategoryIDs[0])
		for i := 1; i < len(market.CategoryIDs); i++ {
			categories += ", " + market.CategoryIDs[i]
		}
	}

	return models.YieldRate{
		ProtocolID:   protocolID,
		Asset:        asset,
		Chain:        chain,
		APY:          apy,
		TVL:          tvl,
		MaturityDate: maturityDate,
		PoolName:     poolName,
		Categories:   categories,
		ExternalURL:  externalURL,
	}
}

// FetchAndStoreBeefyData fetches data from Beefy and stores it in the database
func (f *Fetcher) FetchAndStoreBeefyData() error {
	log.Println("Fetching Beefy vaults...")

	// Ensure Beefy protocol exists in database
	protocol := &models.Protocol{
		Name:        "Beefy",
		URL:         "https://beefy.finance",
		Description: "Beefy is a Decentralized, Multichain Yield Optimizer",
	}

	if err := f.db.CreateOrUpdateProtocol(protocol); err != nil {
		return fmt.Errorf("failed to create/update protocol: %w", err)
	}

	// Fetch vaults with metrics
	vaults, err := f.beefy.GetAllVaultsWithMetrics()
	if err != nil {
		log.Printf("Warning: failed to fetch Beefy vaults: %v", err)
		log.Println("The Beefy API may be unavailable.")
		return nil
	}

	log.Printf("Found %d active Beefy vaults", len(vaults))

	// Store each vault as a yield rate
	successCount := 0
	for i, vault := range vaults {
		yieldRate := f.convertBeefyVaultToYieldRate(vault, protocol.ID)

		// Log first few conversions for debugging
		if i < 3 {
			log.Printf("DEBUG: Converting vault %s - Input APY: %.2f%%, Input TVL: $%.2f",
				vault.Vault.ID, vault.APY, vault.TVL)
			log.Printf("DEBUG: YieldRate for %s - APY: %.2f%%, TVL: $%.2f, Categories: %s",
				yieldRate.PoolName, yieldRate.APY, yieldRate.TVL, yieldRate.Categories)
		}

		if err := f.db.UpsertYieldRate(&yieldRate); err != nil {
			log.Printf("Failed to store yield rate for %s: %v", vault.Vault.Name, err)
			continue
		}
		successCount++
	}

	log.Printf("Successfully stored %d Beefy yield rates", successCount)
	return nil
}

// convertBeefyVaultToYieldRate converts a Beefy vault to our internal YieldRate model
func (f *Fetcher) convertBeefyVaultToYieldRate(vault BeefyVaultWithMetrics, protocolID int64) models.YieldRate {
	// Use vault name as asset
	asset := vault.Vault.Name

	// Get chain name
	chain := vault.Chain

	// APY is already in percentage
	apy := vault.APY

	// TVL from vault metrics
	tvl := vault.TVL

	// Generate pool name with platform info
	poolName := fmt.Sprintf("%s-%s", vault.Vault.PlatformId, vault.Vault.ID)

	// Generate external URL
	externalURL := fmt.Sprintf("https://app.beefy.finance/vault/%s", vault.Vault.ID)

	// Join assets as categories
	categories := ""
	if len(vault.Vault.Assets) > 0 {
		categories = fmt.Sprintf("Beefy, %s", vault.Vault.Assets[0])
		for i := 1; i < len(vault.Vault.Assets) && i < 3; i++ {
			categories += ", " + vault.Vault.Assets[i]
		}
	} else {
		categories = "Beefy"
	}

	return models.YieldRate{
		ProtocolID:   protocolID,
		Asset:        asset,
		Chain:        chain,
		APY:          apy,
		TVL:          tvl,
		MaturityDate: nil, // Beefy vaults don't have maturity dates
		PoolName:     poolName,
		Categories:   categories,
		ExternalURL:  externalURL,
	}
}

// FetchAllData fetches data from all supported protocols
func (f *Fetcher) FetchAllData() error {
	// Fetch Pendle data
	if err := f.FetchAndStorePendleData(); err != nil {
		log.Printf("Error fetching Pendle data: %v", err)
	}

	// Fetch Beefy data
	if err := f.FetchAndStoreBeefyData(); err != nil {
		log.Printf("Error fetching Beefy data: %v", err)
	}

	// Trigger data update callback if set
	if f.onDataUpdate != nil {
		log.Println("Broadcasting data update event...")
		f.onDataUpdate()
	}

	return nil
}

// StartPeriodicFetch starts a background goroutine that fetches data periodically
func (f *Fetcher) StartPeriodicFetch(interval time.Duration) {
	// Fetch immediately on startup
	if err := f.FetchAllData(); err != nil {
		log.Printf("Error fetching data on startup: %v", err)
	}

	// Then fetch periodically
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if err := f.FetchAllData(); err != nil {
				log.Printf("Error fetching data: %v", err)
			}
		}
	}()
}
