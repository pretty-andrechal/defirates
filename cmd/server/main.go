package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/pretty-andrechal/defirates/internal/api"
	"github.com/pretty-andrechal/defirates/internal/database"
	"github.com/pretty-andrechal/defirates/internal/handlers"
)

func main() {
	// Parse command-line flags
	port := flag.String("port", "8080", "Port to run the server on")
	dbPath := flag.String("db", "defirates.db", "Path to SQLite database")
	fetchInterval := flag.Duration("fetch-interval", 5*time.Minute, "Interval for fetching yield data")
	loadSample := flag.Bool("load-sample", false, "Load sample data for demonstration")
	flag.Parse()

	log.Println("Starting DeFi Rates server...")

	// Initialize database
	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Printf("Database initialized at %s", *dbPath)

	// Load sample data if requested
	if *loadSample {
		if err := api.LoadSampleData(db); err != nil {
			log.Printf("Warning: Failed to load sample data: %v", err)
		}
	}

	// Initialize HTTP handlers
	handler, err := handlers.New(db)
	if err != nil {
		log.Fatalf("Failed to initialize handlers: %v", err)
	}

	// Initialize data fetcher and wire up SSE callback
	fetcher := api.NewFetcher(db)
	fetcher.SetOnDataUpdateCallback(func() {
		handler.GetEventManager().BroadcastDataUpdate()
	})
	fetcher.StartPeriodicFetch(*fetchInterval)
	log.Printf("Data fetcher started (interval: %v)", *fetchInterval)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.HandleIndex)
	mux.HandleFunc("/events", handler.HandleEvents)
	mux.HandleFunc("/api/rates", handler.HandleAPIRates)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start server
	addr := ":" + *port
	log.Printf("Server starting on http://localhost%s", addr)
	log.Printf("Open your browser and navigate to http://localhost%s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // No timeout for SSE connections
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
