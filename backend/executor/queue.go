package executor

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/arandu-ai/arandu/database"
	"github.com/arandu-ai/arandu/graph/subscriptions"
	"github.com/arandu-ai/arandu/logging"
	"github.com/arandu-ai/arandu/models"
	"github.com/arandu-ai/arandu/providers"
)

// Constantes de configuración
const (
	// QueueBufferSize es el tamaño del buffer de la cola de tareas
	QueueBufferSize = 1000
	// MaxResultsLength es el máximo de caracteres en resultados de tareas
	MaxResultsLength = 4000
	// DBTimeout es el timeout por defecto para operaciones de base de datos
	DBTimeout = 30 * time.Second
	// LLMTimeout es el timeout para operaciones de LLM
	LLMTimeout = 60 * time.Second
)

// TaskHandler define cómo procesar un tipo de tarea
type TaskHandler struct {
	// Process ejecuta la lógica de la tarea
	Process func(provider providers.Provider, db *database.Queries, task database.Task) error
	// NeedsNextTask indica si después de procesar se debe obtener la siguiente tarea
	NeedsNextTask bool
}

// taskHandlers mapea tipos de tarea a sus handlers
var taskHandlers = map[string]TaskHandler{
	"input": {
		Process:       processInputTask,
		NeedsNextTask: true,
	},
	"ask": {
		Process:       func(_ providers.Provider, db *database.Queries, t database.Task) error { return processAskTask(db, t) },
		NeedsNextTask: false,
	},
	"terminal": {
		Process: func(_ providers.Provider, db *database.Queries, t database.Task) error {
			return processTerminalTask(db, t)
		},
		NeedsNextTask: true,
	},
	"code": {
		Process:       func(_ providers.Provider, db *database.Queries, t database.Task) error { return processCodeTask(db, t) },
		NeedsNextTask: true,
	},
	"done": {
		Process:       func(_ providers.Provider, db *database.Queries, t database.Task) error { return processDoneTask(db, t) },
		NeedsNextTask: false,
	},
	"browser": {
		Process: func(_ providers.Provider, db *database.Queries, t database.Task) error {
			return processBrowserTask(db, t)
		},
		NeedsNextTask: true,
	},
}

// QueueManager maneja las colas de tareas de forma thread-safe
type QueueManager struct {
	mu           sync.RWMutex
	queues       map[int64]chan database.Task
	stopChannels map[int64]chan struct{}
}

var queueManager = &QueueManager{
	queues:       make(map[int64]chan database.Task),
	stopChannels: make(map[int64]chan struct{}),
}

// AddQueue crea una nueva cola para un flow si no existe
func AddQueue(flowId int64, db *database.Queries) {
	queueManager.mu.Lock()
	defer queueManager.mu.Unlock()

	if _, ok := queueManager.queues[flowId]; !ok {
		queueManager.queues[flowId] = make(chan database.Task, QueueBufferSize)
		queueManager.stopChannels[flowId] = make(chan struct{})
		go processQueue(flowId, db)
	}
}

// AddCommand agrega una tarea a la cola del flow
func AddCommand(flowId int64, task database.Task) {
	queueManager.mu.RLock()
	q, ok := queueManager.queues[flowId]
	queueManager.mu.RUnlock()

	if ok && q != nil {
		select {
		case q <- task:
			logging.Debug("Command added to queue", "task_id", task.ID, "flow_id", flowId)
		default:
			logging.Warn("Queue full, command dropped", "flow_id", flowId, "task_id", task.ID)
		}
	}
}

// CleanQueue limpia y cierra la cola del flow
func CleanQueue(flowId int64) {
	queueManager.mu.Lock()
	defer queueManager.mu.Unlock()

	if stop, ok := queueManager.stopChannels[flowId]; ok {
		close(stop)
		delete(queueManager.stopChannels, flowId)
	}

	if q, ok := queueManager.queues[flowId]; ok {
		close(q)
		delete(queueManager.queues, flowId)
	}

	logging.Debug("Queue cleaned", "flow_id", flowId)
}

// getQueue obtiene la cola de forma segura
func getQueue(flowId int64) (chan database.Task, bool) {
	queueManager.mu.RLock()
	defer queueManager.mu.RUnlock()
	q, ok := queueManager.queues[flowId]
	return q, ok
}

// getStopChannel obtiene el canal de stop de forma segura
func getStopChannel(flowId int64) (chan struct{}, bool) {
	queueManager.mu.RLock()
	defer queueManager.mu.RUnlock()
	stop, ok := queueManager.stopChannels[flowId]
	return stop, ok
}

