package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	url := "https://api-v2.pendle.finance/api/core/v1/1/markets/active"

	fmt.Println("=== API Access Diagnostic Tool ===\n")

	// Test 1: curl command
	fmt.Println("Test 1: Using curl (like you did)")
	fmt.Println("----------------------------------")
	testWithCurl(url)
	fmt.Println()

	// Test 2: Go default client
	fmt.Println("Test 2: Go HTTP Client (default)")
	fmt.Println("----------------------------------")
	testGoDefault(url)
	fmt.Println()

	// Test 3: Go with User-Agent
	fmt.Println("Test 3: Go HTTP Client (with User-Agent)")
	fmt.Println("------------------------------------------")
	testGoWithUserAgent(url)
	fmt.Println()

	// Test 4: Check environment
	fmt.Println("Test 4: Environment Check")
	fmt.Println("--------------------------")
	checkEnvironment()
	fmt.Println()

	// Test 5: Try with different timeouts
	fmt.Println("Test 5: Different Timeout Values")
	fmt.Println("----------------------------------")
	testTimeouts(url)
	fmt.Println()

	fmt.Println("=== Diagnostic Complete ===")
	fmt.Println("\nIf curl works but Go fails, the issue is likely:")
	fmt.Println("1. Missing User-Agent header")
	fmt.Println("2. HTTP client configuration")
	fmt.Println("3. TLS/SSL settings")
	fmt.Println("4. Proxy configuration")
}

func testWithCurl(url string) {
	cmd := exec.Command("curl", "-s", "-w", "\\nHTTP_CODE:%{http_code}", url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ curl command failed: %v\n", err)
		return
	}

	result := string(output)
	lines := strings.Split(result, "\n")

	// Look for HTTP_CODE in output
	httpCode := "unknown"
	for _, line := range lines {
		if strings.HasPrefix(line, "HTTP_CODE:") {
			httpCode = strings.TrimPrefix(line, "HTTP_CODE:")
			break
		}
	}

	fmt.Printf("HTTP Status from curl: %s\n", httpCode)

	if httpCode == "200" {
		fmt.Println("✅ curl can access the API successfully")

		// Try to parse response
		jsonData := strings.Split(result, "HTTP_CODE:")[0]
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(jsonData), &data); err == nil {
			if results, ok := data["results"].([]interface{}); ok {
				fmt.Printf("✅ Found %d markets in response\n", len(results))
			}
		}
	} else {
		fmt.Printf("❌ curl also gets non-200 status: %s\n", httpCode)
		fmt.Printf("Response: %s\n", string(output[:min(200, len(output))]))
	}
}

func testGoDefault(url string) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status Code: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		body, _ := io.ReadAll(resp.Body)
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err == nil {
			if results, ok := data["results"].([]interface{}); ok {
				fmt.Printf("✅ Go client works! Found %d markets\n", len(results))
			}
		}
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Got status %d\n", resp.StatusCode)
		fmt.Printf("Response: %s\n", string(body))
	}
}

func testGoWithUserAgent(url string) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}

	// Add User-Agent like browsers do
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	fmt.Printf("Request headers:\n")
	for k, v := range req.Header {
		fmt.Printf("  %s: %s\n", k, v[0])
	}
	fmt.Println()

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status Code: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		body, _ := io.ReadAll(resp.Body)
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err == nil {
			if results, ok := data["results"].([]interface{}); ok {
				fmt.Printf("✅ Go client with User-Agent works! Found %d markets\n", len(results))
			}
		}
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Still got status %d\n", resp.StatusCode)
		fmt.Printf("Response: %s\n", string(body))
	}
}

func checkEnvironment() {
	envVars := []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "http_proxy", "https_proxy", "no_proxy"}

	fmt.Println("Environment variables:")
	for _, env := range envVars {
		val := os.Getenv(env)
		if val != "" {
			fmt.Printf("  %s = %s\n", env, val)
		}
	}

	// Check Go version
	cmd := exec.Command("go", "version")
	output, _ := cmd.Output()
	fmt.Printf("\nGo version: %s", string(output))
}

func testTimeouts(url string) {
	timeouts := []time.Duration{5 * time.Second, 15 * time.Second, 30 * time.Second}

	for _, timeout := range timeouts {
		client := &http.Client{Timeout: timeout}

		fmt.Printf("Timeout: %v - ", timeout)

		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			fmt.Printf("✅ Success!\n")
		} else {
			fmt.Printf("❌ Status: %d\n", resp.StatusCode)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
