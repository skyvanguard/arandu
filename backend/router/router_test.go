package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/arandu-ai/arandu/config"
	"github.com/arandu-ai/arandu/database"
)

func TestMain(m *testing.M) {
	// Setup: configure environment for tests
	os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")
	os.Setenv("PORT", "8080")
	os.Setenv("DATABASE_URL", ":memory:")

	// Initialize config
	config.Init()

	// Run tests
	code := m.Run()

	// Cleanup
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")

	os.Exit(code)
}

func TestRouterInitialization(t *testing.T) {
	var db *database.Queries

	r := New(db)

	if r == nil {
		t.Error("Router should not be nil")
	}
}

func TestPlaygroundEndpoint(t *testing.T) {
	var db *database.Queries

	r := New(db)

	req, err := http.NewRequest("GET", "/playground", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Playground endpoint returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestGraphQLEndpointAcceptsPost(t *testing.T) {
	var db *database.Queries

	r := New(db)

	req, err := http.NewRequest("POST", "/graphql", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Should not be 404 or 405
	if status := rr.Code; status == http.StatusNotFound || status == http.StatusMethodNotAllowed {
		t.Errorf("GraphQL endpoint should accept POST, got status %d", status)
	}
}

func TestCORSPreflight(t *testing.T) {
	var db *database.Queries

	r := New(db)

	req, err := http.NewRequest("OPTIONS", "/graphql", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Check CORS headers are present for allowed origin
	origin := rr.Header().Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:5173" {
		t.Errorf("Expected CORS origin 'http://localhost:5173', got '%s'", origin)
	}

	methods := rr.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Error("Access-Control-Allow-Methods header not set")
	}
}

func TestUnknownOriginBlocked(t *testing.T) {
	var db *database.Queries

	r := New(db)

	req, err := http.NewRequest("OPTIONS", "/graphql", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://malicious-site.com")
	req.Header.Set("Access-Control-Request-Method", "POST")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Unknown origin should not get CORS headers
	origin := rr.Header().Get("Access-Control-Allow-Origin")
	if origin == "http://malicious-site.com" {
		t.Error("CORS should not allow unknown origins")
	}
}
