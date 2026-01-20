package health

import (
	"context"
	"database/sql"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Name    string `json:"name"`
	Status  Status `json:"status"`
	Message string `json:"message,omitempty"`
	Latency int64  `json:"latency_ms,omitempty"`
}

// HealthResponse is the response for health check endpoints
type HealthResponse struct {
	Status     Status            `json:"status"`
	Version    string            `json:"version,omitempty"`
	Uptime     string            `json:"uptime"`
	Components []ComponentHealth `json:"components,omitempty"`
}

// Metrics contains basic application metrics
type Metrics struct {
	Uptime         time.Duration `json:"uptime"`
	UptimeSeconds  float64       `json:"uptime_seconds"`
	RequestsTotal  int64         `json:"requests_total"`
	RequestsActive int64         `json:"requests_active"`
	ErrorsTotal    int64         `json:"errors_total"`
	GoRoutines     int           `json:"goroutines"`
	MemoryAllocMB  float64       `json:"memory_alloc_mb"`
	MemorySysMB    float64       `json:"memory_sys_mb"`
	GCPauseMs      float64       `json:"gc_pause_ms"`
	NumCPU         int           `json:"num_cpu"`
}

// Checker provides health check functionality
type Checker struct {
	startTime      time.Time
	version        string
	db             *sql.DB
	mu             sync.RWMutex
	requestsTotal  int64
	requestsActive int64
	errorsTotal    int64
}

// NewChecker creates a new health checker
func NewChecker(version string, db *sql.DB) *Checker {
	return &Checker{
		startTime: time.Now(),
		version:   version,
		db:        db,
	}
}

// IncrementRequests increments the active request counter
func (c *Checker) IncrementRequests() {
	c.mu.Lock()
	c.requestsActive++
	c.requestsTotal++
	c.mu.Unlock()
}

// DecrementRequests decrements the active request counter
func (c *Checker) DecrementRequests() {
	c.mu.Lock()
	c.requestsActive--
	c.mu.Unlock()
}

// IncrementErrors increments the error counter
func (c *Checker) IncrementErrors() {
	c.mu.Lock()
	c.errorsTotal++
	c.mu.Unlock()
}

// Check performs a full health check
func (c *Checker) Check(ctx context.Context) HealthResponse {
	components := []ComponentHealth{}
	overallStatus := StatusHealthy

	// Check database
	dbHealth := c.checkDatabase(ctx)
	components = append(components, dbHealth)
	if dbHealth.Status == StatusUnhealthy {
		overallStatus = StatusUnhealthy
	} else if dbHealth.Status == StatusDegraded && overallStatus == StatusHealthy {
		overallStatus = StatusDegraded
	}

	return HealthResponse{
		Status:     overallStatus,
		Version:    c.version,
		Uptime:     time.Since(c.startTime).Round(time.Second).String(),
		Components: components,
	}
}

// checkDatabase checks database connectivity
func (c *Checker) checkDatabase(ctx context.Context) ComponentHealth {
	if c.db == nil {
		return ComponentHealth{
			Name:    "database",
			Status:  StatusUnhealthy,
			Message: "database not configured",
		}
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := c.db.PingContext(ctx)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return ComponentHealth{
			Name:    "database",
			Status:  StatusUnhealthy,
			Message: err.Error(),
			Latency: latency,
		}
	}

	status := StatusHealthy
	message := ""
	if latency > 100 {
		status = StatusDegraded
		message = "high latency"
	}

	return ComponentHealth{
		Name:    "database",
		Status:  status,
		Message: message,
		Latency: latency,
	}
}

// GetMetrics returns current application metrics
func (c *Checker) GetMetrics() Metrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.mu.RLock()
	requestsTotal := c.requestsTotal
	requestsActive := c.requestsActive
	errorsTotal := c.errorsTotal
	c.mu.RUnlock()

	uptime := time.Since(c.startTime)

	return Metrics{
		Uptime:         uptime,
		UptimeSeconds:  uptime.Seconds(),
		RequestsTotal:  requestsTotal,
		RequestsActive: requestsActive,
		ErrorsTotal:    errorsTotal,
		GoRoutines:     runtime.NumGoroutine(),
		MemoryAllocMB:  float64(m.Alloc) / 1024 / 1024,
		MemorySysMB:    float64(m.Sys) / 1024 / 1024,
		GCPauseMs:      float64(m.PauseTotalNs) / 1e6,
		NumCPU:         runtime.NumCPU(),
	}
}

// LivenessHandler returns a simple liveness probe handler
func (c *Checker) LivenessHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}

// ReadinessHandler returns a readiness probe handler
func (c *Checker) ReadinessHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		health := c.Check(ctx.Request.Context())

		status := http.StatusOK
		if health.Status == StatusUnhealthy {
			status = http.StatusServiceUnavailable
		}

		ctx.JSON(status, health)
	}
}

// HealthHandler returns a detailed health check handler
func (c *Checker) HealthHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		health := c.Check(ctx.Request.Context())

		status := http.StatusOK
		if health.Status == StatusUnhealthy {
			status = http.StatusServiceUnavailable
		}

		ctx.JSON(status, health)
	}
}

// MetricsHandler returns a metrics handler
func (c *Checker) MetricsHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		metrics := c.GetMetrics()
		ctx.JSON(http.StatusOK, metrics)
	}
}

// RequestCounterMiddleware is a middleware that tracks request counts
func (c *Checker) RequestCounterMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c.IncrementRequests()
		defer c.DecrementRequests()

		ctx.Next()

		// Track errors
		if ctx.Writer.Status() >= 500 {
			c.IncrementErrors()
		}
	}
}
