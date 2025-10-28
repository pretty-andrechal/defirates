package models

import "time"

// Protocol represents a DeFi protocol
type Protocol struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// YieldRate represents a yield opportunity from a protocol
type YieldRate struct {
	ID           int64     `json:"id"`
	ProtocolID   int64     `json:"protocol_id"`
	ProtocolName string    `json:"protocol_name"`
	Asset        string    `json:"asset"`        // e.g., "ETH", "USDC"
	Chain        string    `json:"chain"`        // e.g., "Ethereum", "Arbitrum"
	APY          float64   `json:"apy"`          // Annual Percentage Yield
	TVL          float64   `json:"tvl"`          // Total Value Locked
	MaturityDate *time.Time `json:"maturity_date,omitempty"` // For fixed-term yields like Pendle
	PoolName     string    `json:"pool_name"`    // Specific pool identifier
	Categories   string    `json:"categories"`   // Comma-separated categories (e.g., "PT", "Liquidity")
	ExternalURL  string    `json:"external_url"` // Link to the actual pool
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// FilterParams for querying yield rates
type FilterParams struct {
	MinAPY       float64
	MaxAPY       float64
	MinTVL       float64
	Asset        string
	Chain        string
	ProtocolName string
	Categories   string
	SortBy       string // "apy", "tvl", "updated_at"
	SortOrder    string // "asc", "desc"
}
