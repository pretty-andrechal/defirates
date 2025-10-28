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
	fmt.Println("=== Pendle API Implementation Test Suite ===\n")

	// Test 1: Basic HTTP connectivity
	fmt.Println("Test 1: Basic HTTP Connectivity")
	fmt.Println("--------------------------------")
	testBasicConnectivity()
	fmt.Println()

	// Test 2: Headers and response
	fmt.Println("Test 2: Request Headers & Response")
	fmt.Println("-----------------------------------")
	testHeadersAndResponse()
	fmt.Println()

	// Test 3: JSON parsing
	fmt.Println("Test 3: JSON Response Parsing")
	fmt.Println("------------------------------")
	testJSONParsing()
	fmt.Println()

	// Test 4: Market data structure
	fmt.Println("Test 4: Market Data Structure")
	fmt.Println("------------------------------")
	testMarketStructure()
	fmt.Println()

	// Test 5: PendleClient implementation
	fmt.Println("Test 5: PendleClient.GetMarketsForChain()")
	fmt.Println("------------------------------------------")
	testPendleClient()
	fmt.Println()

	// Test 6: Full GetMarkets() across all chains
	fmt.Println("Test 6: Full GetMarkets() Integration")
	fmt.Println("--------------------------------------")
	testFullIntegration()
	fmt.Println()

	fmt.Println("=== Test Suite Complete ===")
}

func testBasicConnectivity() {
	url := "https://api-v2.pendle.finance/api/core/v1/1/markets/active"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	if resp.StatusCode == 200 {
		fmt.Printf("✅ PASSED: Successfully connected to API\n")
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ FAILED: Expected 200, got %d\n", resp.StatusCode)
		fmt.Printf("Response: %s\n", string(body))
	}
}

func testHeadersAndResponse() {
	url := "https://api-v2.pendle.finance/api/core/v1/1/markets/active"

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("❌ FAILED to create request: %v\n", err)
		return
	}

	// Set headers like our implementation does
	req.Header.Set("Accept", "application/json")

	fmt.Printf("Request URL: %s\n", req.URL.String())
	fmt.Printf("Request Headers:\n")
	for k, v := range req.Header {
		fmt.Printf("  %s: %v\n", k, v)
	}
	fmt.Println()

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Response Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Headers:\n")
	for k, v := range resp.Header {
		fmt.Printf("  %s: %v\n", k, v)
	}

	if resp.StatusCode == 200 {
		fmt.Printf("\n✅ PASSED: Request with headers successful\n")
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("\n❌ FAILED: Got status %d\n", resp.StatusCode)
		fmt.Printf("Response body: %s\n", string(body))
	}
}

func testJSONParsing() {
	url := "https://api-v2.pendle.finance/api/core/v1/1/markets/active"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("❌ FAILED: Got status %d\n", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ FAILED to read body: %v\n", err)
		return
	}

	fmt.Printf("Response body length: %d bytes\n", len(body))

	// Try to parse as generic JSON first
	var rawJSON map[string]interface{}
	if err := json.Unmarshal(body, &rawJSON); err != nil {
		fmt.Printf("❌ FAILED to parse JSON: %v\n", err)
		fmt.Printf("Body preview: %s\n", string(body[:min(200, len(body))]))
		return
	}

	fmt.Printf("JSON structure keys: %v\n", getKeys(rawJSON))

	if results, ok := rawJSON["results"].([]interface{}); ok {
		fmt.Printf("Number of markets in results: %d\n", len(results))
		fmt.Printf("✅ PASSED: JSON parsed successfully\n")
	} else {
		fmt.Printf("❌ FAILED: 'results' field not found or wrong type\n")
	}
}

func testMarketStructure() {
	url := "https://api-v2.pendle.finance/api/core/v1/1/markets/active"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("❌ FAILED: Got status %d\n", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ FAILED to read body: %v\n", err)
		return
	}

	// Use our actual MarketsResponse structure
	var marketsResp api.MarketsResponse
	if err := json.Unmarshal(body, &marketsResp); err != nil {
		fmt.Printf("❌ FAILED to unmarshal into MarketsResponse: %v\n", err)
		return
	}

	fmt.Printf("Total markets: %d\n", marketsResp.Total)
	fmt.Printf("Results count: %d\n", len(marketsResp.Results))
	fmt.Printf("Limit: %d\n", marketsResp.Limit)
	fmt.Printf("Skip: %d\n", marketsResp.Skip)

	if len(marketsResp.Results) > 0 {
		market := marketsResp.Results[0]
		fmt.Printf("\nFirst market sample:\n")
		fmt.Printf("  Symbol: %s\n", market.Symbol)
		fmt.Printf("  Address: %s\n", market.Address)
		fmt.Printf("  ChainID: %d\n", market.ChainID)
		fmt.Printf("  Expiry: %s\n", market.Expiry)
		fmt.Printf("  ImpliedAPY: %.4f\n", market.ImpliedAPY)
		fmt.Printf("  Liquidity USD: $%.2f\n", market.Liquidity.USD)
		fmt.Printf("\n✅ PASSED: Market structure matches\n")
	} else {
		fmt.Printf("❌ FAILED: No markets in results\n")
	}
}

func testPendleClient() {
	client := api.NewPendleClient()

	chainID := 1 // Ethereum
	fmt.Printf("Testing GetMarketsForChain(%d)...\n", chainID)

	markets, err := client.GetMarketsForChain(chainID)
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		return
	}

	fmt.Printf("✅ PASSED: Fetched %d markets for chain %d\n", len(markets), chainID)

	if len(markets) > 0 {
		fmt.Printf("\nSample market:\n")
		fmt.Printf("  Symbol: %s\n", markets[0].Symbol)
		fmt.Printf("  Chain: %d\n", markets[0].ChainID)
		fmt.Printf("  APY: %.2f%%\n", markets[0].ImpliedAPY*100)
	}
}

func testFullIntegration() {
	client := api.NewPendleClient()

	fmt.Println("Testing GetMarkets() across all chains...")

	markets, err := client.GetMarkets()
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		return
	}

	fmt.Printf("✅ PASSED: Fetched total of %d markets\n", len(markets))

	// Group by chain
	chainCounts := make(map[int]int)
	for _, market := range markets {
		chainCounts[market.ChainID]++
	}

	fmt.Println("\nMarkets by chain:")
	for chainID, count := range chainCounts {
		chainName := api.GetChainName(chainID)
		fmt.Printf("  %s (ID: %d): %d markets\n", chainName, chainID, count)
	}
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
