package containerd

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
)

// Client wraps the containerD client and provides higher-level operations
type Client struct {
	client    *containerd.Client
	config    Config
	namespace string
}

// GetContainerdClient returns the underlying containerd client for advanced operations
func (c *Client) GetContainerdClient() *containerd.Client {
	return c.client
}

// NewClient creates a new containerD client wrapper
func NewClient(config Config) (*Client, error) {
	client, err := containerd.New(config.Socket)
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w", err)
	}

	return &Client{
		client:    client,
		config:    config,
		namespace: config.Namespace,
	}, nil
}

// Close closes the containerD client connection
func (c *Client) Close() error {
	return c.client.Close()
}

// WithNamespace returns a context with the configured namespace
func (c *Client) WithNamespace(ctx context.Context) context.Context {
	return namespaces.WithNamespace(ctx, c.namespace)
}

// Ping tests the connection to containerD
func (c *Client) Ping(ctx context.Context) error {
	ctx = c.WithNamespace(ctx)

	// Try to get server version as a simple connectivity test
	_, err := c.client.Version(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping containerd: %w", err)
	}

	return nil
}

// ListContainers returns all containers in the namespace
func (c *Client) ListContainers(ctx context.Context) ([]containerd.Container, error) {
	ctx = c.WithNamespace(ctx)

	containers, err := c.client.Containers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	return containers, nil
}

// GetContainer retrieves a container by ID
func (c *Client) GetContainer(ctx context.Context, id string) (containerd.Container, error) {
	ctx = c.WithNamespace(ctx)

	container, err := c.client.LoadContainer(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get container %s: %w", id, err)
	}

	return container, nil
}

// CreateContainer creates a new container with the given image and options
func (c *Client) CreateContainer(ctx context.Context, id, image string, opts ...containerd.NewContainerOpts) (containerd.Container, error) {
	ctx = c.WithNamespace(ctx)

	// Pull image if it doesn't exist
	img, err := c.pullImage(ctx, image)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image %s: %w", image, err)
	}

	// Add image to container options
	opts = append(opts, containerd.WithImage(img))

	container, err := c.client.NewContainer(ctx, id, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create container %s: %w", id, err)
	}

	return container, nil
}

// pullImage pulls an image from registry if it doesn't exist locally
func (c *Client) pullImage(ctx context.Context, ref string) (containerd.Image, error) {
	// Check if image already exists
	img, err := c.client.GetImage(ctx, ref)
	if err == nil {
		return img, nil
	}

	// Pull the image
	img, err = c.client.Pull(ctx, ref, containerd.WithPullUnpack)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image %s: %w", ref, err)
	}

	return img, nil
}

// Health check for the containerD service
type HealthCheck struct {
	client *Client
}

// NewHealthCheck creates a new health checker
func NewHealthCheck(client *Client) *HealthCheck {
	return &HealthCheck{client: client}
}

// Check performs a health check on the containerD service
func (h *HealthCheck) Check(ctx context.Context) error {
	// Set a timeout for the health check
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return h.client.Ping(ctx)
}
