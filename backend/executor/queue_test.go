package executor

import (
	"testing"
	"time"
)

func TestQueueConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"QueueBufferSize", QueueBufferSize, 1000},
		{"MaxResultsLength", MaxResultsLength, 4000},
		{"DBTimeout", DBTimeout, 30 * time.Second},
		{"LLMTimeout", LLMTimeout, 60 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestTaskHandlersRegistration(t *testing.T) {
	// Verificar que todos los tipos de tarea esperados están registrados
	expectedTypes := []string{"input", "ask", "terminal", "code", "done", "browser"}

	for _, taskType := range expectedTypes {
		t.Run(taskType, func(t *testing.T) {
			handler, ok := taskHandlers[taskType]
			if !ok {
				t.Errorf("taskHandlers missing handler for %q", taskType)
				return
			}
			if handler.Process == nil {
				t.Errorf("taskHandlers[%q].Process is nil", taskType)
			}
		})
	}
}

func TestTaskHandlersNeedsNextTask(t *testing.T) {
	// Verificar que NeedsNextTask está configurado correctamente
	tests := []struct {
		taskType      string
		needsNextTask bool
	}{
		{"input", true},
		{"ask", false},
		{"terminal", true},
		{"code", true},
		{"done", false},
		{"browser", true},
	}

	for _, tt := range tests {
		t.Run(tt.taskType, func(t *testing.T) {
			handler, ok := taskHandlers[tt.taskType]
			if !ok {
				t.Fatalf("handler for %q not found", tt.taskType)
			}
			if handler.NeedsNextTask != tt.needsNextTask {
				t.Errorf("NeedsNextTask = %v, want %v", handler.NeedsNextTask, tt.needsNextTask)
			}
		})
	}
}

func TestTaskHandlersCount(t *testing.T) {
	// Verificar que tenemos exactamente 6 handlers
	expected := 6
	if len(taskHandlers) != expected {
		t.Errorf("len(taskHandlers) = %d, want %d", len(taskHandlers), expected)
	}
}

func TestUnknownTaskType(t *testing.T) {
	// Verificar que tipos desconocidos no están en el mapa
	unknownTypes := []string{"unknown", "invalid", "", "INPUT", "TERMINAL"}

	for _, taskType := range unknownTypes {
		t.Run(taskType, func(t *testing.T) {
			if _, ok := taskHandlers[taskType]; ok {
				t.Errorf("taskHandlers should not have handler for %q", taskType)
			}
		})
	}
}

func TestQueueManagerInitialization(t *testing.T) {
	// Verificar que queueManager está inicializado correctamente
	if queueManager == nil {
		t.Fatal("queueManager should not be nil")
	}
	if queueManager.queues == nil {
		t.Error("queueManager.queues should not be nil")
	}
	if queueManager.stopChannels == nil {
		t.Error("queueManager.stopChannels should not be nil")
	}
}

func TestGetQueueNotExists(t *testing.T) {
	// Test que getQueue retorna false para un flow que no existe
	_, ok := getQueue(999999)
	if ok {
		t.Error("getQueue should return false for non-existent flow")
	}
}

func TestGetStopChannelNotExists(t *testing.T) {
	// Test que getStopChannel retorna false para un flow que no existe
	_, ok := getStopChannel(999999)
	if ok {
		t.Error("getStopChannel should return false for non-existent flow")
	}
}
