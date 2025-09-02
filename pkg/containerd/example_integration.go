package containerd

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
)

// ExampleIntegration demonstrates how to use the containerD client for common operations
type ExampleIntegration struct {
	manager *Manager
}

// NewExampleIntegration creates a new example integration
func NewExampleIntegration(manager *Manager) *ExampleIntegration {
	return &ExampleIntegration{
		manager: manager,
	}
}

// DeployContainer demonstrates deploying a simple container
func (e *ExampleIntegration) DeployContainer(ctx context.Context, containerID, imageRef string) error {
	client, err := e.manager.GetClient()
	if err != nil {
		return fmt.Errorf("failed to get containerd client: %w", err)
	}

	// Pull the image first
	ctx = client.WithNamespace(ctx)
	containerdClient := client.GetContainerdClient()
	image, err := containerdClient.Pull(ctx, imageRef, containerd.WithPullUnpack)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	// Create container with basic configuration
	container, err := containerdClient.NewContainer(ctx, containerID,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(containerID, image),
		containerd.WithNewSpec(oci.WithImageConfig(image)),
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Create a task (running instance of the container)
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Start the task
	if err := task.Start(ctx); err != nil {
		return fmt.Errorf("failed to start task: %w", err)
	}

	fmt.Printf("Container %s deployed successfully\n", containerID)
	return nil
}

// ListAllContainers demonstrates listing containers with their status
func (e *ExampleIntegration) ListAllContainers(ctx context.Context) error {
	client, err := e.manager.GetClient()
	if err != nil {
		return fmt.Errorf("failed to get containerd client: %w", err)
	}

	containers, err := client.ListContainers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	fmt.Printf("Found %d containers:\n", len(containers))
	for _, container := range containers {
		info, err := container.Info(ctx)
		if err != nil {
			fmt.Printf("  %s: failed to get info - %v\n", container.ID(), err)
			continue
		}

		// Check if container has a running task
		task, err := container.Task(ctx, nil)
		status := "stopped"
		if err == nil {
			taskStatus, err := task.Status(ctx)
			if err == nil {
				status = string(taskStatus.Status)
			}
		}

		fmt.Printf("  %s: image=%s, status=%s\n", container.ID(), info.Image, status)
	}

	return nil
}

// StopContainer demonstrates stopping a running container
func (e *ExampleIntegration) StopContainer(ctx context.Context, containerID string) error {
	client, err := e.manager.GetClient()
	if err != nil {
		return fmt.Errorf("failed to get containerd client: %w", err)
	}

	container, err := client.GetContainer(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to get container: %w", err)
	}

	// Get the task
	task, err := container.Task(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Kill the task
	if err := task.Kill(ctx, 15); err != nil { // SIGTERM
		return fmt.Errorf("failed to kill task: %w", err)
	}

	// Wait for the task to stop
	_, err = task.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for task: %w", err)
	}

	// Delete the task
	if _, err := task.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	fmt.Printf("Container %s stopped successfully\n", containerID)
	return nil
}

// RemoveContainer demonstrates removing a container completely
func (e *ExampleIntegration) RemoveContainer(ctx context.Context, containerID string) error {
	client, err := e.manager.GetClient()
	if err != nil {
		return fmt.Errorf("failed to get containerd client: %w", err)
	}

	container, err := client.GetContainer(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to get container: %w", err)
	}

	// Stop the container first if it's running
	task, err := container.Task(ctx, nil)
	if err == nil {
		// Task exists, so stop it
		if err := task.Kill(ctx, 15); err == nil {
			task.Wait(ctx)
			task.Delete(ctx)
		}
	}

	// Delete the container
	if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}

	fmt.Printf("Container %s removed successfully\n", containerID)
	return nil
}
