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
			Categories:   "PT, LRT, Liquidity",
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
			Categories:   "PT, LRT",
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
			Categories:   "PT, LRT, Liquidity",
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
			Categories:   "PT, Stablecoin",
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
			Categories:   "PT, BTC",
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
			Categories:   "PT, LST",
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
			Categories:   "PT, LRT",
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
			Categories:   "PT, Stablecoin, Liquidity",
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
			Categories:   "PT, LST, Liquidity",
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
			Categories:   "PT, Stablecoin",
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
			Categories:   "PT, BTC, Liquidity",
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
			Categories:   "PT, LST",
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
			Categories:   "PT, LST",
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
			Categories:   "PT, Stablecoin, Liquidity",
			ExternalURL:  "https://app.pendle.finance/trade/pools/0xb2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3/",
		},
	}

	for _, rate := range sampleRates {
		if err := db.UpsertYieldRate(&rate); err != nil {
			log.Printf("Failed to insert sample rate: %v", err)
			continue
		}
	}

	log.Printf("Successfully loaded %d Pendle sample yield rates across multiple chains", len(sampleRates))

	// Create Beefy protocol
	beefyProtocol := &models.Protocol{
		Name:        "Beefy",
		URL:         "https://beefy.finance",
		Description: "Beefy is a Decentralized, Multichain Yield Optimizer",
	}

	if err := db.CreateOrUpdateProtocol(beefyProtocol); err != nil {
		return err
	}

	// Sample Beefy vaults
	beefySampleRates := []models.YieldRate{
		{
			ProtocolID:   beefyProtocol.ID,
			Asset:        "WETH-USDC LP",
			Chain:        "Ethereum",
			APY:          8.45,
			TVL:          12_345_678.90,
			MaturityDate: nil, // Beefy vaults don't have maturity
			PoolName:     "uniswap-v3-eth-usdc",
			Categories:   "Beefy, WETH, USDC",
			ExternalURL:  "https://app.beefy.finance/vault/uniswap-v3-eth-usdc",
		},
		{
			ProtocolID:   beefyProtocol.ID,
			Asset:        "wstETH",
			Chain:        "Ethereum",
			APY:          4.23,
			TVL:          45_678_901.23,
			MaturityDate: nil,
			PoolName:     "lido-wsteth",
			Categories:   "Beefy, wstETH",
			ExternalURL:  "https://app.beefy.finance/vault/lido-wsteth",
		},
		{
			ProtocolID:   beefyProtocol.ID,
			Asset:        "USDC-USDT LP",
			Chain:        "Arbitrum",
			APY:          6.78,
			TVL:          23_456_789.01,
			MaturityDate: nil,
			PoolName:     "curve-arb-2pool",
			Categories:   "Beefy, USDC, USDT",
			ExternalURL:  "https://app.beefy.finance/vault/curve-arb-2pool",
		},
		{
			ProtocolID:   beefyProtocol.ID,
			Asset:        "ETH-USDC LP",
			Chain:        "Arbitrum",
			APY:          12.34,
			TVL:          18_901_234.56,
			MaturityDate: nil,
			PoolName:     "sushi-arb-eth-usdc",
			Categories:   "Beefy, ETH, USDC",
			ExternalURL:  "https://app.beefy.finance/vault/sushi-arb-eth-usdc",
		},
		{
			ProtocolID:   beefyProtocol.ID,
			Asset:        "BNB-BUSD LP",
			Chain:        "BSC",
			APY:          15.67,
			TVL:          34_567_890.12,
			MaturityDate: nil,
			PoolName:     "pancake-bnb-busd",
			Categories:   "Beefy, BNB, BUSD",
			ExternalURL:  "https://app.beefy.finance/vault/pancake-bnb-busd",
		},
		{
			ProtocolID:   beefyProtocol.ID,
			Asset:        "CAKE",
			Chain:        "BSC",
			APY:          22.45,
			TVL:          9_876_543.21,
			MaturityDate: nil,
			PoolName:     "pancake-cake",
			Categories:   "Beefy, CAKE",
			ExternalURL:  "https://app.beefy.finance/vault/pancake-cake",
		},
		{
			ProtocolID:   beefyProtocol.ID,
			Asset:        "MATIC-USDC LP",
			Chain:        "Polygon",
			APY:          9.87,
			TVL:          15_678_901.23,
			MaturityDate: nil,
			PoolName:     "quickswap-matic-usdc",
			Categories:   "Beefy, MATIC, USDC",
			ExternalURL:  "https://app.beefy.finance/vault/quickswap-matic-usdc",
		},
		{
			ProtocolID:   beefyProtocol.ID,
			Asset:        "AVAX-USDC LP",
			Chain:        "Avalanche",
			APY:          11.23,
			TVL:          21_234_567.89,
			MaturityDate: nil,
			PoolName:     "trader-joe-avax-usdc",
			Categories:   "Beefy, AVAX, USDC",
			ExternalURL:  "https://app.beefy.finance/vault/trader-joe-avax-usdc",
		},
		{
			ProtocolID:   beefyProtocol.ID,
			Asset:        "OP-USDC LP",
			Chain:        "Optimism",
			APY:          10.56,
			TVL:          8_765_432.10,
			MaturityDate: nil,
			PoolName:     "velodrome-op-usdc",
			Categories:   "Beefy, OP, USDC",
			ExternalURL:  "https://app.beefy.finance/vault/velodrome-op-usdc",
		},
		{
			ProtocolID:   beefyProtocol.ID,
			Asset:        "ETH-USDC LP",
			Chain:        "Base",
			APY:          13.89,
			TVL:          17_890_123.45,
			MaturityDate: nil,
			PoolName:     "aerodrome-base-eth-usdc",
			Categories:   "Beefy, ETH, USDC",
			ExternalURL:  "https://app.beefy.finance/vault/aerodrome-base-eth-usdc",
		},
	}

	for _, rate := range beefySampleRates {
		if err := db.UpsertYieldRate(&rate); err != nil {
			log.Printf("Failed to insert Beefy sample rate: %v", err)
			continue
		}
	}

	log.Printf("Successfully loaded %d Beefy sample yield rates across multiple chains", len(beefySampleRates))
	log.Printf("Total sample data loaded: %d yield rates from 2 protocols", len(sampleRates)+len(beefySampleRates))
	return nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}
