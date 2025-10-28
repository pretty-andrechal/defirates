package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pretty-andrechal/defirates/internal/database"
	"github.com/pretty-andrechal/defirates/internal/models"
)

// DebugHTTPClient wraps http.Client to log all requests and responses
type DebugHTTPClient struct {
	client  *http.Client
	db      *database.DB
	source  string
	enabled bool
}

// NewDebugHTTPClient creates a new debug HTTP client
func NewDebugHTTPClient(client *http.Client, db *database.DB, source string, enabled bool) *DebugHTTPClient {
	return &DebugHTTPClient{
		client:  client,
		db:      db,
		source:  source,
		enabled: enabled,
	}
}

// Do executes an HTTP request and logs it to the database
func (c *DebugHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if !c.enabled || c.db == nil {
		// Debug logging disabled, just execute normally
		return c.client.Do(req)
	}

	startTime := time.Now()

	// Capture request details
	requestHeaders := c.formatHeaders(req.Header)
	requestBody := ""
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			requestBody = string(bodyBytes)
			// Restore the body so it can be read again
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	// Execute request
	resp, err := c.client.Do(req)
	duration := time.Since(startTime)

	// Prepare log entry
	log := &models.HTTPDebugLog{
		Timestamp:      startTime,
		Method:         req.Method,
		URL:            req.URL.String(),
		RequestHeaders: requestHeaders,
		RequestBody:    requestBody,
		DurationMS:     duration.Milliseconds(),
		Source:         c.source,
	}

	if err != nil {
		// Request failed
		log.Error = err.Error()
		c.storeLog(log)
		return nil, err
	}

	// Capture response details
	log.ResponseStatus = resp.StatusCode
	log.ResponseHeaders = c.formatHeaders(resp.Header)

	// Read response body
	if resp.Body != nil {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if readErr == nil {
			// Limit body size to 100KB for storage
			if len(bodyBytes) > 100*1024 {
				log.ResponseBody = string(bodyBytes[:100*1024]) + "\n... (truncated)"
			} else {
				log.ResponseBody = string(bodyBytes)
			}
			// Restore the body so it can be read by the caller
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		} else {
			log.Error = fmt.Sprintf("Failed to read response body: %v", readErr)
		}
	}

	// Store log entry asynchronously to avoid blocking
	go c.storeLog(log)

	return resp, nil
}

// formatHeaders converts http.Header to a readable string
func (c *DebugHTTPClient) formatHeaders(headers http.Header) string {
	var builder strings.Builder
	for key, values := range headers {
		builder.WriteString(fmt.Sprintf("%s: %s\n", key, strings.Join(values, ", ")))
	}
	return builder.String()
}

// storeLog stores the log entry to the database
func (c *DebugHTTPClient) storeLog(log *models.HTTPDebugLog) {
	if err := c.db.StoreHTTPDebugLog(log); err != nil {
		// Don't fail the request if logging fails, just print error
		fmt.Printf("WARNING: Failed to store HTTP debug log: %v\n", err)
	}
}

// GetClient returns the underlying http.Client
func (c *DebugHTTPClient) GetClient() *http.Client {
	return c.client
}
