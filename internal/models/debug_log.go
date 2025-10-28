package models

import "time"

// HTTPDebugLog represents an HTTP request/response debug log entry
type HTTPDebugLog struct {
	ID              int64     `json:"id"`
	Timestamp       time.Time `json:"timestamp"`
	Method          string    `json:"method"`
	URL             string    `json:"url"`
	RequestHeaders  string    `json:"request_headers"`
	RequestBody     string    `json:"request_body"`
	ResponseStatus  int       `json:"response_status"`
	ResponseHeaders string    `json:"response_headers"`
	ResponseBody    string    `json:"response_body"`
	Error           string    `json:"error"`
	DurationMS      int64     `json:"duration_ms"`
	Source          string    `json:"source"` // e.g., "beefy", "pendle"
}
