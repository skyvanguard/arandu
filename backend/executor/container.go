package executor

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/arandu-ai/arandu/database"
	"github.com/arandu-ai/arandu/logging"
	"github.com/arandu-ai/arandu/models"
	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

var (
	dockerClient *client.Client
)

const defaultImage = "debian:latest"

// ensureImageExists verifica si la imagen existe localmente, si no la descarga
// Retorna el nombre de la imagen a usar (puede ser defaultImage si falla el pull)
func ensureImageExists(ctx context.Context, imageName string) string {
	filterArgs := filters.NewArgs()
	filterArgs.Add("reference", imageName)
	images, err := dockerClient.ImageList(ctx, image.ListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		logging.Warn("Error listing images, using default", "error", err.Error())
		return defaultImage
	}

	if len(images) > 0 {
		logging.Debug("Image exists locally", "image", imageName)
		return imageName
	}

	logging.Info("Pulling image", "image", imageName)
	readCloser, pullErr := dockerClient.ImagePull(ctx, imageName, image.PullOptions{})
	if pullErr != nil {
		logging.Warn("Failed to pull image, using default",
			"error", pullErr.Error(),
			"default_image", defaultImage,
		)
		return defaultImage
	}

	// Wait for the pull to finish
	if _, copyErr := io.Copy(io.Discard, readCloser); copyErr != nil {
		logging.Warn("Error waiting for image pull", "error", copyErr.Error())
	}

	return imageName
}

// createAndStartContainer crea e inicia un container Docker
func createAndStartContainer(ctx context.Context, name string, cfg *container.Config, hostCfg *container.HostConfig) (containerID string, err error) {
	logging.Debug("Creating container", "name", name)
	resp, err := dockerClient.ContainerCreate(ctx, cfg, hostCfg, nil, nil, name)
	if err != nil {
		return "", fmt.Errorf("error creating container: %w", err)
	}

	if err := dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return resp.ID, fmt.Errorf("error starting container: %w", err)
	}

	return resp.ID, nil
}

// updateContainerInDB actualiza el estado y localID del container en la base de datos
func updateContainerInDB(ctx context.Context, db *database.Queries, dbID int64, localID string, status models.ContainerStatus) {
	if _, err := db.UpdateContainerStatus(ctx, database.UpdateContainerStatusParams{
		ID:     dbID,
		Status: database.StringToNullString(string(status)),
	}); err != nil {
		logging.Error("Failed to update container status", "error", err.Error())
	}

	if _, err := db.UpdateContainerLocalId(ctx, database.UpdateContainerLocalIdParams{
		ID:      dbID,
		LocalID: database.StringToNullString(localID),
	}); err != nil {
		logging.Error("Failed to update container local id", "error", err.Error())
	}
}

func InitClient() error {
	start := time.Now()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return fmt.Errorf("error initializing docker client: %w", err)
	}
	cli.NegotiateAPIVersion(context.Background())

	dockerClient = cli
	info, err := dockerClient.Info(context.Background())
	if err != nil {
		return fmt.Errorf("error getting docker info: %w", err)
	}

	logging.LogDockerOp("init_client", "", time.Since(start), nil,
		"name", info.Name,
		"arch", info.Architecture,
		"server_version", info.ServerVersion,
		"client_version", dockerClient.ClientVersion(),
	)

	return nil
}

