package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arandu-ai/arandu/config"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAPIKeyAuth_NoKeyRequired(t *testing.T) {
	// Reset config
	config.Config.RequireAPIKey = false
	config.Config.APIKey = ""

	router := gin.New()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAPIKeyAuth_ValidKey(t *testing.T) {
	config.Config.RequireAPIKey = true
	config.Config.APIKey = "test-api-key-12345"

	router := gin.New()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "test-api-key-12345")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	config.Config.RequireAPIKey = true
	config.Config.APIKey = "test-api-key-12345"

	router := gin.New()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	config.Config.RequireAPIKey = true
	config.Config.APIKey = "test-api-key-12345"

	router := gin.New()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAPIKeyAuth_BearerToken(t *testing.T) {
	config.Config.RequireAPIKey = true
	config.Config.APIKey = "test-api-key-12345"

	router := gin.New()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer test-api-key-12345")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAPIKeyAuth_HealthEndpointBypass(t *testing.T) {
	config.Config.RequireAPIKey = true
	config.Config.APIKey = "test-api-key-12345"

	router := gin.New()
	router.Use(APIKeyAuth())
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for health endpoint, got %d", w.Code)
	}
}

func TestSecureCompare(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected bool
	}{
		{"abc", "abc", true},
		{"abc", "abd", false},
		{"abc", "ab", false},
		{"", "", true},
		{"abc", "", false},
	}

	for _, tt := range tests {
		result := secureCompare(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("secureCompare(%q, %q) = %v, want %v", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestOptionalAPIKeyAuth_NoKey(t *testing.T) {
	config.Config.APIKey = "test-api-key"

	router := gin.New()
	router.Use(OptionalAPIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		authenticated, _ := c.Get("authenticated")
		if authenticated == true {
			c.String(http.StatusOK, "authenticated")
		} else {
			c.String(http.StatusOK, "anonymous")
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "anonymous" {
		t.Errorf("Expected 'anonymous', got %s", w.Body.String())
	}
}

func TestOptionalAPIKeyAuth_ValidKey(t *testing.T) {
	config.Config.APIKey = "test-api-key"

	router := gin.New()
	router.Use(OptionalAPIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		authenticated, _ := c.Get("authenticated")
		if authenticated == true {
			c.String(http.StatusOK, "authenticated")
		} else {
			c.String(http.StatusOK, "anonymous")
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "test-api-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "authenticated" {
		t.Errorf("Expected 'authenticated', got %s", w.Body.String())
	}
}
