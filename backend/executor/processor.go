package executor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/arandu-ai/arandu/config"
	"github.com/arandu-ai/arandu/database"
	gmodel "github.com/arandu-ai/arandu/graph/model"
	"github.com/arandu-ai/arandu/graph/subscriptions"
	"github.com/arandu-ai/arandu/logging"
	"github.com/arandu-ai/arandu/models"
	"github.com/arandu-ai/arandu/providers"
	"github.com/arandu-ai/arandu/security"
	"github.com/docker/docker/api/types/container"
)

// SummaryWordCount es el número de palabras para el resumen del flujo
const SummaryWordCount = 10

// BrowserActionFunc type for browser action functions (Content, URLs)
type BrowserActionFunc func(url string) (content string, screenshot string, err error)

// unmarshalTaskArgs deserializa los argumentos de una tarea a un tipo específico
func unmarshalTaskArgs[T any](task database.Task) (T, error) {
	var args T
	if err := json.Unmarshal([]byte(task.Args.String), &args); err != nil {
		return args, fmt.Errorf("failed to unmarshal args: %w", err)
	}
	return args, nil
}

// validateBrowserSecurity valida la seguridad de una URL antes de acceder
func validateBrowserSecurity(url string) error {
	if err := security.ValidateURL(url); err != nil {
		return fmt.Errorf("URL security validation failed: %w", err)
	}
	return nil
}

// validateCodeSecurity valida la seguridad de una ruta de archivo
func validateCodeSecurity(path string) error {
	if err := security.ValidatePath(path, "/app"); err != nil {
		return fmt.Errorf("path security validation failed: %w", err)
	}
	return nil
}

// updateTaskResults is a helper to update task results in the database
func updateTaskResults(db *database.Queries, taskID int64, results string) error {
	_, err := db.UpdateTaskResults(context.Background(), database.UpdateTaskResultsParams{
		ID:      taskID,
		Results: database.StringToNullString(results),
	})
	if err != nil {
		return fmt.Errorf("failed to update task results: %w", err)
	}
	return nil
}

func processBrowserTask(db *database.Queries, task database.Task) error {
	args, err := unmarshalTaskArgs[providers.BrowserArgs](task)
	if err != nil {
		return err
	}

	if err := validateBrowserSecurity(args.Url); err != nil {
		return err
	}

	// Select the appropriate browser action based on the action type
	var actionFn BrowserActionFunc
	switch args.Action {
	case providers.Read:
		actionFn = Content
	case providers.Url:
		actionFn = URLs
	default:
		return fmt.Errorf("unknown browser action: %s", args.Action)
	}

	// Execute the browser action
	content, screenshotName, err := actionFn(args.Url)
	if err != nil {
		return fmt.Errorf("failed to execute browser action: %w", err)
	}

	logging.Debug("Browser action completed", "url", args.Url, "action", args.Action, "screenshot", screenshotName)

	// Update task results
	if err := updateTaskResults(db, task.ID, content); err != nil {
		return err
	}

	// Broadcast browser update
	baseURL := fmt.Sprintf("http://localhost:%d", config.Config.Port)
	subscriptions.BroadcastBrowserUpdated(task.FlowID.Int64, &gmodel.Browser{
		URL:           args.Url,
		ScreenshotURL: baseURL + "/browser/" + screenshotName,
	})

	return nil
}

func processDoneTask(db *database.Queries, task database.Task) error {
	flow, err := db.UpdateFlowStatus(context.Background(), database.UpdateFlowStatusParams{
		ID:     task.FlowID.Int64,
		Status: database.StringToNullString(string(models.FlowFinished)),
	})

	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	subscriptions.BroadcastFlowUpdated(task.FlowID.Int64, &gmodel.Flow{
		ID:       uint(flow.ID),
		Status:   gmodel.FlowStatus(models.FlowFinished),
		Terminal: &gmodel.Terminal{},
	})

	return nil
}

