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
	db     *database.DB
	pendle *PendleClient
}

// NewFetcher creates a new data fetcher
func NewFetcher(db *database.DB) *Fetcher {
	return &Fetcher{
		db:     db,
		pendle: NewPendleClient(),
	}
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
		return nil
	}

	log.Printf("Found %d active Pendle markets", len(markets))

	// Store each market as a yield rate
	successCount := 0
	for _, market := range markets {
		yieldRate := f.convertMarketToYieldRate(market, protocol.ID)

		if err := f.db.UpsertYieldRate(&yieldRate); err != nil {
			log.Printf("Failed to store yield rate for %s: %v", market.Symbol, err)
			continue
		}
		successCount++
	}

	log.Printf("Successfully stored %d yield rates", successCount)
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

	// Determine the asset (use the underlying token symbol if available)
	asset := market.Symbol
	if len(market.Underlyings) > 0 {
		asset = market.Underlyings[0].Symbol
	} else if market.PT.Symbol != "" {
		// Extract base asset from PT symbol (e.g., "PT-eETH" -> "eETH")
		asset = market.PT.Symbol
	}

	// Get chain name
	chain := GetChainName(market.ChainID)

	// Convert implied APY from decimal to percentage
	apy := market.ImpliedAPY * 100

	// TVL is the liquidity in USD
	tvl := market.Liquidity.USD

	// Generate pool name and external URL
	poolName := fmt.Sprintf("%s-%d", market.Symbol, market.ChainID)
	externalURL := fmt.Sprintf("https://app.pendle.finance/trade/pools/%s/", market.Address)

	return models.YieldRate{
		ProtocolID:   protocolID,
		Asset:        asset,
		Chain:        chain,
		APY:          apy,
		TVL:          tvl,
		MaturityDate: maturityDate,
		PoolName:     poolName,
		ExternalURL:  externalURL,
	}
}

// StartPeriodicFetch starts a background goroutine that fetches data periodically
func (f *Fetcher) StartPeriodicFetch(interval time.Duration) {
	// Fetch immediately on startup
	if err := f.FetchAndStorePendleData(); err != nil {
		log.Printf("Error fetching Pendle data on startup: %v", err)
	}

	// Then fetch periodically
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if err := f.FetchAndStorePendleData(); err != nil {
				log.Printf("Error fetching Pendle data: %v", err)
			}
		}
	}()
}
