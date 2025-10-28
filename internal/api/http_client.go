package api

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

// HTTPClientConfig holds configuration for making resilient HTTP requests
type HTTPClientConfig struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	ProxyURL      string // Optional proxy URL
	RequestDelay  time.Duration // Delay between requests to avoid rate limiting
}

// DefaultHTTPConfig returns sensible defaults
func DefaultHTTPConfig() HTTPClientConfig {
	return HTTPClientConfig{
		MaxRetries:   3,
		InitialDelay: 2 * time.Second,
		MaxDelay:     30 * time.Second,
		RequestDelay: 500 * time.Millisecond, // Small delay between requests
	}
}

// ResilientHTTPClient wraps http.Client with retry logic
type ResilientHTTPClient struct {
	client *http.Client
	config HTTPClientConfig
	lastRequest time.Time
}

// NewResilientHTTPClient creates a new HTTP client with retry and delay logic
func NewResilientHTTPClient(config HTTPClientConfig) *ResilientHTTPClient {
	transport := &http.Transport{}

	// Configure proxy if provided
	if config.ProxyURL != "" {
		proxyURL, err := url.Parse(config.ProxyURL)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
			fmt.Printf("INFO: Using proxy: %s\n", config.ProxyURL)
		} else {
			fmt.Printf("WARNING: Invalid proxy URL: %s\n", config.ProxyURL)
		}
	}

	return &ResilientHTTPClient{
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		config: config,
	}
}

// DoWithRetry executes an HTTP request with exponential backoff retry logic
func (c *ResilientHTTPClient) DoWithRetry(req *http.Request) (*http.Response, error) {
	// Add delay since last request to avoid rate limiting
	if !c.lastRequest.IsZero() && c.config.RequestDelay > 0 {
		elapsed := time.Since(c.lastRequest)
		if elapsed < c.config.RequestDelay {
			time.Sleep(c.config.RequestDelay - elapsed)
		}
	}
	c.lastRequest = time.Now()

	var lastErr error
	delay := c.config.InitialDelay

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		// Clone the request for retry attempts
		reqClone := req.Clone(req.Context())

		resp, err := c.client.Do(reqClone)
		if err == nil {
			// Success - check status code
			if resp.StatusCode == http.StatusOK {
				return resp, nil
			}

			// Handle 403 specifically - might be temporary
			if resp.StatusCode == 403 && attempt < c.config.MaxRetries {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				lastErr = fmt.Errorf("API returned status 403: %s", string(body))

				// Exponential backoff with jitter
				jitter := time.Duration(rand.Int63n(int64(delay / 2)))
				sleepTime := delay + jitter

				if attempt < c.config.MaxRetries {
					fmt.Printf("INFO: Attempt %d/%d failed with 403, retrying in %v...\n",
						attempt+1, c.config.MaxRetries+1, sleepTime)
					time.Sleep(sleepTime)

					// Exponential backoff
					delay *= 2
					if delay > c.config.MaxDelay {
						delay = c.config.MaxDelay
					}
					continue
				}
			}

			// For other status codes or last attempt, return the response
			return resp, nil
		}

		lastErr = err

		// Network error - retry with backoff
		if attempt < c.config.MaxRetries {
			jitter := time.Duration(rand.Int63n(int64(delay / 2)))
			sleepTime := delay + jitter

			fmt.Printf("INFO: Attempt %d/%d failed: %v, retrying in %v...\n",
				attempt+1, c.config.MaxRetries+1, err, sleepTime)
			time.Sleep(sleepTime)

			// Exponential backoff
			delay *= 2
			if delay > c.config.MaxDelay {
				delay = c.config.MaxDelay
			}
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.config.MaxRetries+1, lastErr)
}

// GetClient returns the underlying http.Client
func (c *ResilientHTTPClient) GetClient() *http.Client {
	return c.client
}
