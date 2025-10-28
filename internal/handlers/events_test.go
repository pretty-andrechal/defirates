package handlers

import (
	"sync"
	"testing"
	"time"
)

func TestEventManager_RegisterUnregister(t *testing.T) {
	em := NewEventManager()

	// Create a client channel
	client := make(chan string, 10)

	// Register client
	em.Register(client)

	// Verify client is registered
	em.mu.RLock()
	if len(em.clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(em.clients))
	}
	if !em.clients[client] {
		t.Error("Client should be registered")
	}
	em.mu.RUnlock()

	// Unregister client
	em.Unregister(client)

	// Verify client is unregistered
	em.mu.RLock()
	if len(em.clients) != 0 {
		t.Errorf("Expected 0 clients, got %d", len(em.clients))
	}
	em.mu.RUnlock()

	// Verify channel is closed
	_, ok := <-client
	if ok {
		t.Error("Client channel should be closed")
	}
}

func TestEventManager_Broadcast(t *testing.T) {
	em := NewEventManager()

	// Create multiple client channels
	client1 := make(chan string, 10)
	client2 := make(chan string, 10)
	client3 := make(chan string, 10)

	em.Register(client1)
	em.Register(client2)
	em.Register(client3)

	// Broadcast a message
	testData := map[string]interface{}{
		"message": "test broadcast",
		"value":   42,
	}
	em.Broadcast("test", testData)

	// Give broadcast time to send
	time.Sleep(100 * time.Millisecond)

	// Verify all clients received the message
	clients := []chan string{client1, client2, client3}
	for i, client := range clients {
		select {
		case msg := <-client:
			if msg == "" {
				t.Errorf("Client %d received empty message", i)
			}
			// Check message format
			if len(msg) < 10 {
				t.Errorf("Client %d received invalid message: %s", i, msg)
			}
			// Should contain event type and data
			if !contains(msg, "event: test") {
				t.Errorf("Client %d message missing event type: %s", i, msg)
			}
			if !contains(msg, "data:") {
				t.Errorf("Client %d message missing data field: %s", i, msg)
			}
		case <-time.After(1 * time.Second):
			t.Errorf("Client %d did not receive message", i)
		}
	}

	// Clean up
	em.Unregister(client1)
	em.Unregister(client2)
	em.Unregister(client3)
}

func TestEventManager_BroadcastNoClients(t *testing.T) {
	em := NewEventManager()

	// Broadcast with no clients should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Broadcast panicked with no clients: %v", r)
		}
	}()

	em.Broadcast("test", map[string]string{"message": "test"})
}

func TestEventManager_ConcurrentAccess(t *testing.T) {
	em := NewEventManager()
	var wg sync.WaitGroup

	// Simulate concurrent registrations and broadcasts
	numClients := 10
	numBroadcasts := 5

	// Register clients concurrently
	clients := make([]chan string, numClients)
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			client := make(chan string, 10)
			clients[idx] = client
			em.Register(client)
		}(i)
	}
	wg.Wait()

	// Broadcast concurrently
	for i := 0; i < numBroadcasts; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			em.Broadcast("test", map[string]int{"broadcast": idx})
		}(i)
	}
	wg.Wait()

	// Give time for broadcasts to complete
	time.Sleep(200 * time.Millisecond)

	// Verify all clients are still registered
	em.mu.RLock()
	registeredCount := len(em.clients)
	em.mu.RUnlock()

	if registeredCount != numClients {
		t.Errorf("Expected %d clients, got %d", numClients, registeredCount)
	}

	// Unregister all clients
	for _, client := range clients {
		if client != nil {
			wg.Add(1)
			go func(c chan string) {
				defer wg.Done()
				em.Unregister(c)
			}(client)
		}
	}
	wg.Wait()

	// Verify all clients are unregistered
	em.mu.RLock()
	if len(em.clients) != 0 {
		t.Errorf("Expected 0 clients, got %d", len(em.clients))
	}
	em.mu.RUnlock()
}

func TestEventManager_BroadcastDataUpdate(t *testing.T) {
	em := NewEventManager()

	client := make(chan string, 10)
	em.Register(client)

	// Broadcast data update
	em.BroadcastDataUpdate()

	// Give broadcast time to send
	time.Sleep(100 * time.Millisecond)

	// Verify client received update event
	select {
	case msg := <-client:
		if !contains(msg, "event: update") {
			t.Errorf("Expected 'update' event, got: %s", msg)
		}
		if !contains(msg, "data:") {
			t.Errorf("Expected data field, got: %s", msg)
		}
	case <-time.After(1 * time.Second):
		t.Error("Client did not receive update event")
	}

	em.Unregister(client)
}

func TestEventManager_MultipleUnregister(t *testing.T) {
	em := NewEventManager()

	client := make(chan string, 10)
	em.Register(client)

	// Unregister multiple times should not panic
	em.Unregister(client)
	em.Unregister(client) // Second unregister should be safe

	em.mu.RLock()
	if len(em.clients) != 0 {
		t.Errorf("Expected 0 clients, got %d", len(em.clients))
	}
	em.mu.RUnlock()
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && hasSubstring(s, substr)
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
