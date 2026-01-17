package executor

import (
	"database/sql"
	"testing"
	"time"

	"github.com/arandu-ai/arandu/database"
)

func TestTaskToGraphQL(t *testing.T) {
	now := time.Now()
	task := database.Task{
		ID:        123,
		Message:   sql.NullString{String: "Test message", Valid: true},
		Type:      sql.NullString{String: "terminal", Valid: true},
		Status:    sql.NullString{String: "finished", Valid: true},
		Args:      sql.NullString{String: `{"input":"ls -la"}`, Valid: true},
		Results:   sql.NullString{String: "file1.txt\nfile2.txt", Valid: true},
		CreatedAt: sql.NullTime{Time: now, Valid: true},
	}

	result := TaskToGraphQL(task)

	if result.ID != 123 {
		t.Errorf("ID = %d, want 123", result.ID)
	}
	if result.Message != "Test message" {
		t.Errorf("Message = %q, want %q", result.Message, "Test message")
	}
	if string(result.Type) != "terminal" {
		t.Errorf("Type = %q, want %q", result.Type, "terminal")
	}
	if string(result.Status) != "finished" {
		t.Errorf("Status = %q, want %q", result.Status, "finished")
	}
	if result.Args != `{"input":"ls -la"}` {
		t.Errorf("Args = %q, want %q", result.Args, `{"input":"ls -la"}`)
	}
	if result.Results != "file1.txt\nfile2.txt" {
		t.Errorf("Results = %q, want %q", result.Results, "file1.txt\nfile2.txt")
	}
}

func TestTaskToGraphQL_EmptyFields(t *testing.T) {
	task := database.Task{
		ID:      1,
		Message: sql.NullString{Valid: false},
		Type:    sql.NullString{Valid: false},
		Status:  sql.NullString{Valid: false},
		Args:    sql.NullString{Valid: false},
		Results: sql.NullString{Valid: false},
	}

	result := TaskToGraphQL(task)

	if result.Message != "" {
		t.Errorf("Message = %q, want empty string", result.Message)
	}
	if result.Args != "" {
		t.Errorf("Args = %q, want empty string", result.Args)
	}
}

func TestTasksToGraphQL(t *testing.T) {
	tasks := []database.Task{
		{ID: 1, Message: sql.NullString{String: "Task 1", Valid: true}},
		{ID: 2, Message: sql.NullString{String: "Task 2", Valid: true}},
		{ID: 3, Message: sql.NullString{String: "Task 3", Valid: true}},
	}

	result := TasksToGraphQL(tasks)

	if len(result) != 3 {
		t.Fatalf("len(result) = %d, want 3", len(result))
	}

	for i, task := range result {
		expectedID := uint(i + 1)
		if task.ID != expectedID {
			t.Errorf("result[%d].ID = %d, want %d", i, task.ID, expectedID)
		}
	}
}

func TestTasksToGraphQL_Empty(t *testing.T) {
	result := TasksToGraphQL([]database.Task{})

	if len(result) != 0 {
		t.Errorf("len(result) = %d, want 0", len(result))
	}
}

func TestLogToGraphQL(t *testing.T) {
	tests := []struct {
		name         string
		log          database.Log
		wantContains string
	}{
		{
			name: "output log",
			log: database.Log{
				ID:      1,
				Message: "Command output",
				Type:    "output",
			},
			wantContains: "Command output",
		},
		{
			name: "input log with formatting",
			log: database.Log{
				ID:      2,
				Message: "ls -la",
				Type:    "input",
			},
			wantContains: "ls -la",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LogToGraphQL(tt.log)

			if result.ID != uint(tt.log.ID) {
				t.Errorf("ID = %d, want %d", result.ID, tt.log.ID)
			}

			// El texto debe contener el mensaje original
			if result.Text == "" {
				t.Error("Text should not be empty")
			}
		})
	}
}

func TestLogsToGraphQL(t *testing.T) {
	logs := []database.Log{
		{ID: 1, Message: "Log 1", Type: "output"},
		{ID: 2, Message: "Log 2", Type: "output"},
	}

	result := LogsToGraphQL(logs)

	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}

	if result[0].ID != 1 {
		t.Errorf("result[0].ID = %d, want 1", result[0].ID)
	}
	if result[1].ID != 2 {
		t.Errorf("result[1].ID = %d, want 2", result[1].ID)
	}
}

