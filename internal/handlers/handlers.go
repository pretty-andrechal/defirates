package handlers

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/pretty-andrechal/defirates/internal/database"
	"github.com/pretty-andrechal/defirates/internal/models"
)

//go:embed templates
var templatesFS embed.FS

// Handler manages HTTP requests
type Handler struct {
	db           *database.DB
	templates    *template.Template
	eventManager *EventManager
}

// New creates a new handler
func New(db *database.DB) (*Handler, error) {
	// Create template functions
	funcMap := template.FuncMap{
		"divf": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
	}

	// Parse templates with functions
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		db:           db,
		templates:    tmpl,
		eventManager: NewEventManager(),
	}, nil
}

// parseFilterParams extracts filter parameters from the request
func (h *Handler) parseFilterParams(r *http.Request) models.FilterParams {
	filters := models.FilterParams{
		SortBy:       r.URL.Query().Get("sort_by"),
		SortOrder:    r.URL.Query().Get("sort_order"),
		Asset:        r.URL.Query().Get("asset"),
		Chain:        r.URL.Query().Get("chain"),
		ProtocolName: r.URL.Query().Get("protocol"),
		Categories:   r.URL.Query().Get("categories"),
	}

	if minAPY := r.URL.Query().Get("min_apy"); minAPY != "" {
		if val, err := strconv.ParseFloat(minAPY, 64); err == nil {
			filters.MinAPY = val
		}
	}

	if maxAPY := r.URL.Query().Get("max_apy"); maxAPY != "" {
		if val, err := strconv.ParseFloat(maxAPY, 64); err == nil {
			filters.MaxAPY = val
		}
	}

	if minTVL := r.URL.Query().Get("min_tvl"); minTVL != "" {
		if val, err := strconv.ParseFloat(minTVL, 64); err == nil {
			filters.MinTVL = val
		}
	}

	// Set defaults
	if filters.SortBy == "" {
		filters.SortBy = "apy"
	}
	if filters.SortOrder == "" {
		filters.SortOrder = "desc"
	}

	return filters
}

// HandleIndex serves the main page
func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	filters := h.parseFilterParams(r)

	rates, err := h.db.GetYieldRates(filters)
	if err != nil {
		log.Printf("Error fetching yield rates: %v", err)
		http.Error(w, "Failed to fetch yield rates", http.StatusInternalServerError)
		return
	}

	assets, err := h.db.GetDistinctAssets()
	if err != nil {
		log.Printf("Error fetching assets: %v", err)
		assets = []string{}
	}

	chains, err := h.db.GetDistinctChains()
	if err != nil {
		log.Printf("Error fetching chains: %v", err)
		chains = []string{}
	}

	categories, err := h.db.GetDistinctCategories()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		categories = []string{}
	}

	data := struct {
		YieldRates []models.YieldRate
		Assets     []string
		Chains     []string
		Categories []string
		Filters    models.FilterParams
	}{
		YieldRates: rates,
		Assets:     assets,
		Chains:     chains,
		Categories: categories,
		Filters:    filters,
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Return only the table partial
		if err := h.templates.ExecuteTemplate(w, "table.html", data); err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
		return
	}

	// Return full page
	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// HandleStatic serves static files
func (h *Handler) HandleStatic(w http.ResponseWriter, r *http.Request) {
	// Remove /static/ prefix
	path := strings.TrimPrefix(r.URL.Path, "/static/")
	http.ServeFile(w, r, "static/"+path)
}

// HandleEvents serves Server-Sent Events for real-time updates
func (h *Handler) HandleEvents(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for this client
	clientChan := make(chan string, 10)

	// Register the client
	h.eventManager.Register(clientChan)
	defer h.eventManager.Unregister(clientChan)

	// Send initial connection message
	fmt.Fprintf(w, "event: connected\ndata: {\"message\": \"Connected to real-time updates\"}\n\n")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// Listen for messages and client disconnect
	for {
		select {
		case msg := <-clientChan:
			// Send message to client
			fmt.Fprint(w, msg)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		case <-r.Context().Done():
			// Client disconnected
			return
		}
	}
}

// HandleAPIRates returns rates for specific IDs (for real-time updates)
func (h *Handler) HandleAPIRates(w http.ResponseWriter, r *http.Request) {
	// Get IDs from query parameter
	idsParam := r.URL.Query().Get("ids")
	if idsParam == "" {
		http.Error(w, "Missing 'ids' parameter", http.StatusBadRequest)
		return
	}

	// Parse comma-separated IDs
	idStrs := strings.Split(idsParam, ",")
	var ids []int64
	for _, idStr := range idStrs {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			continue // Skip invalid IDs
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		http.Error(w, "No valid IDs provided", http.StatusBadRequest)
		return
	}

	// Fetch rates by IDs
	rates, err := h.db.GetYieldRatesByIDs(ids)
	if err != nil {
		log.Printf("Error fetching rates by IDs: %v", err)
		http.Error(w, "Failed to fetch rates", http.StatusInternalServerError)
		return
	}

	// Return as JSON for HTMX to consume
	// We'll render the table rows HTML
	data := struct {
		YieldRates []models.YieldRate
	}{
		YieldRates: rates,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "rows.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// GetEventManager returns the event manager (for use by fetcher)
func (h *Handler) GetEventManager() *EventManager {
	return h.eventManager
}
