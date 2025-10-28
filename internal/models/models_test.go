package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestYieldRate_JSONMarshaling(t *testing.T) {
	maturityDate := time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)

	rate := YieldRate{
		ID:           1,
		ProtocolID:   10,
		ProtocolName: "Pendle",
		Asset:        "ETH",
		Chain:        "Ethereum",
		APY:          12.5,
		TVL:          1000000.0,
		MaturityDate: &maturityDate,
		PoolName:     "wstETH Pool",
		ExternalURL:  "https://example.com",
		UpdatedAt:    time.Now(),
		CreatedAt:    time.Now(),
	}

	// Marshal to JSON
	data, err := json.Marshal(rate)
	if err != nil {
		t.Fatalf("Failed to marshal YieldRate: %v", err)
	}

	// Unmarshal back
	var decoded YieldRate
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal YieldRate: %v", err)
	}

	// Verify fields
	if decoded.ID != rate.ID {
		t.Errorf("ID = %d, want %d", decoded.ID, rate.ID)
	}
	if decoded.ProtocolName != rate.ProtocolName {
		t.Errorf("ProtocolName = %s, want %s", decoded.ProtocolName, rate.ProtocolName)
	}
	if decoded.Asset != rate.Asset {
		t.Errorf("Asset = %s, want %s", decoded.Asset, rate.Asset)
	}
	if decoded.APY != rate.APY {
		t.Errorf("APY = %f, want %f", decoded.APY, rate.APY)
	}
	if decoded.MaturityDate == nil {
		t.Error("MaturityDate should not be nil")
	}
}

func TestYieldRate_NilMaturityDate(t *testing.T) {
	rate := YieldRate{
		ID:           1,
		ProtocolID:   10,
		ProtocolName: "Pendle",
		Asset:        "ETH",
		Chain:        "Ethereum",
		APY:          12.5,
		TVL:          1000000.0,
		MaturityDate: nil, // No maturity date
		PoolName:     "Perpetual Pool",
		ExternalURL:  "https://example.com",
	}

	// Marshal to JSON
	data, err := json.Marshal(rate)
	if err != nil {
		t.Fatalf("Failed to marshal YieldRate: %v", err)
	}

	// Unmarshal back
	var decoded YieldRate
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal YieldRate: %v", err)
	}

	// Verify maturity_date is omitted in JSON
	var jsonMap map[string]interface{}
	json.Unmarshal(data, &jsonMap)
	if _, exists := jsonMap["maturity_date"]; !exists {
		t.Log("maturity_date correctly omitted when nil")
	}

	if decoded.MaturityDate != nil {
		t.Errorf("MaturityDate should be nil, got %v", decoded.MaturityDate)
	}
}

func TestProtocol_JSONMarshaling(t *testing.T) {
	protocol := Protocol{
		ID:          1,
		Name:        "Pendle",
		URL:         "https://pendle.finance",
		Description: "Fixed-rate yield protocol",
		CreatedAt:   time.Now(),
	}

	// Marshal to JSON
	data, err := json.Marshal(protocol)
	if err != nil {
		t.Fatalf("Failed to marshal Protocol: %v", err)
	}

	// Unmarshal back
	var decoded Protocol
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal Protocol: %v", err)
	}

	// Verify fields
	if decoded.Name != protocol.Name {
		t.Errorf("Name = %s, want %s", decoded.Name, protocol.Name)
	}
	if decoded.URL != protocol.URL {
		t.Errorf("URL = %s, want %s", decoded.URL, protocol.URL)
	}
	if decoded.Description != protocol.Description {
		t.Errorf("Description = %s, want %s", decoded.Description, protocol.Description)
	}
}

func TestFilterParams_DefaultValues(t *testing.T) {
	// Test that FilterParams has sensible zero values
	filters := FilterParams{}

	if filters.MinAPY != 0 {
		t.Errorf("MinAPY default should be 0, got %f", filters.MinAPY)
	}
	if filters.MaxAPY != 0 {
		t.Errorf("MaxAPY default should be 0, got %f", filters.MaxAPY)
	}
	if filters.MinTVL != 0 {
		t.Errorf("MinTVL default should be 0, got %f", filters.MinTVL)
	}
	if filters.Asset != "" {
		t.Errorf("Asset default should be empty, got %s", filters.Asset)
	}
	if filters.Chain != "" {
		t.Errorf("Chain default should be empty, got %s", filters.Chain)
	}
}

func TestFilterParams_WithValues(t *testing.T) {
	filters := FilterParams{
		MinAPY:       5.0,
		MaxAPY:       20.0,
		MinTVL:       1000000.0,
		Asset:        "ETH",
		Chain:        "Ethereum",
		ProtocolName: "Pendle",
		SortBy:       "apy",
		SortOrder:    "desc",
	}

	// Verify all fields are set
	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"MinAPY", filters.MinAPY, 5.0},
		{"MaxAPY", filters.MaxAPY, 20.0},
		{"MinTVL", filters.MinTVL, 1000000.0},
		{"Asset", filters.Asset, "ETH"},
		{"Chain", filters.Chain, "Ethereum"},
		{"ProtocolName", filters.ProtocolName, "Pendle"},
		{"SortBy", filters.SortBy, "apy"},
		{"SortOrder", filters.SortOrder, "desc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}
