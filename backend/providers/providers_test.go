package providers

import (
	"database/sql"
	"testing"

	"github.com/arandu-ai/arandu/database"
)

func TestTruncateTasks(t *testing.T) {
	// Create test tasks with long results
	tasks := []database.Task{
		{
			ID:      1,
			Results: sql.NullString{String: makeString(3000), Valid: true},
		},
		{
			ID:      2,
			Results: sql.NullString{String: makeString(3000), Valid: true},
		},
		{
			ID:      3,
			Results: sql.NullString{String: makeString(3000), Valid: true},
		},
		{
			ID:      4,
			Results: sql.NullString{String: makeString(3000), Valid: true},
		},
		{
			ID:      5,
			Results: sql.NullString{String: makeString(3000), Valid: true},
		},
		{
			ID:      6,
			Results: sql.NullString{String: makeString(3000), Valid: true},
		},
		{
			ID:      7,
			Results: sql.NullString{String: makeString(3000), Valid: true},
		},
		{
			ID:      8,
			Results: sql.NullString{String: makeString(3000), Valid: true},
		},
	}

	// Test truncation with maxLength 500
	truncated := truncateTasks(tasks, 500)

	// Verify we have same number of tasks
	if len(truncated) != len(tasks) {
		t.Errorf("Expected %d tasks, got %d", len(tasks), len(truncated))
	}

	// First 3 tasks should have results truncated to ~500
	for i := 0; i < 3; i++ {
		if len(truncated[i].Results.String) > 520 { // Allow some buffer for "... [truncated]"
			t.Errorf("Task %d should be truncated to ~500 chars, got %d", i, len(truncated[i].Results.String))
		}
	}

	// Middle tasks (3-4) should be truncated more aggressively (500/4 = 125)
	for i := 3; i < 5; i++ {
		if len(truncated[i].Results.String) > 145 { // Allow buffer
			t.Errorf("Middle task %d should be truncated to ~125 chars, got %d", i, len(truncated[i].Results.String))
		}
	}

	// Last 3 tasks should have results truncated to ~500
	for i := 5; i < 8; i++ {
		if len(truncated[i].Results.String) > 520 {
			t.Errorf("Task %d should be truncated to ~500 chars, got %d", i, len(truncated[i].Results.String))
		}
	}
}

func TestTruncateTasksSmallList(t *testing.T) {
	// Test with less than 6 tasks
	tasks := []database.Task{
		{
			ID:      1,
			Results: sql.NullString{String: makeString(3000), Valid: true},
		},
		{
			ID:      2,
			Results: sql.NullString{String: makeString(3000), Valid: true},
		},
		{
			ID:      3,
			Results: sql.NullString{String: makeString(3000), Valid: true},
		},
	}

	truncated := truncateTasks(tasks, 500)

	// All tasks should be truncated equally
	for i, task := range truncated {
		if len(task.Results.String) > 520 {
			t.Errorf("Task %d should be truncated to ~500 chars, got %d", i, len(task.Results.String))
		}
	}
}

func TestTruncateTasksPreservesShortResults(t *testing.T) {
	// Test that short results are not modified
	shortResult := "Short result"
	tasks := []database.Task{
		{
			ID:      1,
			Results: sql.NullString{String: shortResult, Valid: true},
		},
	}

	truncated := truncateTasks(tasks, 500)

	if truncated[0].Results.String != shortResult {
		t.Errorf("Short results should not be modified, got %s", truncated[0].Results.String)
	}
}

func TestDefaultAskTask(t *testing.T) {
	message := "Test error message"
	task := defaultAskTask(message)

	if task == nil {
		t.Fatal("defaultAskTask should not return nil")
	}

	if task.Type.String != "ask" {
		t.Errorf("Expected task type 'ask', got %s", task.Type.String)
	}

	expectedMessage := "Test error message. What should I do next?"
	if task.Message.String != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, task.Message.String)
	}

	if task.Args.String != "{}" {
		t.Errorf("Expected empty args {}, got %s", task.Args.String)
	}
}

func TestProviderFactory(t *testing.T) {
	tests := []struct {
		name         string
		providerType ProviderType
		wantErr      bool
	}{
		{
			name:         "unknown provider",
			providerType: "unknown",
			wantErr:      true,
		},
		// Note: We can't test actual providers without config
		// This test just verifies the factory handles unknown providers
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ProviderFactory(tt.providerType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProviderFactory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTextToTask(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantErr  bool
		wantType string
	}{
		{
			name:     "valid terminal task",
			text:     `{"tool": "terminal", "tool_input": {"input": "ls -la"}, "message": "Listing files"}`,
			wantErr:  false,
			wantType: "terminal",
		},
		{
			name:     "valid browser task",
			text:     `{"tool": "browser", "tool_input": {"url": "https://google.com", "action": "read"}, "message": "Reading page"}`,
			wantErr:  false,
			wantType: "browser",
		},
		{
			name:     "valid ask task",
			text:     `{"tool": "ask", "tool_input": {}, "message": "What should I do?"}`,
			wantErr:  false,
			wantType: "ask",
		},
		{
			name:     "valid done task",
			text:     `{"tool": "done", "tool_input": {}, "message": "Task completed"}`,
			wantErr:  false,
			wantType: "done",
		},
		{
			name:    "invalid JSON",
			text:    "not valid json",
			wantErr: true,
		},
		{
			name:    "empty tool",
			text:    `{"tool": "", "tool_input": {}, "message": "test"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := textToTask(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("textToTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && task.Type.String != tt.wantType {
				t.Errorf("textToTask() type = %v, want %v", task.Type.String, tt.wantType)
			}
		})
	}
}

// Helper function to create a string of specified length
func makeString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}
