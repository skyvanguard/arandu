package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewChecker(t *testing.T) {
	checker := NewChecker("1.0.0", nil)

	if checker == nil {
		t.Fatal("NewChecker() returned nil")
	}

	if checker.version != "1.0.0" {
		t.Errorf("version = %v, want 1.0.0", checker.version)
	}

	if checker.startTime.IsZero() {
		t.Error("startTime should not be zero")
	}
}

func TestChecker_IncrementDecrement(t *testing.T) {
	checker := NewChecker("1.0.0", nil)

	checker.IncrementRequests()
	checker.IncrementRequests()

	metrics := checker.GetMetrics()
	if metrics.RequestsTotal != 2 {
		t.Errorf("RequestsTotal = %d, want 2", metrics.RequestsTotal)
	}
	if metrics.RequestsActive != 2 {
		t.Errorf("RequestsActive = %d, want 2", metrics.RequestsActive)
	}

	checker.DecrementRequests()
	metrics = checker.GetMetrics()
	if metrics.RequestsActive != 1 {
		t.Errorf("RequestsActive = %d, want 1", metrics.RequestsActive)
	}
}

func TestChecker_IncrementErrors(t *testing.T) {
	checker := NewChecker("1.0.0", nil)

	checker.IncrementErrors()
	checker.IncrementErrors()

	metrics := checker.GetMetrics()
	if metrics.ErrorsTotal != 2 {
		t.Errorf("ErrorsTotal = %d, want 2", metrics.ErrorsTotal)
	}
}

func TestChecker_GetMetrics(t *testing.T) {
	checker := NewChecker("1.0.0", nil)

	// Wait a bit to have measurable uptime
	time.Sleep(10 * time.Millisecond)

	metrics := checker.GetMetrics()

	if metrics.UptimeSeconds <= 0 {
		t.Error("UptimeSeconds should be positive")
	}
	if metrics.GoRoutines <= 0 {
		t.Error("GoRoutines should be positive")
	}
	if metrics.NumCPU <= 0 {
		t.Error("NumCPU should be positive")
	}
	if metrics.MemoryAllocMB <= 0 {
		t.Error("MemoryAllocMB should be positive")
	}
}

func TestChecker_CheckWithoutDB(t *testing.T) {
	checker := NewChecker("1.0.0", nil)

	health := checker.Check(context.Background())

	if health.Status != StatusUnhealthy {
		t.Errorf("Status = %v, want unhealthy (no db)", health.Status)
	}
	if health.Version != "1.0.0" {
		t.Errorf("Version = %v, want 1.0.0", health.Version)
	}
	if len(health.Components) != 1 {
		t.Errorf("Expected 1 component, got %d", len(health.Components))
	}
}

func TestChecker_LivenessHandler(t *testing.T) {
	checker := NewChecker("1.0.0", nil)

	router := gin.New()
	router.GET("/livez", checker.LivenessHandler())

	req := httptest.NewRequest("GET", "/livez", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestChecker_ReadinessHandler_Unhealthy(t *testing.T) {
	checker := NewChecker("1.0.0", nil)

	router := gin.New()
	router.GET("/readyz", checker.ReadinessHandler())

	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Without DB, should be unhealthy
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}

func TestChecker_MetricsHandler(t *testing.T) {
	checker := NewChecker("1.0.0", nil)

	router := gin.New()
	router.GET("/metrics", checker.MetricsHandler())

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestChecker_RequestCounterMiddleware(t *testing.T) {
	checker := NewChecker("1.0.0", nil)

	router := gin.New()
	router.Use(checker.RequestCounterMiddleware())
	router.GET("/test", func(c *gin.Context) {
		// Check that request is counted during handling
		metrics := checker.GetMetrics()
		if metrics.RequestsActive != 1 {
			t.Errorf("RequestsActive during request = %d, want 1", metrics.RequestsActive)
		}
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	metrics := checker.GetMetrics()
	if metrics.RequestsTotal != 1 {
		t.Errorf("RequestsTotal = %d, want 1", metrics.RequestsTotal)
	}
	if metrics.RequestsActive != 0 {
		t.Errorf("RequestsActive after request = %d, want 0", metrics.RequestsActive)
	}
}

func TestChecker_RequestCounterMiddleware_TrackErrors(t *testing.T) {
	checker := NewChecker("1.0.0", nil)

	router := gin.New()
	router.Use(checker.RequestCounterMiddleware())
	router.GET("/error", func(c *gin.Context) {
		c.String(http.StatusInternalServerError, "error")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	metrics := checker.GetMetrics()
	if metrics.ErrorsTotal != 1 {
		t.Errorf("ErrorsTotal = %d, want 1", metrics.ErrorsTotal)
	}
}

func TestStatus_Values(t *testing.T) {
	if StatusHealthy != "healthy" {
		t.Error("StatusHealthy should be 'healthy'")
	}
	if StatusUnhealthy != "unhealthy" {
		t.Error("StatusUnhealthy should be 'unhealthy'")
	}
	if StatusDegraded != "degraded" {
		t.Error("StatusDegraded should be 'degraded'")
	}
}