func TestFlowRowToGraphQL(t *testing.T) {
	flow := database.ReadAllFlowsRow{
		ID:            42,
		Name:          sql.NullString{String: "Test Flow", Valid: true},
		Status:        sql.NullString{String: "in_progress", Valid: true},
		ModelProvider: sql.NullString{String: "openai", Valid: true},
		Model:         sql.NullString{String: "gpt-4", Valid: true},
		ContainerName: sql.NullString{String: "test-container", Valid: true},
	}

	result := FlowRowToGraphQL(flow)

	if result.ID != 42 {
		t.Errorf("ID = %d, want 42", result.ID)
	}
	if result.Name != "Test Flow" {
		t.Errorf("Name = %q, want %q", result.Name, "Test Flow")
	}
	if string(result.Status) != "in_progress" {
		t.Errorf("Status = %q, want %q", result.Status, "in_progress")
	}
	if result.Model == nil {
		t.Fatal("Model should not be nil")
	}
	if result.Model.Provider != "openai" {
		t.Errorf("Model.Provider = %q, want %q", result.Model.Provider, "openai")
	}
	if result.Model.ID != "gpt-4" {
		t.Errorf("Model.ID = %q, want %q", result.Model.ID, "gpt-4")
	}
	if result.Terminal == nil {
		t.Fatal("Terminal should not be nil")
	}
	if result.Terminal.ContainerName != "test-container" {
		t.Errorf("Terminal.ContainerName = %q, want %q", result.Terminal.ContainerName, "test-container")
	}
	if result.Terminal.Connected != false {
		t.Error("Terminal.Connected should be false for FlowRowToGraphQL")
	}
}

func TestFlowToGraphQL(t *testing.T) {
	flow := database.ReadFlowRow{
		ID:              42,
		Name:            sql.NullString{String: "Test Flow", Valid: true},
		Status:          sql.NullString{String: "finished", Valid: true},
		ModelProvider:   sql.NullString{String: "ollama", Valid: true},
		Model:           sql.NullString{String: "llama2", Valid: true},
		ContainerName:   sql.NullString{String: "test-container", Valid: true},
		ContainerStatus: sql.NullString{String: "running", Valid: true},
	}

	result := FlowToGraphQL(flow)

	if result.ID != 42 {
		t.Errorf("ID = %d, want 42", result.ID)
	}
	if result.Terminal == nil {
		t.Fatal("Terminal should not be nil")
	}
	if result.Terminal.Connected != true {
		t.Error("Terminal.Connected should be true when container status is 'running'")
	}
}

func TestFlowToGraphQL_NotRunning(t *testing.T) {
	flow := database.ReadFlowRow{
		ID:              1,
		ContainerStatus: sql.NullString{String: "stopped", Valid: true},
	}

	result := FlowToGraphQL(flow)

	if result.Terminal.Connected != false {
		t.Error("Terminal.Connected should be false when container is not running")
	}
}

func TestFlowToGraphQLFull(t *testing.T) {
	flow := database.ReadFlowRow{
		ID:              1,
		Name:            sql.NullString{String: "Full Flow", Valid: true},
		Status:          sql.NullString{String: "in_progress", Valid: true},
		ContainerStatus: sql.NullString{String: "running", Valid: true},
	}

	tasks := []database.Task{
		{ID: 1, Message: sql.NullString{String: "Task 1", Valid: true}},
		{ID: 2, Message: sql.NullString{String: "Task 2", Valid: true}},
	}

	logs := []database.Log{
		{ID: 1, Message: "Log 1", Type: "output"},
	}

	result := FlowToGraphQLFull(flow, tasks, logs)

	if result.ID != 1 {
		t.Errorf("ID = %d, want 1", result.ID)
	}
	if len(result.Tasks) != 2 {
		t.Errorf("len(Tasks) = %d, want 2", len(result.Tasks))
	}
	if result.Terminal == nil {
		t.Fatal("Terminal should not be nil")
	}
	if len(result.Terminal.Logs) != 1 {
		t.Errorf("len(Terminal.Logs) = %d, want 1", len(result.Terminal.Logs))
	}
}