// initializeProvider crea el provider para un flow
func initializeProvider(flowId int64, db *database.Queries) (providers.Provider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DBTimeout)
	defer cancel()

	flow, err := db.ReadFlow(ctx, flowId)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow: %w", err)
	}

	provider, err := providers.ProviderFactory(providers.ProviderType(flow.ModelProvider.String))
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	logging.Info("Provider initialized",
		"provider", provider.Name(),
		"model", flow.Model.String,
		"flow_id", flowId,
	)

	return provider, nil
}

// runQueueLoop ejecuta el loop principal de procesamiento de tareas
func runQueueLoop(flowId int64, provider providers.Provider, db *database.Queries) {
	queue, queueOk := getQueue(flowId)
	stopChan, stopOk := getStopChannel(flowId)

	if !queueOk || !stopOk {
		logging.Error("Queue or stop channel not found", "flow_id", flowId)
		return
	}

	for {
		select {
		case <-stopChan:
			logging.Info("Stopping task processor", "flow_id", flowId)
			return
		case task, ok := <-queue:
			if !ok {
				logging.Debug("Queue closed", "flow_id", flowId)
				return
			}
			processTask(flowId, task, provider, db)
		}
	}
}

// processQueue procesa las tareas de la cola
func processQueue(flowId int64, db *database.Queries) {
	logging.Info("Starting task processor", "flow_id", flowId)

	provider, err := initializeProvider(flowId, db)
	if err != nil {
		logging.Error("Failed to initialize provider", "flow_id", flowId, "error", err.Error())
		CleanQueue(flowId)
		return
	}

	runQueueLoop(flowId, provider, db)
}

// processTask procesa una tarea individual usando el mapa de handlers
func processTask(flowId int64, task database.Task, provider providers.Provider, db *database.Queries) {
	start := time.Now()
	logging.Debug("Processing task", "task_id", task.ID, "type", task.Type.String)

	// Broadcast task added
	subscriptions.BroadcastTaskAdded(task.FlowID.Int64, TaskToGraphQL(task))

	// Buscar handler para este tipo de tarea
	handler, ok := taskHandlers[task.Type.String]
	if !ok {
		logging.Warn("Unknown task type", "type", task.Type.String)
		return
	}

	// Ejecutar el handler
	err := handler.Process(provider, db, task)
	if err != nil {
		logging.Error("Failed to process task",
			"task_id", task.ID,
			"type", task.Type.String,
			"error", err.Error(),
		)
		updateTaskError(db, task.ID, err)
		return
	}

	logging.LogTask(task.ID, task.Type.String, "completed", time.Since(start))

	// Obtener siguiente tarea si el handler lo requiere
	if handler.NeedsNextTask {
		nextTask, err := getNextTask(provider, db, task.FlowID.Int64)
		if err != nil {
			logging.Error("Failed to get next task", "flow_id", task.FlowID.Int64, "error", err.Error())
			return
		}
		AddCommand(flowId, *nextTask)
	}
}

// updateTaskError actualiza el estado de una tarea a error
func updateTaskError(db *database.Queries, taskID int64, taskErr error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.UpdateTaskStatus(ctx, database.UpdateTaskStatusParams{
		ID:     taskID,
		Status: database.StringToNullString("error"),
	})
	if err != nil {
		logging.Error("Failed to update task status to error", "task_id", taskID, "error", err.Error())
	}
}

func getNextTask(provider providers.Provider, db *database.Queries, flowId int64) (*database.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), LLMTimeout)
	defer cancel()

	flow, err := db.ReadFlow(ctx, flowId)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow: %w", err)
	}

	tasks, err := db.ReadTasksByFlowId(ctx, sql.NullInt64{
		Int64: flowId,
		Valid: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by flow id: %w", err)
	}

	// Truncar resultados largos
	for i, task := range tasks {
		if len(task.Results.String) > MaxResultsLength {
			results := task.Results.String[len(task.Results.String)-MaxResultsLength:]
			tasks[i].Results = database.StringToNullString(results)
		}
	}

	c := provider.NextTask(providers.NextTaskOptions{
		Tasks:       tasks,
		DockerImage: flow.ContainerImage.String,
	})

	lastTask := tasks[len(tasks)-1]

	_, err = db.UpdateTaskToolCallId(ctx, database.UpdateTaskToolCallIdParams{
		ToolCallID: c.ToolCallID,
		ID:         lastTask.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update task tool call id: %w", err)
	}

	nextTask, err := db.CreateTask(ctx, database.CreateTaskParams{
		Args:       c.Args,
		Message:    c.Message,
		Type:       c.Type,
		Status:     database.StringToNullString(models.TaskInProgress),
		FlowID:     sql.NullInt64{Int64: flowId, Valid: true},
		ToolCallID: c.ToolCallID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save command: %w", err)
	}

	return &nextTask, nil
}