func SpawnContainer(ctx context.Context, name string, config *container.Config, hostConfig *container.HostConfig, db *database.Queries) (dbContainerID int64, err error) {
	start := time.Now()

	if config == nil {
		return 0, fmt.Errorf("no config found for container %s", name)
	}

	logging.Info("Spawning container", "image", config.Image, "name", name)

	// Create DB record
	dbContainer, err := db.CreateContainer(ctx, database.CreateContainerParams{
		Name:   database.StringToNullString(name),
		Image:  database.StringToNullString(config.Image),
		Status: database.StringToNullString(string(models.ContainerStarting)),
	})
	if err != nil {
		return 0, fmt.Errorf("error creating container in database: %w", err)
	}

	var localContainerID string

	// Cleanup on error or update status on success
	defer func() {
		status := models.ContainerRunning
		if err != nil {
			status = models.ContainerFailed
			if stopErr := StopContainer(localContainerID, dbContainer.ID, db); stopErr != nil {
				logging.Error("Failed to stop container after error",
					"container_id", dbContainer.ID,
					"error", stopErr.Error(),
				)
			}
		}
		updateContainerInDB(ctx, db, dbContainer.ID, localContainerID, status)
	}()

	// Ensure image is available
	config.Image = ensureImageExists(ctx, config.Image)

	// Create and start container
	localContainerID, err = createAndStartContainer(ctx, name, config, hostConfig)
	if err != nil {
		return dbContainer.ID, err
	}

	logging.LogDockerOp("spawn_container", localContainerID, time.Since(start), nil,
		"name", name,
		"image", config.Image,
	)

	return dbContainer.ID, nil
}

func StopContainer(containerID string, dbID int64, db *database.Queries) error {
	start := time.Now()

	if err := dockerClient.ContainerStop(context.Background(), containerID, container.StopOptions{}); err != nil {
		if errdefs.IsNotFound(err) {
			logging.Debug("Container not found, marking as stopped", "container_id", containerID)
			_, _ = db.UpdateContainerStatus(context.Background(), database.UpdateContainerStatusParams{
				Status: database.StringToNullString("stopped"),
				ID:     dbID,
			})
			return nil
		}
		return fmt.Errorf("error stopping container: %w", err)
	}

	_, err := db.UpdateContainerStatus(context.Background(), database.UpdateContainerStatusParams{
		Status: database.StringToNullString("stopped"),
		ID:     dbID,
	})
	if err != nil {
		return fmt.Errorf("error updating container status to stopped: %w", err)
	}

	logging.LogDockerOp("stop_container", containerID, time.Since(start), nil)
	return nil
}

func DeleteContainer(containerID string, dbID int64, db *database.Queries) error {
	start := time.Now()

	if err := StopContainer(containerID, dbID, db); err != nil {
		return fmt.Errorf("error stopping container: %w", err)
	}

	if err := dockerClient.ContainerRemove(context.Background(), containerID, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("error removing container: %w", err)
	}

	logging.LogDockerOp("delete_container", containerID, time.Since(start), nil)
	return nil
}

func Cleanup(db *database.Queries) error {
	start := time.Now()

	// Remove tmp files
	logging.Info("Cleanup starting", "step", "remove_tmp_files")
	if err := os.RemoveAll("./tmp/"); err != nil {
		return fmt.Errorf("error removing tmp files: %w", err)
	}

	logging.Info("Cleanup", "step", "stop_containers")

	var wg sync.WaitGroup

	containers, err := db.GetAllRunningContainers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting running containers: %w", err)
	}

	for _, c := range containers {
		wg.Add(1)
		go func(cont database.Container) {
			defer wg.Done()
			localId := cont.LocalID.String
			if err := DeleteContainer(localId, cont.ID, db); err != nil {
				logging.Error("Failed to delete container",
					"container_id", localId,
					"error", err.Error(),
				)
			}
		}(c)
	}

	wg.Wait()

	// Update flows status
	flows, err := db.ReadAllFlows(context.Background())
	if err != nil {
		return fmt.Errorf("error getting all flows: %w", err)
	}

	for _, flow := range flows {
		if flow.Status.String == string(models.FlowInProgress) {
			_, err := db.UpdateFlowStatus(context.Background(), database.UpdateFlowStatusParams{
				Status: database.StringToNullString(string(models.FlowFinished)),
				ID:     flow.ID,
			})
			if err != nil {
				logging.Error("Failed to update flow status", "flow_id", flow.ID, "error", err.Error())
			}
		}
	}

	logging.Info("Cleanup completed",
		"containers_cleaned", len(containers),
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return nil
}

func IsContainerRunning(containerID string) (bool, error) {
	containerInfo, err := dockerClient.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return false, fmt.Errorf("error inspecting container: %w", err)
	}
	return containerInfo.State.Running, nil
}
