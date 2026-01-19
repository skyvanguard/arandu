package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arandu-ai/arandu/database"
)

func TestHealthEndpoint(t *testing.T) {
	// Create a mock database queries object
	// In a real test, you'd use a test database
	var db *database.Queries = nil

	r := New(db)

	// Test health endpoint
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Health endpoint returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestPlaygroundEndpoint(t *testing.T) {
	var db *database.Queries = nil

	r := New(db)

	req, err := http.NewRequest("GET", "/playground", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Playground should return 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Playground endpoint returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestGraphQLEndpointExists(t *testing.T) {
	var db *database.Queries = nil

	r := New(db)

	// POST to /graphql should work (even if query is invalid, endpoint exists)
	req, err := http.NewRequest("POST", "/graphql", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Should not be 404
	if status := rr.Code; status == http.StatusNotFound {
		t.Error("GraphQL endpoint not found")
	}
}

func TestCORSHeaders(t *testing.T) {
	var db *database.Queries = nil

	r := New(db)

	req, err := http.NewRequest("OPTIONS", "/graphql", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Check CORS headers are present
	if rr.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("CORS headers not set")
	}
}
