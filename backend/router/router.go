package router

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	gorillaWs "github.com/gorilla/websocket"
	"github.com/vektah/gqlparser/v2/ast"

	appConfig "github.com/arandu-ai/arandu/config"
	"github.com/arandu-ai/arandu/database"
	"github.com/arandu-ai/arandu/graph"
	"github.com/arandu-ai/arandu/logging"
	"github.com/arandu-ai/arandu/models"
	"github.com/arandu-ai/arandu/websocket"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	// Start cleanup goroutine
	go rl.cleanup()
	return rl
}

// Allow checks if the request should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Filter out old requests
	var validRequests []time.Time
	for _, t := range rl.requests[ip] {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	if len(validRequests) >= rl.limit {
		rl.requests[ip] = validRequests
		return false
	}

	rl.requests[ip] = append(validRequests, now)
	return true
}

// cleanup periodically removes old entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		windowStart := now.Add(-rl.window)
		for ip, times := range rl.requests {
			var valid []time.Time
			for _, t := range times {
				if t.After(windowStart) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = valid
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware creates a gin middleware for rate limiting
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			return
		}
		c.Next()
	}
}

// securityHeadersMiddleware adds security headers to responses
func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (adjust as needed)
		if appConfig.Config.ProductionMode {
			c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; connect-src 'self' wss: ws:")
		}

		c.Next()
	}
}

// allowedOrigins stores the parsed list of allowed CORS origins
var allowedOrigins []string

// isOriginAllowed checks if the given origin is in the allowed list
func isOriginAllowed(origin string) bool {
	// If wildcard is configured, allow all (development only)
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
	}
	return false
}

func New(db *database.Queries) *gin.Engine {
	// Initialize Gin router
	r := gin.Default()

	// Parse allowed origins from config
	originsConfig := appConfig.Config.CORSAllowedOrigins
	if originsConfig != "" {
		allowedOrigins = strings.Split(originsConfig, ",")
		// Trim whitespace from each origin
		for i, origin := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(origin)
		}
	}
	logging.Info("CORS configuration loaded", "allowed_origins", allowedOrigins)

	// Configure CORS middleware with secure settings
	corsConfig := cors.DefaultConfig()

	// Check if wildcard is allowed (development mode)
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		logging.Warn("CORS is configured to allow all origins - should only be used in development")
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowAllOrigins = false
		corsConfig.AllowOrigins = allowedOrigins
	}

	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 86400 // 24 hours preflight cache

	r.Use(cors.New(corsConfig))

	// Security: Rate limiting middleware
	rateLimiter := NewRateLimiter(appConfig.Config.RateLimitPerMinute, time.Minute)
	r.Use(RateLimitMiddleware(rateLimiter))
	logging.Info("Rate limiting enabled", "requests_per_minute", appConfig.Config.RateLimitPerMinute)

	// Security headers middleware
	r.Use(securityHeadersMiddleware())

	r.Use(static.Serve("/", static.LocalFile("./fe", true)))

	// GraphQL endpoint
	r.Any("/graphql", graphqlHandler(db))

	// GraphQL playground route
	r.GET("/playground", playgroundHandler())

	// WebSocket endpoint for Docker daemon
	r.GET("/terminal/:id", wsHandler(db))

	// Static file server
	r.Static("/browser", "./tmp/browser")

	r.NoRoute(func(c *gin.Context) {
		c.Redirect(301, "/")
	})

	return r
}

func graphqlHandler(db *database.Queries) gin.HandlerFunc {
	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in the resolver.go file
	h := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		Db: db,
	}}))

	h.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		res := next(ctx)
		if res == nil {
			return res
		}

		errMsg := res.Errors.Error()

		if errMsg != "" {
			logging.Error("GraphQL error", "error", errMsg)
		}

		return res
	})

	// We can't use the default error handler because it doesn't work with websockets
	// https://stackoverflow.com/a/75444816
	// So we add all the transports manually (see handler.NewDefaultServer in gqlgen for reference)
	h.AddTransport(transport.Options{})
	h.AddTransport(transport.GET{})
	h.AddTransport(transport.POST{})
	h.AddTransport(transport.MultipartForm{})

	h.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	// Security: Disable introspection in production
	if !appConfig.Config.DisableIntrospection {
		h.Use(extension.Introspection{})
	} else {
		logging.Info("GraphQL introspection is disabled for security")
	}

	h.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	// Add transport to support GraphQL subscriptions
	h.AddTransport(&transport.Websocket{
		Upgrader: gorillaWs.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					// Allow requests without Origin header (same-origin requests)
					return true
				}
				allowed := isOriginAllowed(origin)
				if !allowed {
					logging.Warn("WebSocket connection rejected", "origin", origin)
				}
				return allowed
			},
		},
		InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, *transport.InitPayload, error) {
			return ctx, &initPayload, nil
		},
	})

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func playgroundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		playground.Handler("GraphQL", "/graphql").ServeHTTP(c.Writer, c.Request)
	}
}

func wsHandler(db *database.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")

		// convert id to uint
		id, err := strconv.ParseUint(idParam, 10, 64)

		if err != nil {
			_ = c.AbortWithError(400, err)
			return
		}

		flow, err := db.ReadFlow(c, int64(id))

		if err != nil {
			_ = c.AbortWithError(404, err)
			return
		}

		if flow.Status.String != string(models.FlowInProgress) {
			_ = c.AbortWithError(404, fmt.Errorf("flow is not in progress"))
			return
		}

		if flow.ContainerStatus.String != string(models.ContainerRunning) {
			_ = c.AbortWithError(404, fmt.Errorf("container is not running"))
			return
		}

		websocket.HandleWebsocket(c)
	}
}
