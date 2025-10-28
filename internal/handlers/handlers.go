package handlers

import (
	"embed"
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
	db        *database.DB
	templates *template.Template
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
		db:        db,
		templates: tmpl,
	}, nil
}

// parseFilterParams extracts filter parameters from the request
func (h *Handler) parseFilterParams(r *http.Request) models.FilterParams {
	filters := models.FilterParams{
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
		Asset:     r.URL.Query().Get("asset"),
		Chain:     r.URL.Query().Get("chain"),
		ProtocolName: r.URL.Query().Get("protocol"),
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

	data := struct {
		YieldRates []models.YieldRate
		Assets     []string
		Chains     []string
		Filters    models.FilterParams
	}{
		YieldRates: rates,
		Assets:     assets,
		Chains:     chains,
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
