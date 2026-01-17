package security

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/arandu-ai/arandu/config"
)

// AllowedDockerImages is a whitelist of allowed Docker images
// Add more images as needed
var AllowedDockerImages = map[string]bool{
	"node:latest":      true,
	"node:20":          true,
	"node:18":          true,
	"python:latest":    true,
	"python:3.12":      true,
	"python:3.11":      true,
	"python:3.10":      true,
	"golang:latest":    true,
	"golang:1.22":      true,
	"golang:1.21":      true,
	"rust:latest":      true,
	"ruby:latest":      true,
	"php:latest":       true,
	"openjdk:latest":   true,
	"ubuntu:latest":    true,
	"debian:latest":    true,
	"alpine:latest":    true,
}

// BlockedURLPatterns contains patterns that should not be accessed
var BlockedURLPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)^file://`),                    // Local file access
	regexp.MustCompile(`(?i)^(http://)?localhost`),        // Localhost
	regexp.MustCompile(`(?i)^(http://)?127\.0\.0\.1`),     // Loopback
	regexp.MustCompile(`(?i)^(http://)?0\.0\.0\.0`),       // All interfaces
	regexp.MustCompile(`(?i)^(http://)?10\.\d+\.\d+\.\d+`),   // Private network 10.x
	regexp.MustCompile(`(?i)^(http://)?172\.(1[6-9]|2\d|3[01])\.\d+\.\d+`), // Private 172.16-31.x
	regexp.MustCompile(`(?i)^(http://)?192\.168\.\d+\.\d+`),  // Private 192.168.x
	regexp.MustCompile(`(?i)^(http://)?169\.254\.\d+\.\d+`),  // Link-local
	regexp.MustCompile(`(?i)metadata\.google\.internal`),     // GCP metadata
	regexp.MustCompile(`(?i)169\.254\.169\.254`),             // AWS/Azure metadata
}

// ValidatePath checks if a path is safe (no path traversal)
// It ensures the path doesn't escape the working directory
func ValidatePath(path string, workingDir string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Clean the path to resolve any .. or .
	cleanPath := filepath.Clean(path)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal detected: %s", path)
	}

	// Block access to sensitive paths FIRST (before workingDir check)
	sensitivePatterns := []string{
		"etc/passwd",
		"etc/shadow",
		"etc/hosts",
		"etc/ssh",
		".ssh",
		".env",
		".git/config",
		"proc/",
		"sys/",
		"dev/",
		"passwd",
		"shadow",
	}

	lowerPath := strings.ToLower(cleanPath)
	// Also check the original path for patterns
	lowerOriginal := strings.ToLower(path)

	for _, pattern := range sensitivePatterns {
		patternLower := strings.ToLower(pattern)
		if strings.Contains(lowerPath, patternLower) || strings.Contains(lowerOriginal, patternLower) {
			return fmt.Errorf("access to sensitive path blocked: %s", path)
		}
	}

	// If workingDir is provided, ensure the path stays within it
	if workingDir != "" {
		absWorkingDir, err := filepath.Abs(workingDir)
		if err != nil {
			return fmt.Errorf("invalid working directory: %w", err)
		}

		var absPath string
		if filepath.IsAbs(cleanPath) {
			absPath = cleanPath
		} else {
			absPath = filepath.Join(absWorkingDir, cleanPath)
		}

		// Ensure the resolved path is within the working directory
		if !strings.HasPrefix(absPath, absWorkingDir) {
			return fmt.Errorf("path escapes working directory: %s", path)
		}
	}

	return nil
}

// ValidateURL checks if a URL is safe to access
func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow http and https schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("only http and https URLs are allowed, got: %s", parsedURL.Scheme)
	}

	// Check against blocked patterns
	for _, pattern := range BlockedURLPatterns {
		if pattern.MatchString(rawURL) {
			return fmt.Errorf("URL matches blocked pattern: %s", rawURL)
		}
	}

	return nil
}

// ValidateDockerImage checks if a Docker image is in the whitelist
// If ALLOW_ANY_DOCKER_IMAGE env is set, all images are allowed (for development)
func ValidateDockerImage(image string) error {
	if image == "" {
		return fmt.Errorf("docker image cannot be empty")
	}

	// Allow any image in development mode
	if config.Config.AllowAnyDockerImage {
		return nil
	}

	// Check if the image is in the whitelist
	if AllowedDockerImages[image] {
		return nil
	}

	// Check for common image patterns (e.g., node:20-alpine)
	baseName := strings.Split(image, ":")[0]
	if AllowedDockerImages[baseName+":latest"] {
		return nil
	}

	return fmt.Errorf("docker image not in whitelist: %s. Set ALLOW_ANY_DOCKER_IMAGE=true to allow any image", image)
}

// SanitizeLogMessage removes sensitive data from log messages
func SanitizeLogMessage(message string) string {
	// Patterns for sensitive data
	patterns := map[string]*regexp.Regexp{
		"[REDACTED_API_KEY]":    regexp.MustCompile(`(?i)(api[_-]?key|apikey|api_secret)["\s:=]+["\']?[\w\-]{20,}["\']?`),
		"[REDACTED_TOKEN]":      regexp.MustCompile(`(?i)(bearer|token|auth)["\s:=]+["\']?[\w\-\.]{20,}["\']?`),
		"[REDACTED_PASSWORD]":   regexp.MustCompile(`(?i)(password|passwd|pwd|secret)["\s:=]+["\']?[^\s"\']{8,}["\']?`),
		"[REDACTED_CREDENTIAL]": regexp.MustCompile(`(?i)(credential|cred)["\s:=]+["\']?[^\s"\']{8,}["\']?`),
	}

	result := message
	for replacement, pattern := range patterns {
		result = pattern.ReplaceAllString(result, replacement)
	}

	return result
}

// IsProductionEnvironment checks if running in production
func IsProductionEnvironment() bool {
	// This can be enhanced to check for GIN_MODE=release or other env vars
	return false // Default to development for now
}
