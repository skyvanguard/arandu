package config

import (
	"os"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	// Save original env vars
	originalPort := os.Getenv("PORT")
	originalDBURL := os.Getenv("DATABASE_URL")

	// Clear env vars for testing defaults
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")

	// Restore after test
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		}
		if originalDBURL != "" {
			os.Setenv("DATABASE_URL", originalDBURL)
		}
	}()

	// Re-initialize config
	Init()

	// Test defaults
	if Config.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", Config.Port)
	}

	if Config.DatabaseURL != "database.db" {
		t.Errorf("Expected default database URL 'database.db', got %s", Config.DatabaseURL)
	}
}

func TestConfigFromEnv(t *testing.T) {
	// Set test env vars
	os.Setenv("PORT", "3000")
	os.Setenv("DATABASE_URL", "test.db")
	os.Setenv("OLLAMA_MODEL", "llama2:7b")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("OLLAMA_MODEL")
	}()

	// Re-initialize config
	Init()

	if Config.Port != 3000 {
		t.Errorf("Expected port 3000 from env, got %d", Config.Port)
	}

	if Config.DatabaseURL != "test.db" {
		t.Errorf("Expected database URL 'test.db' from env, got %s", Config.DatabaseURL)
	}

	if Config.OllamaModel != "llama2:7b" {
		t.Errorf("Expected Ollama model 'llama2:7b' from env, got %s", Config.OllamaModel)
	}
}

func TestOllamaServerURLDefault(t *testing.T) {
	os.Unsetenv("OLLAMA_SERVER_URL")
	defer os.Unsetenv("OLLAMA_SERVER_URL")

	Init()

	expected := "http://localhost:11434"
	if Config.OllamaServerURL != expected {
		t.Errorf("Expected default Ollama server URL '%s', got %s", expected, Config.OllamaServerURL)
	}
}
