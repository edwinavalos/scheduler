package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "scheduler/proto/gen"
)

const (
	defaultAddress = "localhost:8000"
)

func main() {
	// Connect to the gRPC server
	conn, err := grpc.Dial(defaultAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Create the client
	client := pb.NewSchedulerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("🚀 Testing Scheduler Service gRPC Client")
	fmt.Println("=========================================")

	// Test all the service methods
	testCreateEnvironment(ctx, client)
	testListEnvironments(ctx, client)
	testGetEnvironment(ctx, client)
	testStartEnvironment(ctx, client)
	testStopEnvironment(ctx, client)
	testRestartEnvironment(ctx, client)
	testGetEnvironmentStatus(ctx, client)
	testUpdateEnvironment(ctx, client)
	testGetEnvironmentLogs(ctx, client)
	testDeleteEnvironment(ctx, client)

	fmt.Println("\n✅ All test calls completed!")
}

func testCreateEnvironment(ctx context.Context, client pb.SchedulerServiceClient) {
	fmt.Println("\n📝 Testing CreateEnvironment...")

	// Create a comprehensive environment specification
	req := &pb.CreateEnvironmentRequest{
		Spec: &pb.EnvironmentSpecification{
			Name:        "test-webapp-stack",
			Description: "A complete web application stack with frontend, backend, and database",
			ApplicationStack: &pb.ApplicationStack{
				Name:    "webapp-stack",
				Version: "1.0.0",
				Frontend: &pb.FrontendConfig{
					Container: &pb.ContainerConfig{
						Name:  "webapp-frontend",
						Image: "nginx:alpine",
						Ports: []*pb.PortMapping{
							{
								ContainerPort: 80,
								HostPort:      8080,
								Protocol:      "tcp",
							},
						},
						EnvironmentVariables: map[string]string{
							"API_URL":     "http://backend:3000",
							"NODE_ENV":    "production",
							"SERVER_NAME": "webapp.local",
						},
						Resources: &pb.ResourceLimits{
							MemoryMb: 512,
							CpuCores: 0.5,
							DiskMb:   1024,
						},
						HealthCheck: &pb.HealthCheck{
							Command:            []string{"curl", "-f", "http://localhost/health"},
							IntervalSeconds:    30,
							TimeoutSeconds:     5,
							Retries:            3,
							StartPeriodSeconds: 10,
						},
						RestartPolicy: pb.RestartPolicy_RESTART_POLICY_UNLESS_STOPPED,
					},
					Domains:    []string{"webapp.local", "www.webapp.local"},
					SslEnabled: false,
				},
				Backend: &pb.BackendConfig{
					Container: &pb.ContainerConfig{
						Name:    "webapp-backend",
						Image:   "node:18-alpine",
						Command: []string{"node"},
						Args:    []string{"server.js"},
						Ports: []*pb.PortMapping{
							{
								ContainerPort: 3000,
								HostPort:      3000,
								Protocol:      "tcp",
							},
						},
						EnvironmentVariables: map[string]string{
							"NODE_ENV":    "production",
							"PORT":        "3000",
							"DB_HOST":     "database",
							"DB_PORT":     "5432",
							"DB_NAME":     "webapp",
							"DB_USER":     "webapp_user",
							"JWT_SECRET":  "your-secret-key",
							"API_VERSION": "v1",
						},
						Volumes: []*pb.VolumeMount{
							{
								Name:      "app-code",
								MountPath: "/usr/src/app",
								HostPath:  "/opt/webapp/backend",
								ReadOnly:  true,
							},
						},
						Resources: &pb.ResourceLimits{
							MemoryMb: 1024,
							CpuCores: 1.0,
							DiskMb:   2048,
						},
						HealthCheck: &pb.HealthCheck{
							Command:            []string{"curl", "-f", "http://localhost:3000/health"},
							IntervalSeconds:    30,
							TimeoutSeconds:     5,
							Retries:            3,
							StartPeriodSeconds: 30,
						},
						RestartPolicy: pb.RestartPolicy_RESTART_POLICY_UNLESS_STOPPED,
					},
					DatabaseConnectionString: "postgresql://webapp_user:password@database:5432/webapp",
					ApiKeys: map[string]string{
						"EXTERNAL_API_KEY": "your-external-api-key",
						"PAYMENT_API_KEY":  "your-payment-api-key",
					},
				},
				Database: &pb.DatabaseConfig{
					Container: &pb.ContainerConfig{
						Name:  "webapp-database",
						Image: "postgres:15-alpine",
						Ports: []*pb.PortMapping{
							{
								ContainerPort: 5432,
								HostPort:      5432,
								Protocol:      "tcp",
							},
						},
						EnvironmentVariables: map[string]string{
							"POSTGRES_DB":       "webapp",
							"POSTGRES_USER":     "webapp_user",
							"POSTGRES_PASSWORD": "secure_password_123",
							"PGDATA":            "/var/lib/postgresql/data/pgdata",
						},
						Volumes: []*pb.VolumeMount{
							{
								Name:      "postgres-data",
								MountPath: "/var/lib/postgresql/data",
								HostPath:  "/opt/webapp/postgres-data",
								ReadOnly:  false,
							},
						},
						Resources: &pb.ResourceLimits{
							MemoryMb: 2048,
							CpuCores: 1.0,
							DiskMb:   10240,
						},
						HealthCheck: &pb.HealthCheck{
							Command:            []string{"pg_isready", "-U", "webapp_user", "-d", "webapp"},
							IntervalSeconds:    30,
							TimeoutSeconds:     5,
							Retries:            5,
							StartPeriodSeconds: 60,
						},
						RestartPolicy: pb.RestartPolicy_RESTART_POLICY_UNLESS_STOPPED,
					},
					DatabaseName:      "webapp",
					Username:          "webapp_user",
					Password:          "secure_password_123",
					PersistentStorage: true,
					StoragePath:       "/opt/webapp/postgres-data",
				},
				AdditionalServices: map[string]*pb.ContainerConfig{
					"redis-cache": {
						Name:  "webapp-redis",
						Image: "redis:7-alpine",
						Ports: []*pb.PortMapping{
							{
								ContainerPort: 6379,
								HostPort:      6379,
								Protocol:      "tcp",
							},
						},
						Resources: &pb.ResourceLimits{
							MemoryMb: 256,
							CpuCores: 0.25,
							DiskMb:   512,
						},
						RestartPolicy: pb.RestartPolicy_RESTART_POLICY_UNLESS_STOPPED,
					},
				},
			},
			Labels: map[string]string{
				"environment": "development",
				"team":        "backend",
				"project":     "webapp",
				"version":     "1.0.0",
			},
			Network: &pb.NetworkConfig{
				NetworkName: "webapp-network",
				Subnet:      "172.20.0.0/16",
				Gateway:     "172.20.0.1",
				Isolated:    false,
			},
		},
	}

	resp, err := client.CreateEnvironment(ctx, req)
	if err != nil {
		fmt.Printf("❌ CreateEnvironment failed: %v\n", err)
	} else {
		fmt.Printf("✅ CreateEnvironment response: %v\n", resp)
	}
}

func testListEnvironments(ctx context.Context, client pb.SchedulerServiceClient) {
	fmt.Println("\n📋 Testing ListEnvironments...")

	req := &pb.ListEnvironmentsRequest{
		PageSize:  10,
		PageToken: "",
		Filters: map[string]string{
			"environment": "development",
			"team":        "backend",
		},
	}

	resp, err := client.ListEnvironments(ctx, req)
	if err != nil {
		fmt.Printf("❌ ListEnvironments failed: %v\n", err)
	} else {
		fmt.Printf("✅ ListEnvironments response: %v\n", resp)
	}
}

func testGetEnvironment(ctx context.Context, client pb.SchedulerServiceClient) {
	fmt.Println("\n🔍 Testing GetEnvironment...")

	req := &pb.GetEnvironmentRequest{
		Id: "env-12345-webapp-stack",
	}

	resp, err := client.GetEnvironment(ctx, req)
	if err != nil {
		fmt.Printf("❌ GetEnvironment failed: %v\n", err)
	} else {
		fmt.Printf("✅ GetEnvironment response: %v\n", resp)
	}
}

func testStartEnvironment(ctx context.Context, client pb.SchedulerServiceClient) {
	fmt.Println("\n▶️ Testing StartEnvironment...")

	req := &pb.StartEnvironmentRequest{
		Id: "env-12345-webapp-stack",
	}

	resp, err := client.StartEnvironment(ctx, req)
	if err != nil {
		fmt.Printf("❌ StartEnvironment failed: %v\n", err)
	} else {
		fmt.Printf("✅ StartEnvironment response: %v\n", resp)
	}
}

func testStopEnvironment(ctx context.Context, client pb.SchedulerServiceClient) {
	fmt.Println("\n⏹️ Testing StopEnvironment...")

	req := &pb.StopEnvironmentRequest{
		Id:    "env-12345-webapp-stack",
		Force: false,
	}

	resp, err := client.StopEnvironment(ctx, req)
	if err != nil {
		fmt.Printf("❌ StopEnvironment failed: %v\n", err)
	} else {
		fmt.Printf("✅ StopEnvironment response: %v\n", resp)
	}
}

func testRestartEnvironment(ctx context.Context, client pb.SchedulerServiceClient) {
	fmt.Println("\n🔄 Testing RestartEnvironment...")

	req := &pb.RestartEnvironmentRequest{
		Id: "env-12345-webapp-stack",
	}

	resp, err := client.RestartEnvironment(ctx, req)
	if err != nil {
		fmt.Printf("❌ RestartEnvironment failed: %v\n", err)
	} else {
		fmt.Printf("✅ RestartEnvironment response: %v\n", resp)
	}
}

func testGetEnvironmentStatus(ctx context.Context, client pb.SchedulerServiceClient) {
	fmt.Println("\n📊 Testing GetEnvironmentStatus...")

	req := &pb.GetEnvironmentStatusRequest{
		Id: "env-12345-webapp-stack",
	}

	resp, err := client.GetEnvironmentStatus(ctx, req)
	if err != nil {
		fmt.Printf("❌ GetEnvironmentStatus failed: %v\n", err)
	} else {
		fmt.Printf("✅ GetEnvironmentStatus response: %v\n", resp)
	}
}

func testUpdateEnvironment(ctx context.Context, client pb.SchedulerServiceClient) {
	fmt.Println("\n✏️ Testing UpdateEnvironment...")

	// Create an updated specification (e.g., scaling up resources)
	req := &pb.UpdateEnvironmentRequest{
		Id: "env-12345-webapp-stack",
		Spec: &pb.EnvironmentSpecification{
			Name:        "test-webapp-stack",
			Description: "Updated web application stack with increased resources",
			ApplicationStack: &pb.ApplicationStack{
				Name:    "webapp-stack",
				Version: "1.1.0",
				Backend: &pb.BackendConfig{
					Container: &pb.ContainerConfig{
						Name:  "webapp-backend",
						Image: "node:18-alpine",
						Resources: &pb.ResourceLimits{
							MemoryMb: 2048, // Increased from 1024
							CpuCores: 2.0,  // Increased from 1.0
							DiskMb:   4096, // Increased from 2048
						},
					},
				},
			},
			Labels: map[string]string{
				"environment": "development",
				"team":        "backend",
				"project":     "webapp",
				"version":     "1.1.0", // Updated version
			},
		},
	}

	resp, err := client.UpdateEnvironment(ctx, req)
	if err != nil {
		fmt.Printf("❌ UpdateEnvironment failed: %v\n", err)
	} else {
		fmt.Printf("✅ UpdateEnvironment response: %v\n", resp)
	}
}

func testGetEnvironmentLogs(ctx context.Context, client pb.SchedulerServiceClient) {
	fmt.Println("\n📝 Testing GetEnvironmentLogs...")

	req := &pb.GetEnvironmentLogsRequest{
		Id:            "env-12345-webapp-stack",
		ContainerName: "webapp-backend", // Get logs from specific container
		Since:         timestamppb.New(time.Now().Add(-1 * time.Hour)),
		TailLines:     100,
		Follow:        false, // Don't follow for this test
	}

	stream, err := client.GetEnvironmentLogs(ctx, req)
	if err != nil {
		fmt.Printf("❌ GetEnvironmentLogs failed: %v\n", err)
		return
	}

	fmt.Printf("✅ GetEnvironmentLogs stream opened successfully\n")

	// Try to receive a few log messages (will likely get error since service is not implemented)
	for i := 0; i < 3; i++ {
		resp, err := stream.Recv()
		if err != nil {
			fmt.Printf("❌ GetEnvironmentLogs stream error: %v\n", err)
			break
		}
		fmt.Printf("📄 Log: [%s] %s: %s\n", resp.Timestamp.AsTime().Format("15:04:05"), resp.ContainerName, resp.Message)
	}
}

func testDeleteEnvironment(ctx context.Context, client pb.SchedulerServiceClient) {
	fmt.Println("\n🗑️ Testing DeleteEnvironment...")

	req := &pb.DeleteEnvironmentRequest{
		Id: "env-12345-webapp-stack",
	}

	resp, err := client.DeleteEnvironment(ctx, req)
	if err != nil {
		fmt.Printf("❌ DeleteEnvironment failed: %v\n", err)
	} else {
		fmt.Printf("✅ DeleteEnvironment response: %v\n", resp)
	}
}
