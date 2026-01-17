package security

import (
	"testing"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		workingDir string
		wantErr    bool
	}{
		{
			name:       "valid simple path",
			path:       "file.txt",
			workingDir: "/app",
			wantErr:    false,
		},
		{
			name:       "valid nested path",
			path:       "src/main.go",
			workingDir: "/app",
			wantErr:    false,
		},
		{
			name:       "path traversal attack",
			path:       "../../../etc/passwd",
			workingDir: "/app",
			wantErr:    true,
		},
		{
			name:       "path traversal with dots",
			path:       "src/../../../etc/passwd",
			workingDir: "/app",
			wantErr:    true,
		},
		{
			name:       "sensitive path /etc/passwd with workdir",
			path:       "/etc/passwd",
			workingDir: "/app",
			wantErr:    true,
		},
		{
			name:       "sensitive path /etc/shadow with workdir",
			path:       "/etc/shadow",
			workingDir: "/app",
			wantErr:    true,
		},
		{
			name:       "sensitive path .ssh with workdir",
			path:       "/root/.ssh/id_rsa",
			workingDir: "/app",
			wantErr:    true,
		},
		{
			name:       "sensitive path .env with workdir",
			path:       "/.env",
			workingDir: "/app",
			wantErr:    true,
		},
		{
			name:       "empty path",
			path:       "",
			workingDir: "/app",
			wantErr:    true,
		},
		{
			name:       "valid absolute path within workdir",
			path:       "/app/src/main.go",
			workingDir: "/app",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path, tt.workingDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid https URL",
			url:     "https://google.com",
			wantErr: false,
		},
		{
			name:    "valid http URL",
			url:     "http://example.com/path",
			wantErr: false,
		},
		{
			name:    "localhost blocked",
			url:     "http://localhost:8080",
			wantErr: true,
		},
		{
			name:    "127.0.0.1 blocked",
			url:     "http://127.0.0.1:3000",
			wantErr: true,
		},
		{
			name:    "private IP 10.x blocked",
			url:     "http://10.0.0.1/admin",
			wantErr: true,
		},
		{
			name:    "private IP 192.168.x blocked",
			url:     "http://192.168.1.1",
			wantErr: true,
		},
		{
			name:    "private IP 172.16.x blocked",
			url:     "http://172.16.0.1",
			wantErr: true,
		},
		{
			name:    "AWS metadata blocked",
			url:     "http://169.254.169.254/latest/meta-data",
			wantErr: true,
		},
		{
			name:    "GCP metadata blocked",
			url:     "http://metadata.google.internal/computeMetadata",
			wantErr: true,
		},
		{
			name:    "file:// scheme blocked",
			url:     "file:///etc/passwd",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "ftp scheme blocked",
			url:     "ftp://files.example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDockerImage(t *testing.T) {
	tests := []struct {
		name    string
		image   string
		wantErr bool
	}{
		{
			name:    "allowed node:latest",
			image:   "node:latest",
			wantErr: false,
		},
		{
			name:    "allowed python:3.12",
			image:   "python:3.12",
			wantErr: false,
		},
		{
			name:    "allowed golang:latest",
			image:   "golang:latest",
			wantErr: false,
		},
		{
			name:    "node with tag variant",
			image:   "node:20-alpine",
			wantErr: false, // Should work because node:latest is in whitelist
		},
		{
			name:    "empty image",
			image:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDockerImage(tt.image)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDockerImage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeLogMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		contains string
		notContains string
	}{
		{
			name:        "sanitize API key",
			message:     `{"api_key": "sk-1234567890abcdefghijklmnop"}`,
			contains:    "[REDACTED_API_KEY]",
			notContains: "sk-1234567890",
		},
		{
			name:        "sanitize bearer token",
			message:     `Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9`,
			contains:    "[REDACTED_TOKEN]",
			notContains: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		},
		{
			name:        "sanitize password",
			message:     `{"password": "mysecretpassword123"}`,
			contains:    "[REDACTED_PASSWORD]",
			notContains: "mysecretpassword123",
		},
		{
			name:        "normal message unchanged",
			message:     "This is a normal log message",
			contains:    "This is a normal log message",
			notContains: "REDACTED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeLogMessage(tt.message)
			if tt.contains != "" && !containsString(result, tt.contains) {
				t.Errorf("SanitizeLogMessage() should contain %q, got %q", tt.contains, result)
			}
			if tt.notContains != "" && containsString(result, tt.notContains) {
				t.Errorf("SanitizeLogMessage() should not contain %q, got %q", tt.notContains, result)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
