package executor

import (
	"archive/tar"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/arandu-ai/arandu/database"
	gmodel "github.com/arandu-ai/arandu/graph/model"
	"github.com/arandu-ai/arandu/graph/subscriptions"
	"github.com/arandu-ai/arandu/logging"
	"github.com/arandu-ai/arandu/websocket"
)

// LogType representa el tipo de log de terminal
type LogType string

const (
	LogTypeInput  LogType = "input"
	LogTypeOutput LogType = "output"
	LogTypeSystem LogType = "system"
)

// createAndBroadcastLog crea un log en la base de datos y lo transmite via subscriptions
// Esta función centraliza el patrón repetido de crear log + broadcast
func createAndBroadcastLog(flowID int64, message string, logType LogType, db *database.Queries) error {
	log, err := db.CreateLog(context.Background(), database.CreateLogParams{
		FlowID:  sql.NullInt64{Int64: flowID, Valid: true},
		Message: message,
		Type:    string(logType),
	})

	if err != nil {
		logging.Error("Error creating terminal log",
			"flow_id", flowID,
			"type", logType,
			"error", err.Error(),
		)
		return fmt.Errorf("error creating log: %w", err)
	}

	// Formatear el texto según el tipo
	var text string
	switch logType {
	case LogTypeInput:
		text = websocket.FormatTerminalInput(message)
	case LogTypeSystem:
		text = websocket.FormatTerminalSystemOutput(message)
	default:
		text = message
	}

	subscriptions.BroadcastTerminalLogsAdded(flowID, &gmodel.Log{
		ID:   uint(log.ID),
		Text: text,
	})

	return nil
}

func ExecCommand(flowID int64, command string, db *database.Queries) (result string, err error) {
	containerName, err := ensureContainerRunning(flowID)
	if err != nil {
		return "", err
	}

	// Create options for starting the exec process
	cmd := []string{
		"sh",
		"-c",
		command,
	}

	// Log input command
	if err := createAndBroadcastLog(flowID, command, LogTypeInput, db); err != nil {
		return "", err
	}

	createResp, err := dockerClient.ContainerExecCreate(context.Background(), containerName, container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	})
	if err != nil {
		return "", fmt.Errorf("Error creating exec process: %w", err)
	}

	// Attach to the exec process
	resp, err := dockerClient.ContainerExecAttach(context.Background(), createResp.ID, container.ExecAttachOptions{
		Tty: true,
	})
	if err != nil {
		return "", fmt.Errorf("Error attaching to exec process: %w", err)
	}
	defer resp.Close()

	dst := bytes.Buffer{}
	_, err = io.Copy(&dst, resp.Reader)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("Error copying output: %w", err)
	}

	// Wait for the exec process to finish
	_, err = dockerClient.ContainerExecInspect(context.Background(), createResp.ID)
	if err != nil {
		return "", fmt.Errorf("Error inspecting exec process: %w", err)
	}

	results := dst.String()

	// Log output result
	if err := createAndBroadcastLog(flowID, results, LogTypeOutput, db); err != nil {
		return "", err
	}

	result = dst.String()

	if result == "" {
		result = "Command executed successfully"
	}

	return result, nil
}

func WriteFile(flowID int64, content string, path string, db *database.Queries) (err error) {
	containerName, err := ensureContainerRunning(flowID)
	if err != nil {
		return err
	}

	// Log file content as input
	if err := createAndBroadcastLog(flowID, content, LogTypeInput, db); err != nil {
		return err
	}

	// Put content into a tar archive
	archive := &bytes.Buffer{}
	tarWriter := tar.NewWriter(archive)
	filename := filepath.Base(path)
	tarHeader := &tar.Header{
		Name: filename,
		Mode: 0600,
		Size: int64(len(content)),
	}
	err = tarWriter.WriteHeader(tarHeader)
	if err != nil {
		return fmt.Errorf("Error writing tar header: %w", err)
	}

	_, err = tarWriter.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("Error writing tar content: %w", err)
	}

	dir := filepath.Dir(path)
	err = dockerClient.CopyToContainer(context.Background(), containerName, dir, archive, container.CopyToContainerOptions{})

	if err != nil {
		return fmt.Errorf("Error writing file: %w", err)
	}

	message := fmt.Sprintf("Wrote to %s", path)

	// Log success message
	if err := createAndBroadcastLog(flowID, message, LogTypeOutput, db); err != nil {
		return err
	}

	return nil
}

func TerminalName(flowID int64) string {
	return fmt.Sprintf("arandu-terminal-%d", flowID)
}

// ensureContainerRunning verifica que el container del flow está corriendo
// Retorna el nombre del container si está corriendo, o un error si no
func ensureContainerRunning(flowID int64) (containerName string, err error) {
	containerName = TerminalName(flowID)

	isRunning, err := IsContainerRunning(containerName)
	if err != nil {
		return "", fmt.Errorf("error inspecting container: %w", err)
	}

	if !isRunning {
		return "", fmt.Errorf("container is not running")
	}

	return containerName, nil
}
