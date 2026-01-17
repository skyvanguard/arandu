package executor

import (
	"database/sql"
	"testing"

	"github.com/arandu-ai/arandu/database"
	"github.com/arandu-ai/arandu/providers"
)

// TestUnmarshalTaskArgs tests the generic unmarshal function
func TestUnmarshalTaskArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		wantErr bool
	}{
		{
			name:    "valid terminal args",
			args:    `{"input":"ls -la"}`,
			wantErr: false,
		},
		{
			name:    "empty JSON object",
			args:    `{}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			args:    `{invalid}`,
			wantErr: true,
		},
		{
			name:    "empty string",
			args:    ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := database.Task{
				Args: sql.NullString{String: tt.args, Valid: true},
			}

			_, err := unmarshalTaskArgs[providers.TerminalArgs](task)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshalTaskArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnmarshalTaskArgs_TerminalArgs(t *testing.T) {
	task := database.Task{
		Args: sql.NullString{String: `{"input":"echo hello"}`, Valid: true},
	}

	args, err := unmarshalTaskArgs[providers.TerminalArgs](task)
	if err != nil {
		t.Fatalf("unmarshalTaskArgs() error = %v", err)
	}

	if args.Input != "echo hello" {
		t.Errorf("args.Input = %q, want %q", args.Input, "echo hello")
	}
}

func TestUnmarshalTaskArgs_BrowserArgs(t *testing.T) {
	task := database.Task{
		Args: sql.NullString{String: `{"url":"https://example.com","action":"read"}`, Valid: true},
	}

	args, err := unmarshalTaskArgs[providers.BrowserArgs](task)
	if err != nil {
		t.Fatalf("unmarshalTaskArgs() error = %v", err)
	}

	if args.Url != "https://example.com" {
		t.Errorf("args.Url = %q, want %q", args.Url, "https://example.com")
	}
	if args.Action != "read" {
		t.Errorf("args.Action = %q, want %q", args.Action, "read")
	}
}

func TestUnmarshalTaskArgs_CodeArgs(t *testing.T) {
	task := database.Task{
		Args: sql.NullString{String: `{"path":"/app/test.go","action":"read_file","content":""}`, Valid: true},
	}

	args, err := unmarshalTaskArgs[providers.CodeArgs](task)
	if err != nil {
		t.Fatalf("unmarshalTaskArgs() error = %v", err)
	}

	if args.Path != "/app/test.go" {
		t.Errorf("args.Path = %q, want %q", args.Path, "/app/test.go")
	}
	if args.Action != "read_file" {
		t.Errorf("args.Action = %q, want %q", args.Action, "read_file")
	}
}

func TestValidateBrowserSecurity(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid HTTPS URL",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "valid HTTP URL",
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "file protocol should fail",
			url:     "file:///etc/passwd",
			wantErr: true,
		},
		{
			name:    "javascript protocol should fail",
			url:     "javascript:alert(1)",
			wantErr: true,
		},
		{
			name:    "empty URL should fail",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBrowserSecurity(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBrowserSecurity(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCodeSecurity(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid path inside /app",
			path:    "/app/src/main.go",
			wantErr: false,
		},
		{
			name:    "path traversal should fail",
			path:    "/app/../etc/passwd",
			wantErr: true,
		},
		{
			name:    "absolute path outside /app should fail",
			path:    "/etc/passwd",
			wantErr: true,
		},
		{
			name:    "relative path traversal should fail",
			path:    "../../etc/passwd",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCodeSecurity(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCodeSecurity(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestBrowserActionFuncType(t *testing.T) {
	// Test that BrowserActionFunc type signature is correct
	// This is a compile-time check to ensure the type is properly defined
	var actionFunc BrowserActionFunc

	// Assign Content function to verify signature matches
	actionFunc = Content
	if actionFunc == nil {
		t.Error("Content function should match BrowserActionFunc signature")
	}

	// Assign URLs function to verify signature matches
	actionFunc = URLs
	if actionFunc == nil {
		t.Error("URLs function should match BrowserActionFunc signature")
	}
}

func TestBrowserActionTypes(t *testing.T) {
	// Verify that browser action constants are defined correctly
	tests := []struct {
		action   providers.BrowserAction
		expected string
	}{
		{providers.Read, "read"},
		{providers.Url, "url"},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			if string(tt.action) != tt.expected {
				t.Errorf("BrowserAction = %q, want %q", tt.action, tt.expected)
			}
		})
	}
}

func TestCodeActionTypes(t *testing.T) {
	// Verify that code action constants are defined correctly
	tests := []struct {
		action   providers.CodeAction
		expected string
	}{
		{providers.ReadFile, "read_file"},
		{providers.UpdateFile, "update_file"},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			if string(tt.action) != tt.expected {
				t.Errorf("CodeAction = %q, want %q", tt.action, tt.expected)
			}
		})
	}
}