func processInputTask(provider providers.Provider, db *database.Queries, task database.Task) error {
	tasks, err := db.ReadTasksByFlowId(context.Background(), sql.NullInt64{
		Int64: task.FlowID.Int64,
		Valid: true,
	})

	if err != nil {
		return fmt.Errorf("failed to get tasks by flow id: %w", err)
	}

	// This is the first task in the flow.
	// We need to get the basic flow data as well as spin up the container
	if len(tasks) == 1 {
		summary, err := provider.Summary(task.Message.String, SummaryWordCount)

		if err != nil {
			return fmt.Errorf("failed to get message summary: %w", err)
		}

		dockerImage, err := provider.DockerImageName(task.Message.String)

		if err != nil {
			return fmt.Errorf("failed to get docker image name: %w", err)
		}

		flow, err := db.UpdateFlowName(context.Background(), database.UpdateFlowNameParams{
			ID:   task.FlowID.Int64,
			Name: database.StringToNullString(summary),
		})

		if err != nil {
			return fmt.Errorf("failed to update flow: %w", err)
		}

		subscriptions.BroadcastFlowUpdated(flow.ID, &gmodel.Flow{
			ID:   uint(flow.ID),
			Name: summary,
			Terminal: &gmodel.Terminal{
				ContainerName: dockerImage,
				Connected:     false,
			},
		})

		msg := fmt.Sprintf("Initializing the docker image %s...", dockerImage)
		if err := createAndBroadcastLog(flow.ID, msg, LogTypeSystem, db); err != nil {
			return err
		}

		terminalContainerName := TerminalName(flow.ID)
		terminalContainerID, err := SpawnContainer(context.Background(),
			terminalContainerName,
			&container.Config{
				Image: dockerImage,
				Cmd:   []string{"tail", "-f", "/dev/null"},
			},
			&container.HostConfig{},
			db,
		)

		if err != nil {
			return fmt.Errorf("failed to spawn container: %w", err)
		}

		subscriptions.BroadcastFlowUpdated(flow.ID, &gmodel.Flow{
			ID:   uint(flow.ID),
			Name: summary,
			Terminal: &gmodel.Terminal{
				Connected:     true,
				ContainerName: dockerImage,
			},
		})

		_, err = db.UpdateFlowContainer(context.Background(), database.UpdateFlowContainerParams{
			ID:          flow.ID,
			ContainerID: sql.NullInt64{Int64: terminalContainerID, Valid: true},
		})

		if err != nil {
			return fmt.Errorf("failed to update flow container: %w", err)
		}

		msg = "Container initialized. Ready to execute commands."
		if err := createAndBroadcastLog(flow.ID, msg, LogTypeSystem, db); err != nil {
			return err
		}
	}

	return nil
}

func processAskTask(db *database.Queries, task database.Task) error {
	task, err := db.UpdateTaskStatus(context.Background(), database.UpdateTaskStatusParams{
		Status: database.StringToNullString(models.TaskFinished),
		ID:     task.ID,
	})

	if err != nil {
		return fmt.Errorf("failed to find task with id %d: %w", task.ID, err)
	}

	return nil
}

func processTerminalTask(db *database.Queries, task database.Task) error {
	args, err := unmarshalTaskArgs[providers.TerminalArgs](task)
	if err != nil {
		return err
	}

	results, err := ExecCommand(task.FlowID.Int64, args.Input, db)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	return updateTaskResults(db, task.ID, results)
}

func processCodeTask(db *database.Queries, task database.Task) error {
	args, err := unmarshalTaskArgs[providers.CodeArgs](task)
	if err != nil {
		return err
	}

	if err := validateCodeSecurity(args.Path); err != nil {
		return err
	}

	var results string

	switch args.Action {
	case providers.ReadFile:
		// Use quoted path to prevent command injection
		cmd := fmt.Sprintf("cat '%s'", args.Path)
		var execErr error
		results, execErr = ExecCommand(task.FlowID.Int64, cmd, db)
		if execErr != nil {
			return fmt.Errorf("error executing cat command: %w", execErr)
		}

	case providers.UpdateFile:
		if writeErr := WriteFile(task.FlowID.Int64, args.Content, args.Path, db); writeErr != nil {
			return fmt.Errorf("error writing a file: %w", writeErr)
		}
		results = "File updated"

	default:
		return fmt.Errorf("unknown code action: %s", args.Action)
	}

	return updateTaskResults(db, task.ID, results)
}
