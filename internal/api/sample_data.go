package api

import (
	"log"
	"time"

	"github.com/pretty-andrechal/defirates/internal/database"
	"github.com/pretty-andrechal/defirates/internal/models"
)

// LoadSampleData loads sample DeFi yield data into the database for demonstration
func LoadSampleData(db *database.DB) error {
	log.Println("Loading sample data...")

	// Create Pendle protocol
	protocol := &models.Protocol{
		Name:        "Pendle",
		URL:         "https://www.pendle.finance",
		Description: "Pendle is a protocol that enables the tokenization and trading of future yield",
	}

	if err := db.CreateOrUpdateProtocol(protocol); err != nil {
		return err
	}

	// Sample yield rates
	sampleRates := []models.YieldRate{
		{
			ProtocolID:   protocol.ID,
			Asset:        "eETH",
			Chain:        "Ethereum",
			APY:          12.45,
			TVL:          15_234_567.89,
			MaturityDate: timePtr(time.Date(2025, 12, 26, 0, 0, 0, 0, time.UTC)),
			PoolName:     "PT-eETH-26DEC2025",
			ExternalURL:  "https://app.pendle.finance/trade/pools/0xf32e58f2f85714a65d2dcbb753e00ce58434f000/",
		},
		{
			ProtocolID:   protocol.ID,
			Asset:        "ezETH",
			Chain:        "Ethereum",
			APY:          15.23,
			TVL:          8_945_123.45,
			MaturityDate: timePtr(time.Date(2025, 12, 26, 0, 0, 0, 0, time.UTC)),
			PoolName:     "PT-ezETH-26DEC2025",
			ExternalURL:  "https://app.pendle.finance/trade/pools/0xd1d7d99764f8a52aff007b7831cc02748b2013b5/",
		},
		{
			ProtocolID:   protocol.ID,
			Asset:        "rsETH",
			Chain:        "Ethereum",
			APY:          13.87,
			TVL:          12_678_901.23,
			MaturityDate: timePtr(time.Date(2025, 12, 26, 0, 0, 0, 0, time.UTC)),
			PoolName:     "PT-rsETH-26DEC2025",
			ExternalURL:  "https://app.pendle.finance/trade/pools/0x4f43c77872db6ba177c270986cd30c3381af37ee/",
		},
		{
			ProtocolID:   protocol.ID,
			Asset:        "sUSDe",
			Chain:        "Ethereum",
			APY:          25.67,
			TVL:          45_123_456.78,
			MaturityDate: timePtr(time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)),
			PoolName:     "PT-sUSDe-29JAN2026",
			ExternalURL:  "https://app.pendle.finance/trade/pools/0x4a8e8befd2cf1480032a6f8a5c45d8c3ae1e8829/",
		},
		{
			ProtocolID:   protocol.ID,
			Asset:        "LBTC",
			Chain:        "Ethereum",
			APY:          8.92,
			TVL:          23_456_789.01,
			MaturityDate: timePtr(time.Date(2025, 12, 26, 0, 0, 0, 0, time.UTC)),
			PoolName:     "PT-LBTC-26DEC2025",
			ExternalURL:  "https://app.pendle.finance/trade/pools/0x8a47b431a7d947c6a3ed6e42d501803615a97eaa/",
		},
		{
			ProtocolID:   protocol.ID,
			Asset:        "agETH",
			Chain:        "Arbitrum",
			APY:          16.34,
			TVL:          5_678_901.23,
			MaturityDate: timePtr(time.Date(2025, 12, 26, 0, 0, 0, 0, time.UTC)),
			PoolName:     "PT-agETH-26DEC2025",
			ExternalURL:  "https://app.pendle.finance/trade/pools/0x4a8e8befd2cf1480032a6f8a5c45d8c3ae1e8829/",
		},
		{
			ProtocolID:   protocol.ID,
			Asset:        "rsETH",
			Chain:        "Arbitrum",
			APY:          14.56,
			TVL:          7_890_123.45,
			MaturityDate: timePtr(time.Date(2025, 12, 26, 0, 0, 0, 0, time.UTC)),
			PoolName:     "PT-rsETH-26DEC2025-ARB",
			ExternalURL:  "https://app.pendle.finance/trade/pools/0xed99fc8bdb8e9e7b8240f62f69609a125a0fbf14/",
		},
		{
			ProtocolID:   protocol.ID,
			Asset:        "USDe",
			Chain:        "Arbitrum",
			APY:          18.23,
			TVL:          34_567_890.12,
			MaturityDate: timePtr(time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)),
			PoolName:     "PT-USDe-29JAN2026",
			ExternalURL:  "https://app.pendle.finance/trade/pools/0xbfef9183b47b3dd89a025f7dbfb44c58f4e0b68f/",
		},
		{
			ProtocolID:   protocol.ID,
			Asset:        "wstETH",
			Chain:        "Optimism",
			APY:          11.78,
			TVL:          9_876_543.21,
			MaturityDate: timePtr(time.Date(2025, 12, 26, 0, 0, 0, 0, time.UTC)),
			PoolName:     "PT-wstETH-26DEC2025",
			ExternalURL:  "https://app.pendle.finance/trade/pools/0x1c27ad8a19ba026adabd615f6bc77158130cfbe4/",
		},
		{
			ProtocolID:   protocol.ID,
			Asset:        "sUSDe",
			Chain:        "Arbitrum",
			APY:          26.45,
			TVL:          28_901_234.56,
			MaturityDate: timePtr(time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)),
			PoolName:     "PT-sUSDe-29JAN2026-ARB",
			ExternalURL:  "https://app.pendle.finance/trade/pools/0xa0192f6567f8f5dc38c53323235fd08b318d2dca/",
		},
		{
			ProtocolID:   protocol.ID,
			Asset:        "cbBTC",
			Chain:        "Base",
			APY:          9.87,
			TVL:          18_234_567.89,
			MaturityDate: timePtr(time.Date(2025, 12, 26, 0, 0, 0, 0, time.UTC)),
			PoolName:     "PT-cbBTC-26DEC2025",
			ExternalURL:  "https://app.pendle.finance/trade/pools/0x94caeb3b9a1b7c61ef364f6c52260cf89b3bc667/",
		},
		{
			ProtocolID:   protocol.ID,
			Asset:        "weETH",
			Chain:        "Base",
			APY:          13.21,
			TVL:          6_543_210.98,
			MaturityDate: timePtr(time.Date(2025, 12, 26, 0, 0, 0, 0, time.UTC)),
			PoolName:     "PT-weETH-26DEC2025",
			ExternalURL:  "https://app.pendle.finance/trade/pools/0x8e5ca4d5f8f3e5e5b2c2e5f5d5c5b5a5e5d5c5b5/",
		},
		{
			ProtocolID:   protocol.ID,
			Asset:        "mETH",
			Chain:        "Mantle",
			APY:          14.89,
			TVL:          4_321_098.76,
			MaturityDate: timePtr(time.Date(2025, 12, 26, 0, 0, 0, 0, time.UTC)),
			PoolName:     "PT-mETH-26DEC2025",
			ExternalURL:  "https://app.pendle.finance/trade/pools/0xa1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2/",
		},
		{
			ProtocolID:   protocol.ID,
			Asset:        "sUSDe",
			Chain:        "Mantle",
			APY:          22.34,
			TVL:          12_345_678.90,
			MaturityDate: timePtr(time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)),
			PoolName:     "PT-sUSDe-29JAN2026-MANTLE",
			ExternalURL:  "https://app.pendle.finance/trade/pools/0xb2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3/",
		},
	}

	for _, rate := range sampleRates {
		if err := db.UpsertYieldRate(&rate); err != nil {
			log.Printf("Failed to insert sample rate: %v", err)
			continue
		}
	}

	log.Printf("Successfully loaded %d sample yield rates across multiple chains", len(sampleRates))
	return nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}
