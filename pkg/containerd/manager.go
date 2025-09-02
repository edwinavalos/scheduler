package containerd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Manager handles containerD client lifecycle and connection management
type Manager struct {
	config           Config
	client           *Client
	healthCheck      *HealthCheck
	mu               sync.RWMutex
	connected        bool
	reconnectTimeout time.Duration
	logger           *logrus.Logger
}

// NewManager creates a new containerD connection manager
func NewManager(config Config, logger *logrus.Logger) *Manager {
	if logger == nil {
		logger = logrus.New()
	}

	return &Manager{
		config:           config,
		reconnectTimeout: 30 * time.Second,
		logger:           logger,
	}
}

// Start initializes the containerD client and starts connection management
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.connect(); err != nil {
		return fmt.Errorf("failed to start containerd manager: %w", err)
	}

	// Start health check routine
	go m.healthCheckLoop(ctx)

	m.logger.Info("ContainerD manager started successfully")
	return nil
}

// Stop closes the containerD client connection
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client != nil {
		if err := m.client.Close(); err != nil {
			m.logger.WithError(err).Error("Error closing containerd client")
		}
		m.client = nil
		m.healthCheck = nil
	}

	m.connected = false
	m.logger.Info("ContainerD manager stopped")
	return nil
}

// GetClient returns the containerD client if connected
func (m *Manager) GetClient() (*Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected || m.client == nil {
		return nil, errors.New("containerd client is not connected")
	}

	return m.client, nil
}

// IsConnected returns whether the client is currently connected
func (m *Manager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// connect establishes a connection to containerD
func (m *Manager) connect() error {
	client, err := NewClient(m.config)
	if err != nil {
		return err
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		client.Close()
		return fmt.Errorf("containerd ping failed: %w", err)
	}

	m.client = client
	m.healthCheck = NewHealthCheck(client)
	m.connected = true

	m.logger.WithFields(logrus.Fields{
		"socket":    m.config.Socket,
		"namespace": m.config.Namespace,
	}).Info("Connected to containerd")

	return nil
}

// reconnect attempts to reconnect to containerD
func (m *Manager) reconnect() error {
	m.logger.Info("Attempting to reconnect to containerd")

	// Close existing connection
	if m.client != nil {
		m.client.Close()
		m.client = nil
		m.healthCheck = nil
	}

	m.connected = false

	// Attempt to reconnect
	if err := m.connect(); err != nil {
		m.logger.WithError(err).Error("Failed to reconnect to containerd")
		return err
	}

	m.logger.Info("Successfully reconnected to containerd")
	return nil
}

// healthCheckLoop continuously monitors the connection health
func (m *Manager) healthCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.performHealthCheck()
		}
	}
}

// performHealthCheck checks the connection health and attempts to reconnect if needed
func (m *Manager) performHealthCheck() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected || m.healthCheck == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.healthCheck.Check(ctx); err != nil {
		m.logger.WithError(err).Warn("ContainerD health check failed, attempting to reconnect")
		m.connected = false

		// Attempt to reconnect
		go func() {
			time.Sleep(5 * time.Second) // Brief delay before reconnect
			m.mu.Lock()
			defer m.mu.Unlock()

			if err := m.reconnect(); err != nil {
				m.logger.WithError(err).Error("Failed to reconnect after health check failure")
			}
		}()
	}
}
