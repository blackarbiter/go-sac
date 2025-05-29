package scanner

import (
	"context"
	"fmt"
	"github.com/blackarbiter/go-sac/pkg/metrics"
	"runtime"
	"strconv"
	"time"

	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/logger"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ContainerExecutor implements TaskExecutor interface using Docker
type ContainerExecutor struct {
	dockerClient   *client.Client
	timeoutManager *TimeoutController
	metrics        metrics.ContainerMetrics
	logger         *zap.Logger
	config         ContainerConfig
}

// calculateTimeout calculates the timeout duration for a task
func calculateTimeout(task *domain.ScanTaskPayload) time.Duration {
	baseTimeout := 5 * time.Minute
	if timeout, ok := task.Options["timeout"]; ok {
		if duration, err := time.ParseDuration(timeout.(string)); err == nil {
			return duration
		}
	}
	return baseTimeout
}

// NewContainerExecutor creates a new container executor instance
func NewContainerExecutor(config ContainerConfig) (*ContainerExecutor, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &ContainerExecutor{
		dockerClient: dockerClient,
		config:       config,
		logger:       logger.Logger,
	}, nil
}

// Meta returns executor metadata
func (e *ContainerExecutor) Meta() ExecutorMeta {
	return ExecutorMeta{
		Type:    "container",
		Version: "1.0.0",
		ResourceProfile: ResourceProfile{
			MinCPU:   1,
			MaxCPU:   4,
			MemoryMB: 2048,
		},
	}
}

// AsyncExecute implements asynchronous task execution
func (e *ContainerExecutor) AsyncExecute(ctx context.Context, task *domain.ScanTaskPayload) (string, error) {
	executionID := uuid.New().String()
	statusChan := make(chan containerStatus, 1)

	go e.monitorContainerExecution(ctx, executionID, task, statusChan)

	select {
	case status := <-statusChan:
		if status.err != nil {
			return "", fmt.Errorf("container startup failed: %w", status.err)
		}
		return executionID, nil
	case <-ctx.Done():
		go e.cleanupPendingResources(executionID)
		return "", ctx.Err()
	}
}

// containerStatus represents container execution status
type containerStatus struct {
	err error
}

// monitorContainerExecution handles container lifecycle
func (e *ContainerExecutor) monitorContainerExecution(ctx context.Context, id string, task *domain.ScanTaskPayload, statusChan chan<- containerStatus) {
	timeout := calculateTimeout(task) + e.config.DeadlineBuffer
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	containerID, err := e.startContainer(ctx, task)
	if err != nil {
		statusChan <- containerStatus{err: err}
		return
	}
	close(statusChan)

	statusCh, errCh := e.dockerClient.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	select {
	case status := <-statusCh:
		e.handleCompletion(ctx, id, task, status)
	case err := <-errCh:
		e.handleExecutionError(ctx, id, task, err)
	case <-ctx.Done():
		e.handleTimeout(ctx, id, task, containerID)
	}
}

// startContainer creates and starts a container
func (e *ContainerExecutor) startContainer(ctx context.Context, task *domain.ScanTaskPayload) (string, error) {
	config := e.buildContainerConfig(task)

	// Convert string resource limits to int64
	memoryLimit, _ := strconv.ParseInt(e.config.ResourceQuota.MemoryLimit, 10, 64)
	cpuLimit, _ := strconv.ParseInt(e.config.ResourceQuota.CPULimit, 10, 64)

	resp, err := e.dockerClient.ContainerCreate(
		ctx,
		config,
		&container.HostConfig{
			Resources: container.Resources{
				Memory:     memoryLimit,
				MemorySwap: -1,
				CPUQuota:   cpuLimit,
			},
			SecurityOpt: []string{
				fmt.Sprintf("no-new-privileges:%v", !e.config.SecurityContext.AllowPrivilegeEscalation),
			},
		},
		&network.NetworkingConfig{},
		nil,
		task.TaskID, // Use TaskID as container name
	)
	if err != nil {
		e.metrics.ContainerCreateError(task.ScanType)
		return "", fmt.Errorf("container creation failed: %w", err)
	}

	if err := e.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		e.metrics.ContainerStartError(task.ScanType)
		return "", fmt.Errorf("container start failed: %w", err)
	}

	e.logger.Info("container started",
		zap.String("taskID", task.TaskID),
		zap.String("containerID", resp.ID),
	)

	go e.collectLogs(resp.ID, task)

	return resp.ID, nil
}

// buildContainerConfig creates container configuration
func (e *ContainerExecutor) buildContainerConfig(task *domain.ScanTaskPayload) *container.Config {
	envVars := []string{
		fmt.Sprintf("TASK_ID=%s", task.TaskID),
		fmt.Sprintf("SCAN_TYPE=%s", task.ScanType),
	}

	return &container.Config{
		Image:        e.resolveImage(task),
		Env:          envVars,
		Cmd:          e.buildCommand(task),
		Labels:       e.buildLabels(task),
		AttachStdout: true,
		AttachStderr: true,
	}
}

// resolveImage determines the correct container image
func (e *ContainerExecutor) resolveImage(task *domain.ScanTaskPayload) string {
	if customImage, ok := task.Options["custom_image"]; ok {
		if imageStr, ok := customImage.(string); ok {
			return e.verifyImageHash(imageStr)
		}
	}
	arch := runtime.GOARCH
	return fmt.Sprintf("%s-%s", e.config.DefaultImage, arch)
}

// verifyImageHash verifies image integrity
func (e *ContainerExecutor) verifyImageHash(image string) string {
	// TODO: Implement image hash verification
	return image
}

// buildCommand constructs container command
func (e *ContainerExecutor) buildCommand(task *domain.ScanTaskPayload) []string {
	// TODO: Implement command building logic
	return []string{"scan", string(task.ScanType)}
}

// buildLabels creates container labels
func (e *ContainerExecutor) buildLabels(task *domain.ScanTaskPayload) map[string]string {
	return map[string]string{
		"task.id":   task.TaskID,
		"scan.type": string(task.ScanType),
	}
}

// collectLogs handles container log collection
func (e *ContainerExecutor) collectLogs(containerID string, task *domain.ScanTaskPayload) {
	// TODO: Implement log collection logic
}

// handleCompletion processes successful task completion
func (e *ContainerExecutor) handleCompletion(ctx context.Context, id string, task *domain.ScanTaskPayload, status container.WaitResponse) {
	// TODO: Implement completion handling
}

// handleExecutionError processes execution errors
func (e *ContainerExecutor) handleExecutionError(ctx context.Context, id string, task *domain.ScanTaskPayload, err error) {
	// TODO: Implement error handling
}

// handleTimeout processes execution timeouts
func (e *ContainerExecutor) handleTimeout(ctx context.Context, id string, task *domain.ScanTaskPayload, containerID string) {
	// TODO: Implement timeout handling
}

// cleanupPendingResources cleans up resources for failed tasks
func (e *ContainerExecutor) cleanupPendingResources(taskID string) {
	// TODO: Implement resource cleanup
}

// Cancel implements task cancellation
func (e *ContainerExecutor) Cancel(handle string) error {
	// TODO: Implement cancellation logic
	return nil
}

// GetStatus retrieves task status
func (e *ContainerExecutor) GetStatus(handle string) (domain.TaskStatus, error) {
	// TODO: Implement status retrieval
	return domain.TaskStatusPending, nil
}

// HealthCheck performs executor health check
func (e *ContainerExecutor) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := e.dockerClient.Ping(ctx)
	return err
}
