package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// EventManager manages Server-Sent Events connections
type EventManager struct {
	clients map[chan string]bool
	mu      sync.RWMutex
}

// NewEventManager creates a new event manager
func NewEventManager() *EventManager {
	return &EventManager{
		clients: make(map[chan string]bool),
	}
}

// Register adds a new client connection
func (em *EventManager) Register(client chan string) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.clients[client] = true
	log.Printf("SSE client registered. Total clients: %d", len(em.clients))
}

// Unregister removes a client connection
func (em *EventManager) Unregister(client chan string) {
	em.mu.Lock()
	defer em.mu.Unlock()
	if _, exists := em.clients[client]; exists {
		delete(em.clients, client)
		close(client)
		log.Printf("SSE client unregistered. Total clients: %d", len(em.clients))
	}
}

// Broadcast sends an event to all connected clients
func (em *EventManager) Broadcast(eventType string, data interface{}) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if len(em.clients) == 0 {
		return // No clients connected
	}

	// Marshal data to JSON
	var message string
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Printf("Failed to marshal SSE data: %v", err)
			return
		}
		message = fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(jsonData))
	} else {
		message = fmt.Sprintf("event: %s\ndata: {}\n\n", eventType)
	}

	// Send to all clients
	for client := range em.clients {
		select {
		case client <- message:
			// Message sent successfully
		case <-time.After(1 * time.Second):
			// Client not responding, will be cleaned up on next request
			log.Printf("Client not responding, skipping")
		}
	}

	log.Printf("Broadcasted %s event to %d clients", eventType, len(em.clients))
}

// BroadcastDataUpdate sends a data update event to all clients
func (em *EventManager) BroadcastDataUpdate() {
	em.Broadcast("update", map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"message":   "Data has been updated",
	})
}
