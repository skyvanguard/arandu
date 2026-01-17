package executor

import (
	"github.com/arandu-ai/arandu/database"
	gmodel "github.com/arandu-ai/arandu/graph/model"
	"github.com/arandu-ai/arandu/websocket"
)

// TaskToGraphQL convierte una tarea de database a modelo GraphQL
func TaskToGraphQL(task database.Task) *gmodel.Task {
	return &gmodel.Task{
		ID:        uint(task.ID),
		Message:   task.Message.String,
		Type:      gmodel.TaskType(task.Type.String),
		CreatedAt: task.CreatedAt.Time,
		Status:    gmodel.TaskStatus(task.Status.String),
		Args:      task.Args.String,
		Results:   task.Results.String,
	}
}

// TasksToGraphQL convierte una lista de tareas de database a modelos GraphQL
func TasksToGraphQL(tasks []database.Task) []*gmodel.Task {
	gTasks := make([]*gmodel.Task, len(tasks))
	for i, task := range tasks {
		gTasks[i] = TaskToGraphQL(task)
	}
	return gTasks
}

// LogToGraphQL convierte un log de database a modelo GraphQL
func LogToGraphQL(log database.Log) *gmodel.Log {
	text := log.Message
	if log.Type == "input" {
		text = websocket.FormatTerminalInput(log.Message)
	}
	return &gmodel.Log{
		ID:   uint(log.ID),
		Text: text,
	}
}

// LogsToGraphQL convierte una lista de logs de database a modelos GraphQL
func LogsToGraphQL(logs []database.Log) []*gmodel.Log {
	gLogs := make([]*gmodel.Log, len(logs))
	for i, log := range logs {
		gLogs[i] = LogToGraphQL(log)
	}
	return gLogs
}

// FlowRowToGraphQL convierte un ReadAllFlowsRow a modelo GraphQL básico
// Usado para listados de flows
func FlowRowToGraphQL(flow database.ReadAllFlowsRow) *gmodel.Flow {
	return &gmodel.Flow{
		ID:     uint(flow.ID),
		Name:   flow.Name.String,
		Status: gmodel.FlowStatus(flow.Status.String),
		Model: &gmodel.Model{
			Provider: flow.ModelProvider.String,
			ID:       flow.Model.String,
		},
		Terminal: &gmodel.Terminal{
			ContainerName: flow.ContainerName.String,
			Connected:     false, // En listados no tenemos el estado del container
		},
		Browser: &gmodel.Browser{
			URL:           "",
			ScreenshotURL: "",
		},
	}
}

// FlowToGraphQL convierte un ReadFlowRow a modelo GraphQL con detalles de container
func FlowToGraphQL(flow database.ReadFlowRow) *gmodel.Flow {
	return &gmodel.Flow{
		ID:     uint(flow.ID),
		Name:   flow.Name.String,
		Status: gmodel.FlowStatus(flow.Status.String),
		Model: &gmodel.Model{
			Provider: flow.ModelProvider.String,
			ID:       flow.Model.String,
		},
		Terminal: &gmodel.Terminal{
			ContainerName: flow.ContainerName.String,
			Connected:     flow.ContainerStatus.String == "running",
		},
		Browser: &gmodel.Browser{
			URL:           "",
			ScreenshotURL: "",
		},
	}
}

// FlowToGraphQLFull convierte un ReadFlowRow con tasks y logs a modelo GraphQL completo
func FlowToGraphQLFull(flow database.ReadFlowRow, tasks []database.Task, logs []database.Log) *gmodel.Flow {
	gFlow := FlowToGraphQL(flow)
	gFlow.Tasks = TasksToGraphQL(tasks)
	gFlow.Terminal.Logs = LogsToGraphQL(logs)
	return gFlow
}
