// Package containerd provides a wrapper around the containerD client with connection management,
// health checking, and configuration support.
//
// The package includes:
// - Client: A wrapper around containerd.Client with higher-level operations
// - Manager: Connection lifecycle management with automatic reconnection
// - Config: Configuration structures for containerD settings
// - HealthCheck: Health monitoring for containerD service
//
// Example usage:
//
//	config := containerd.DefaultConfig()
//	manager := containerd.NewManager(config, logger)
//
//	ctx := context.Background()
//	if err := manager.Start(ctx); err != nil {
//		log.Fatal(err)
//	}
//	defer manager.Stop()
//
//	client, err := manager.GetClient()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	containers, err := client.ListContainers(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
package containerd
