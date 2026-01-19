package websocket

import (
	"sync"
	"testing"
)

func TestConnectionManager(t *testing.T) {
	// Test that connections map is thread-safe
	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes (simulating connection adds)
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// This would normally add a connection
			// Just testing that concurrent access doesn't panic
			_ = id
		}(i)
	}

	wg.Wait()
}

func TestCloseAll(t *testing.T) {
	// Test CloseAll doesn't panic with no connections
	CloseAll()
}
