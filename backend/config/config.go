package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type config struct {
	// General
	DatabaseURL string `env:"DATABASE_URL" envDefault:"database.db"`
	Port        int    `env:"PORT" envDefault:"8080"`

	// OpenAI (or OpenAI-compatible API like LM Studio, LocalAI, vLLM, etc.)
	OpenAIKey       string `env:"OPEN_AI_KEY"`
	OpenAIModel     string `env:"OPEN_AI_MODEL" envDefault:"gpt-4o"`
	OpenAIServerURL string `env:"OPEN_AI_SERVER_URL" envDefault:"https://api.openai.com/v1"`

	// Ollama - Local LLM server (https://ollama.ai)
	OllamaModel     string `env:"OLLAMA_MODEL"`
	OllamaServerURL string `env:"OLLAMA_SERVER_URL" envDefault:"http://localhost:11434"`

	// LM Studio - Local LLM with OpenAI-compatible API (https://lmstudio.ai)
	LMStudioModel     string `env:"LMSTUDIO_MODEL"`
	LMStudioServerURL string `env:"LMSTUDIO_SERVER_URL" envDefault:"http://localhost:1234/v1"`

	// LocalAI - Local OpenAI-compatible API (https://localai.io)
	LocalAIModel     string `env:"LOCALAI_MODEL"`
	LocalAIServerURL string `env:"LOCALAI_SERVER_URL" envDefault:"http://localhost:8080/v1"`

	// Generic OpenAI-Compatible provider (for any server with OpenAI API)
	// Use this for: vLLM, text-generation-webui, llama.cpp server, etc.
	OpenAICompatibleModel     string `env:"OPENAI_COMPATIBLE_MODEL"`
	OpenAICompatibleServerURL string `env:"OPENAI_COMPATIBLE_SERVER_URL"`
	OpenAICompatibleAPIKey    string `env:"OPENAI_COMPATIBLE_API_KEY" envDefault:"not-needed"`

	// Browser (Bug fix #65: configurable Chrome debugging URL)
	ChromeDebugURL string `env:"CHROME_DEBUG_URL" envDefault:""`

	// Security: CORS allowed origins (comma-separated list)
	// Use "*" for development only, specify exact origins in production
	// Example: "http://localhost:3000,https://myapp.com"
	CORSAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:3000,http://localhost:5173,http://127.0.0.1:3000,http://127.0.0.1:5173"`

	// Security: Production mode settings
	// Set to "true" in production to enable security hardening
	ProductionMode bool `env:"PRODUCTION_MODE" envDefault:"false"`

	// Security: Disable GraphQL introspection in production
	DisableIntrospection bool `env:"DISABLE_INTROSPECTION" envDefault:"false"`

	// Security: Rate limiting (requests per minute per IP)
	RateLimitPerMinute int `env:"RATE_LIMIT_PER_MINUTE" envDefault:"60"`

	// Security: Allow any Docker image (development only)
	AllowAnyDockerImage bool `env:"ALLOW_ANY_DOCKER_IMAGE" envDefault:"false"`

	// Server: Base URL for external access (used for screenshot URLs, etc.)
	BaseURL string `env:"BASE_URL" envDefault:""`

	// Authentication: API key for securing the API
	// If set, all requests must include X-API-Key header
	APIKey string `env:"API_KEY" envDefault:""`

	// Authentication: Enable API key requirement
	RequireAPIKey bool `env:"REQUIRE_API_KEY" envDefault:"false"`

	// Logging: Log level (debug, info, warn, error)
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	// Logging: Log format (text, json)
	LogFormat string `env:"LOG_FORMAT" envDefault:"text"`
}

var Config config

func Init() {
	_ = godotenv.Load() // Ignore error - .env file is optional

	if err := env.ParseWithOptions(&Config, env.Options{
		RequiredIfNoDef: false,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Unable to parse config: %v\n", err)
		os.Exit(1)
	}
}
