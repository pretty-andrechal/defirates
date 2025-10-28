package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pretty-andrechal/defirates/internal/api"
)

func main() {
	fmt.Println("=== Testing Each Chain Individually ===\n")

	// Test each chain that the app tries
	chainIDs := []int{1, 10, 56, 146, 999, 5000, 8453, 9745, 42161, 80094}

	client := &http.Client{Timeout: 30 * time.Second}

	successCount := 0
	totalMarkets := 0

	for _, chainID := range chainIDs {
		chainName := api.GetChainName(chainID)
		url := fmt.Sprintf("https://api-v2.pendle.finance/api/core/v1/%d/markets/active", chainID)

		fmt.Printf("Chain %d (%s)\n", chainID, chainName)
		fmt.Printf("  URL: %s\n", url)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("  ❌ Failed to create request: %v\n\n", err)
			continue
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "DeFiRates/1.0")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("  ❌ Request failed: %v\n\n", err)
			continue
		}

		fmt.Printf("  Status: %d\n", resp.StatusCode)

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Printf("  ❌ Error: %s\n\n", string(body))
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("  ❌ Failed to read body: %v\n\n", err)
			continue
		}

		var result api.MarketsResponse
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Printf("  ❌ Failed to parse JSON: %v\n", err)
			fmt.Printf("  Body preview: %s\n\n", string(body[:min(200, len(body))]))
			continue
		}

		fmt.Printf("  ✅ Success! %d markets\n", len(result.Results))
		if len(result.Results) > 0 {
			fmt.Printf("     Sample: %s (APY: %.2f%%)\n", result.Results[0].Symbol, result.Results[0].ImpliedAPY*100)
		}
		fmt.Println()

		successCount++
		totalMarkets += len(result.Results)
	}

	fmt.Println("=== Summary ===")
	fmt.Printf("Chains tested: %d\n", len(chainIDs))
	fmt.Printf("Successful: %d\n", successCount)
	fmt.Printf("Failed: %d\n", len(chainIDs)-successCount)
	fmt.Printf("Total markets: %d\n", totalMarkets)

	if successCount == 0 {
		fmt.Println("\n❌ No chains returned data - there's an issue with the requests")
	} else if totalMarkets == 0 {
		fmt.Println("\n⚠️ Chains responded but no markets found")
	} else {
		fmt.Println("\n✅ Everything works! Markets are being fetched successfully")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
