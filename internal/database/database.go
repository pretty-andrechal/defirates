package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pretty-andrechal/defirates/internal/models"
)

type DB struct {
	conn *sql.DB
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// migrate creates the database schema
func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS protocols (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		url TEXT,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS yield_rates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		protocol_id INTEGER NOT NULL,
		asset TEXT NOT NULL,
		chain TEXT NOT NULL,
		apy REAL NOT NULL,
		tvl REAL NOT NULL,
		maturity_date DATETIME,
		pool_name TEXT NOT NULL,
		categories TEXT,
		external_url TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (protocol_id) REFERENCES protocols(id)
	);

	CREATE INDEX IF NOT EXISTS idx_yield_rates_protocol ON yield_rates(protocol_id);
	CREATE INDEX IF NOT EXISTS idx_yield_rates_apy ON yield_rates(apy);
	CREATE INDEX IF NOT EXISTS idx_yield_rates_asset ON yield_rates(asset);
	CREATE INDEX IF NOT EXISTS idx_yield_rates_chain ON yield_rates(chain);
	CREATE INDEX IF NOT EXISTS idx_yield_rates_categories ON yield_rates(categories);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// CreateOrUpdateProtocol creates or updates a protocol
func (db *DB) CreateOrUpdateProtocol(protocol *models.Protocol) error {
	query := `
		INSERT INTO protocols (name, url, description)
		VALUES (?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			url = excluded.url,
			description = excluded.description
		RETURNING id, created_at
	`

	return db.conn.QueryRow(
		query,
		protocol.Name,
		protocol.URL,
		protocol.Description,
	).Scan(&protocol.ID, &protocol.CreatedAt)
}

// GetProtocolByName retrieves a protocol by name
func (db *DB) GetProtocolByName(name string) (*models.Protocol, error) {
	protocol := &models.Protocol{}
	query := `SELECT id, name, url, description, created_at FROM protocols WHERE name = ?`

	err := db.conn.QueryRow(query, name).Scan(
		&protocol.ID,
		&protocol.Name,
		&protocol.URL,
		&protocol.Description,
		&protocol.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return protocol, nil
}

// UpsertYieldRate creates or updates a yield rate
func (db *DB) UpsertYieldRate(rate *models.YieldRate) error {
	// First, check if this exact pool already exists
	var existingID int64
	checkQuery := `
		SELECT id FROM yield_rates
		WHERE protocol_id = ? AND pool_name = ? AND chain = ?
	`
	err := db.conn.QueryRow(checkQuery, rate.ProtocolID, rate.PoolName, rate.Chain).Scan(&existingID)

	now := time.Now()
	if err == sql.ErrNoRows {
		// Insert new record
		query := `
			INSERT INTO yield_rates (protocol_id, asset, chain, apy, tvl, maturity_date, pool_name, categories, external_url, updated_at, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			RETURNING id
		`
		return db.conn.QueryRow(
			query,
			rate.ProtocolID,
			rate.Asset,
			rate.Chain,
			rate.APY,
			rate.TVL,
			rate.MaturityDate,
			rate.PoolName,
			rate.Categories,
			rate.ExternalURL,
			now,
			now,
		).Scan(&rate.ID)
	} else if err != nil {
		return err
	}

	// Update existing record
	query := `
		UPDATE yield_rates
		SET asset = ?, apy = ?, tvl = ?, maturity_date = ?, categories = ?, external_url = ?, updated_at = ?
		WHERE id = ?
	`
	_, err = db.conn.Exec(
		query,
		rate.Asset,
		rate.APY,
		rate.TVL,
		rate.MaturityDate,
		rate.Categories,
		rate.ExternalURL,
		now,
		existingID,
	)
	rate.ID = existingID
	return err
}

// GetYieldRates retrieves yield rates with optional filtering
func (db *DB) GetYieldRates(filters models.FilterParams) ([]models.YieldRate, error) {
	query := `
		SELECT
			yr.id, yr.protocol_id, p.name as protocol_name, yr.asset, yr.chain,
			yr.apy, yr.tvl, yr.maturity_date, yr.pool_name, yr.categories, yr.external_url,
			yr.updated_at, yr.created_at
		FROM yield_rates yr
		JOIN protocols p ON yr.protocol_id = p.id
		WHERE 1=1
	`

	args := []interface{}{}

	if filters.MinAPY > 0 {
		query += " AND yr.apy >= ?"
		args = append(args, filters.MinAPY)
	}

	if filters.MaxAPY > 0 {
		query += " AND yr.apy <= ?"
		args = append(args, filters.MaxAPY)
	}

	if filters.MinTVL > 0 {
		query += " AND yr.tvl >= ?"
		args = append(args, filters.MinTVL)
	}

	if filters.Asset != "" {
		query += " AND yr.asset = ?"
		args = append(args, filters.Asset)
	}

	if filters.Chain != "" {
		query += " AND yr.chain = ?"
		args = append(args, filters.Chain)
	}

	if filters.ProtocolName != "" {
		query += " AND p.name = ?"
		args = append(args, filters.ProtocolName)
	}

	if filters.Categories != "" {
		query += " AND yr.categories LIKE ?"
		args = append(args, "%"+filters.Categories+"%")
	}

	// Sorting
	sortBy := "yr.apy"
	if filters.SortBy != "" {
		switch filters.SortBy {
		case "apy":
			sortBy = "yr.apy"
		case "tvl":
			sortBy = "yr.tvl"
		case "updated_at":
			sortBy = "yr.updated_at"
		}
	}

	sortOrder := "DESC"
	if strings.ToUpper(filters.SortOrder) == "ASC" {
		sortOrder = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []models.YieldRate
	for rows.Next() {
		var rate models.YieldRate
		var maturityDate sql.NullTime
		var categories sql.NullString

		err := rows.Scan(
			&rate.ID,
			&rate.ProtocolID,
			&rate.ProtocolName,
			&rate.Asset,
			&rate.Chain,
			&rate.APY,
			&rate.TVL,
			&maturityDate,
			&rate.PoolName,
			&categories,
			&rate.ExternalURL,
			&rate.UpdatedAt,
			&rate.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if maturityDate.Valid {
			rate.MaturityDate = &maturityDate.Time
		}

		if categories.Valid {
			rate.Categories = categories.String
		}

		rates = append(rates, rate)
	}

	return rates, rows.Err()
}

// GetDistinctAssets returns all unique assets
func (db *DB) GetDistinctAssets() ([]string, error) {
	query := `SELECT DISTINCT asset FROM yield_rates ORDER BY asset`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []string
	for rows.Next() {
		var asset string
		if err := rows.Scan(&asset); err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}

	return assets, rows.Err()
}

// GetDistinctChains returns all unique chains
func (db *DB) GetDistinctChains() ([]string, error) {
	query := `SELECT DISTINCT chain FROM yield_rates ORDER BY chain`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chains []string
	for rows.Next() {
		var chain string
		if err := rows.Scan(&chain); err != nil {
			return nil, err
		}
		chains = append(chains, chain)
	}

	return chains, rows.Err()
}

// GetDistinctCategories returns all unique categories (flattened from comma-separated values)
func (db *DB) GetDistinctCategories() ([]string, error) {
	query := `SELECT DISTINCT categories FROM yield_rates WHERE categories IS NOT NULL AND categories != '' ORDER BY categories`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categoriesSet := make(map[string]bool)
	for rows.Next() {
		var categoriesStr string
		if err := rows.Scan(&categoriesStr); err != nil {
			return nil, err
		}
		// Split comma-separated categories and add to set
		for _, cat := range strings.Split(categoriesStr, ",") {
			trimmed := strings.TrimSpace(cat)
			if trimmed != "" {
				categoriesSet[trimmed] = true
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Convert set to sorted slice
	categories := make([]string, 0, len(categoriesSet))
	for cat := range categoriesSet {
		categories = append(categories, cat)
	}

	// Sort alphabetically
	sortedCategories := categories
	for i := 0; i < len(sortedCategories)-1; i++ {
		for j := i + 1; j < len(sortedCategories); j++ {
			if sortedCategories[i] > sortedCategories[j] {
				sortedCategories[i], sortedCategories[j] = sortedCategories[j], sortedCategories[i]
			}
		}
	}

	return sortedCategories, nil
}

// GetYieldRatesByIDs retrieves yield rates by their IDs
func (db *DB) GetYieldRatesByIDs(ids []int64) ([]models.YieldRate, error) {
	if len(ids) == 0 {
		return []models.YieldRate{}, nil
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT
			yr.id, yr.protocol_id, p.name as protocol_name, yr.asset, yr.chain,
			yr.apy, yr.tvl, yr.maturity_date, yr.pool_name, yr.categories, yr.external_url,
			yr.updated_at, yr.created_at
		FROM yield_rates yr
		JOIN protocols p ON yr.protocol_id = p.id
		WHERE yr.id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []models.YieldRate
	for rows.Next() {
		var rate models.YieldRate
		var maturityDate sql.NullTime
		var categories sql.NullString

		err := rows.Scan(
			&rate.ID,
			&rate.ProtocolID,
			&rate.ProtocolName,
			&rate.Asset,
			&rate.Chain,
			&rate.APY,
			&rate.TVL,
			&maturityDate,
			&rate.PoolName,
			&categories,
			&rate.ExternalURL,
			&rate.UpdatedAt,
			&rate.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if maturityDate.Valid {
			rate.MaturityDate = &maturityDate.Time
		}

		if categories.Valid {
			rate.Categories = categories.String
		}

		rates = append(rates, rate)
	}

	return rates, rows.Err()
}
