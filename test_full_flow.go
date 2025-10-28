package main

import (
	"fmt"
	"log"

	"github.com/pretty-andrechal/defirates/internal/api"
	"github.com/pretty-andrechal/defirates/internal/database"
	"github.com/pretty-andrechal/defirates/internal/models"
)

func main() {
	fmt.Println("=== Full Application Flow Test ===\n")

	// Test 1: Test PendleClient directly
	fmt.Println("Test 1: Direct PendleClient Test")
	fmt.Println("---------------------------------")
	client := api.NewPendleClient()

	// Test single chain first
	fmt.Println("Testing Ethereum (chain 1)...")
	markets, err := client.GetMarketsForChain(1)
	if err != nil {
		fmt.Printf("❌ GetMarketsForChain(1) failed: %v\n", err)
	} else {
		fmt.Printf("✅ GetMarketsForChain(1) success: %d markets\n", len(markets))
		if len(markets) > 0 {
			fmt.Printf("   First market: %s (APY: %.2f%%)\n", markets[0].Symbol, markets[0].ImpliedAPY*100)
		}
	}
	fmt.Println()

	// Test GetMarkets (all chains)
	fmt.Println("Test 2: GetMarkets() - All Chains")
	fmt.Println("----------------------------------")
	allMarkets, err := client.GetMarkets()
	if err != nil {
		fmt.Printf("❌ GetMarkets() failed: %v\n", err)
	} else {
		fmt.Printf("✅ GetMarkets() success: %d total markets\n", len(allMarkets))

		// Count by chain
		chainCounts := make(map[int]int)
		for _, m := range allMarkets {
			chainCounts[m.ChainID]++
		}

		fmt.Println("Markets by chain:")
		for chainID, count := range chainCounts {
			fmt.Printf("   Chain %d (%s): %d markets\n", chainID, api.GetChainName(chainID), count)
		}
	}
	fmt.Println()

	// Test 3: Test GetActiveMarkets (with expiry filter)
	fmt.Println("Test 3: GetActiveMarkets() - Expiry Filter")
	fmt.Println("--------------------------------------------")
	activeMarkets, err := client.GetActiveMarkets()
	if err != nil {
		fmt.Printf("❌ GetActiveMarkets() failed: %v\n", err)
	} else {
		fmt.Printf("✅ GetActiveMarkets() success: %d active markets\n", len(activeMarkets))
	}
	fmt.Println()

	// Test 4: Full integration with database
	fmt.Println("Test 4: Full Fetcher Integration")
	fmt.Println("----------------------------------")

	// Create temp database
	db, err := database.New("test_defirates.db")
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	fetcher := api.NewFetcher(db)

	fmt.Println("Running FetchAndStorePendleData()...")
	err = fetcher.FetchAndStorePendleData()
	if err != nil {
		fmt.Printf("❌ FetchAndStorePendleData() failed: %v\n", err)
	} else {
		fmt.Printf("✅ FetchAndStorePendleData() completed\n")
	}

	// Query the database
	fmt.Println("\nQuerying database...")
	rates, err := db.GetYieldRates(models.FilterParams{
		SortBy:    "apy",
		SortOrder: "desc",
	})
	if err != nil {
		fmt.Printf("❌ Database query failed: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d yield rates in database\n", len(rates))
		if len(rates) > 0 {
			fmt.Printf("   Highest APY: %s on %s - %.2f%%\n",
				rates[0].Asset, rates[0].Chain, rates[0].APY)
		}
	}

	fmt.Println("\n=== Test Complete ===")
}
